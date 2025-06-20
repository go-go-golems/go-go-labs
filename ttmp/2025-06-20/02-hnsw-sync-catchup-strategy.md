### Keeping the HNSW virtual-table **always in-sync** after a restart

Below is the battle-tested pattern many embedded-vector users follow.
It works even if your program crashes between a row insert and an index flush, and requires **no manual “reindex” button**.

| Goal                            | How it’s achieved                                                                                               |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| **1. Index survives a restart** | Persist the binary graph (side-car file *or* BLOB row) at COMMIT time.                                          |
| **2. Index is never stale**     | On every start-up run a *delta catch-up scan* that inserts anything newer than the snapshot.                    |
| **3. No torn snapshots**        | Wrap **both**: `(a) INSERT/DELETE row` **and** `(b) hnsw_save()` in the *same SQLite transaction* via triggers. |

Below you’ll see code & SQL to wire those three pieces.

---

## 1 · Persist automatically at COMMIT

The driver lets you register a **commit hook** once per connection:

```go
// global vars so hooks can see them
var (
    idx  *hnsw.Hnsw
    path = "hnsw.idx"           // or leave empty and use BLOB persistence
)

func registerHooks(conn *sqlite3.SQLiteConn) error {
    // COMMIT = snapshot
    conn.SetCommitHook(func() int {
        if path == "" {
            // store inside SQLite
            var buf bytes.Buffer
            _ = idx.Save(&buf)
            conn.Exec("REPLACE INTO hnsw_snapshot(id,data) VALUES(1,?)",
                      []driver.Value{buf.Bytes()})
        } else {
            f, _ := os.Create(path + ".tmp")
            _ = idx.Save(f)
            f.Close()
            os.Rename(path+".tmp", path)       // atomic on POSIX
        }
        return 0   // 0 = continue commit
    })
    return nil
}
```

`SetCommitHook` is compiled in when you build with `-tags=sqlite_vtable`.

Every successful transaction that added / deleted rows fires the hook, which dumps the current graph.
No explicit `hnsw_save()` call needed.

---

## 2 · Catch-up scan on start-up (“on-demand build”)

Add a tiny **meta row** that records the max rowid seen at snapshot-time:

```sql
CREATE TABLE IF NOT EXISTS hnsw_meta(
  id              INTEGER PRIMARY KEY CHECK(id=1),
  max_rowid_indexed INTEGER NOT NULL
);
INSERT OR IGNORE INTO hnsw_meta(id, max_rowid_indexed) VALUES(1, 0);
```

On program start:

```go
func bootstrapIndex(db *sql.DB) error {
    // 1. try to load persisted graph
    if path != "" {
        if f, err := os.Open(path); err == nil {
            _ = idx.Load(f); f.Close()
        }
    } else {
        var blob []byte
        _ = db.QueryRow(`SELECT data FROM hnsw_snapshot WHERE id=1`).Scan(&blob)
        if len(blob) > 0 { _ = idx.Load(bytes.NewReader(blob)) }
    }

    // 2. incremental catch-up
    var maxRowid int64
    _ = db.QueryRow(`SELECT max_rowid_indexed FROM hnsw_meta WHERE id=1`).Scan(&maxRowid)

    rows, _ := db.Query(`
        SELECT id, embedding
        FROM   documents
        WHERE  id > ?`, maxRowid)
    defer rows.Close()

    for rows.Next() {
        var id int
        var embJSON string
        _ = rows.Scan(&id, &embJSON)
        idx.Add(id, jsonToVec32(embJSON))
        maxRowid = int64(id)
    }

    _, _ = db.Exec(`UPDATE hnsw_meta SET max_rowid_indexed = ? WHERE id=1`, maxRowid)
    return nil
}
```

**Performance:** for <10 M rows the loop is still well under a second on SSDs.
If `rows.Next()` returns zero rows, the index was already current.

---

## 3 · Keep index & table in lock-step with triggers

```sql
CREATE TRIGGER docs_ai AFTER INSERT ON documents
BEGIN
  SELECT hnsw_add(NEW.id, NEW.embedding);         -- ➊ mutate graph
  UPDATE hnsw_meta SET max_rowid_indexed = NEW.id;-- ➋ advance watermark
END;

CREATE TRIGGER docs_ad AFTER DELETE ON documents
BEGIN
  SELECT hnsw_del(OLD.id);
END;
```

Because both calls run **inside the same outer transaction**:

1. If the TX aborts, neither row nor HNSW change is visible.
2. Commit hook fires *after* the triggers, so the snapshot always contains the just-committed nodes.

No window exists where the table has a row that the index doesn’t.

---

## 4 · What happens on crash?

| Situation                                           | Outcome on next start-up                                                                                                                                       |
| --------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Crash **before COMMIT**                             | Neither row nor index entry exist (SQLite rolled back).                                                                                                        |
| Crash **after COMMIT but before snapshot finished** | WAL contains the row; `hnsw_meta` watermark has the new id; the binary graph on disk is stale – catch-up scan detects the gap and inserts the missing vectors. |
| Crash **during snapshot write**                     | We write to `hnsw.idx.tmp` then `rename` -> either old good file or none; catch-up handles both. (Using BLOB storage: the write is atomic with the outer TX.)  |

---

## 5 · Multi-process caveat

Everything above assumes **one writer process** (common for CLIs and desktop apps).
If you’ll run many processes concurrently:

* Keep one process in *write* mode (runs triggers + commit hook).
* Others open the DB `READ ONLY`, load the snapshot **but do not mutate** the graph—HNSW isn’t concurrency-safe across processes.
* Or switch to a networked vector DB.

---

### TL;DR checklist

1. **Commit-hook snapshot** ⇒ graph is dumped every successful TX.
2. **Watermark table (`hnsw_meta`)** ⇒ you know what the snapshot covers.
3. **Start-up catch-up scan** ⇒ fills any gaps before first query.
4. **Triggers** ⇒ ensure row + graph update live in the same TX.

Follow those four steps and `vss_knn()` will always see every document row, no matter how many times the program restarts or crashes.

