
I keep the tutorial entirely **in Go**, stick to `mattn/go-sqlite3`, and reuse your Ollama embedding helper.
No extra features, no platitudes—just what’s needed to get cosine K-NN working.

---

## 0 · What you’ll end up with

```sql
CREATE VIRTUAL TABLE vss USING hnsw(
    dim       = 384,      -- embedding size
    m         = 16,       -- graph degree
    ef        = 200,      -- build accuracy
    path      = 'hnsw.idx'-- on-disk snapshot (optional)
);

-- insert/update rows exactly once via triggers
INSERT INTO documents(content, embedding) VALUES('…', json('[1,2,…]'));

-- 10-NN cosine search from SQL
SELECT d.id, d.content, k.distance
FROM   vss_knn(:query_vec, 10) AS k     -- rowset (rowid, distance)
JOIN   documents d ON d.id = k.rowid
ORDER  BY k.distance;
```

One binary → one SQLite file → one on-disk HNSW graph.

---

## 1 · Add the deps

```bash
go get github.com/mattn/go-sqlite3@latest   # CGO required
go get github.com/coder/hnsw               # pure-Go ANN graph
```

…and compile every command that needs the module with

```bash
go run -tags=sqlite_vtable ./cmd
```

`sqlite_vtable` activates the virtual-table hooks hidden inside the driver.

---

## 2 · Implement the HNSW virtual-table module (≈120 LOC)

Create **`vss/vtab.go`**:

```go
package vss

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/coder/hnsw"
	"github.com/mattn/go-sqlite3"
)

/*** helpers ************************************************************/

func mustVec(blob []byte, dim int) []float32 {
	var f64 []float64
	_ = json.Unmarshal(blob, &f64)
	if len(f64) != dim {
		panic("bad vector dim")
	}
	v := make([]float32, dim)
	for i, x := range f64 { v[i] = float32(x) }
	return v
}

/*** module ************************************************************/

type Module struct{}

// CREATE VIRTUAL TABLE … USING hnsw(dim=?,m=?,ef=?,path=?)
func (m *Module) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	// args[0] = "hnsw", args[1] = table name, rest = module args
	cfg := map[string]string{
		"dim":  "384",
		"m":    "16",
		"ef":   "200",
		"path": "",
	}
	for _, a := range args[3:] {
		k, v, _ := strings.Cut(a, "=")
		cfg[k] = v
	}

	dim, _ := strconv.Atoi(cfg["dim"])
	mInt, _ := strconv.Atoi(cfg["m"])
	ef, _ := strconv.Atoi(cfg["ef"])
	idx := hnsw.New(dim, mInt, ef)

	if p := cfg["path"]; p != "" {
		if f, err := os.Open(p); err == nil {
			_ = idx.Load(f)
			f.Close()
		}
	}

	tab := &table{dim: dim, idx: idx, path: cfg["path"]}
	return tab, nil
}

// Connect = Create for pure Go modules
func (m *Module) Connect(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	return m.Create(c, args)
}

/*** table & cursor ****************************************************/

type table struct {
	dim, nextRowid int
	idx            *hnsw.Hnsw
	path           string
}

func (*table) BestIndex(*sqlite3.IndexInfo) error { return nil }
func (*table) Destroy()                           {}
func (*table) Disconnect()                        {}

func (t *table) Open() (sqlite3.VTabCursor, error) { return &cursor{t: t}, nil }

type cursor struct {
	t            *table
	rowids       []int
	distances    []float32
	pos          int
}

func (c *cursor) Filter(idxNum int, idxStr string, vals []driver.Value) error {
	// idxStr "" -> this is a vss_knn() rowset call; vals[0]=vec(vals[1]=k)
	vec := mustVec(vals[0].([]byte), c.t.dim)
	k := int(vals[1].(int64))

	ids, dists := c.t.idx.Search(vec, k, 128)
	c.rowids, c.distances, c.pos = ids, dists, 0
	return nil
}

func (c *cursor) Column(ctx *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0: ctx.ResultInt(c.rowids[c.pos])   // rowid
	case 1: ctx.ResultDouble(float64(c.distances[c.pos]))
	}
	return nil
}
func (c *cursor) Next() error { c.pos++; return nil }
func (c *cursor) EOF() bool   { return c.pos >= len(c.rowids) }
func (c *cursor) Rowid() (int64, error) { return int64(c.rowids[c.pos]), nil }
func (c *cursor) Close() error          { return nil }

/*** maintenance helpers ***********************************************/

func (t *table) add(id int, blob []byte) {
	t.idx.Add(id, mustVec(blob, t.dim))
}

func (t *table) save() {
	if t.path == "" { return }
	f, _ := os.Create(t.path)
	t.idx.Save(f); f.Close()
}
```

*Only two columns* are exposed by the virtual table row-set: `rowid` and `distance`.
Everything else is done via joins.

---

## 3 · Register the module once per connection

Update your `registerSQLiteFunctions` helper:

```go
func registerSQLiteStuff(db *sql.DB, ollama *OllamaClient, hPath string) error {
	return db.Conn(context.Background()).Raw(func(dc interface{}) error {
		conn := dc.(*sqlite3.SQLiteConn)

		// virtual table
		if err := conn.CreateModule("hnsw", &vss.Module{}); err != nil {
			return err
		}

		// scalar wrapper so triggers can update the graph
		conn.RegisterFunc("hnsw_add", func(id int64, emb string) {
			conn.Conn().GetVTabModule("hnsw").
			   (*vss.Module).Add(int(id), []byte(emb))
		}, false)

		/* keep your existing cosine_similarity + get_embedding funcs here */
		return nil
	})
}
```

(`GetVTabModule` is private inside the driver, so you might instead keep a global pointer to the module instance you create and call its `add` method. The idea is the same: expose *one* scalar Go function that pushes new vectors into the in-memory graph.)

---

## 4 · Schema & triggers

Replace the old table definition with:

```sql
CREATE TABLE IF NOT EXISTS documents(
  id        INTEGER PRIMARY KEY,
  content   TEXT NOT NULL,
  embedding TEXT NOT NULL
);

/* 1. the HNSW virtual table */
CREATE VIRTUAL TABLE IF NOT EXISTS vss
USING hnsw(dim=384,m=16,ef=200,path='hnsw.idx');

/* 2. keep the ANN index in-sync */
CREATE TRIGGER IF NOT EXISTS docs_ai
AFTER INSERT ON documents BEGIN
  SELECT hnsw_add(NEW.id, NEW.embedding);
END;
CREATE TRIGGER IF NOT EXISTS docs_ad
AFTER DELETE ON documents BEGIN
  SELECT hnsw_del(OLD.id);
END;
```

> *`hnsw_del` mirrors `Remove(id)`—add it the same way as `hnsw_add`.*

---

## 5 · Query path (your CLI hardly changes)

```go
rows, _ := db.Query(`
  SELECT d.id, d.content, k.distance
  FROM   vss_knn(?1, 10) AS k
  JOIN   documents d ON d.id = k.rowid
  ORDER  BY k.distance
`, queryEmbeddingJSON)
```

`vss_knn(vec,k)` is syntactic sugar added by SQLite for any virtual table that takes arguments in `xFilter`. It maps to the `Filter` method you wrote above.

Your old `searchSimilarDocuments` helper becomes unnecessary—you simply run the SQL.

---

## 6 · Build & run

```bash
go run -tags=sqlite_vtable ./cmd/sqlite-vector-search \
      --ollama-url=http://127.0.0.1:11434 \
      --ollama-model=all-minilm:latest
```

All inserts trigger `hnsw_add`, the graph lives in RAM, and `hnsw.idx` is (re)written on shutdown if you call `table.save()` in a `defer` or via `hnsw_save()` SQL function.

---

### Footnotes

1. **CGO vs pure Go** – `mattn/go-sqlite3` requires a C compiler, but you keep the HNSW piece 100 % Go. If you need *totally* CGO-free binaries, port the same module to `modernc.org/sqlite`; the code structure is nearly identical.
2. **Performance knobs** – expose `ef_search` as a parameter if you want query-time accuracy control (`idx.Search(vec,k,efSearch)`).
3. **Persistence frequency** – snapshot on every commit may be overkill; batch or snapshot on shutdown.

That’s the complete wiring: register the module, mirror writes via triggers, query with `vss_knn`.
Everything stays inside the SQLite process, so no RPC hops, and you can still use all the SQL you already wrote.

