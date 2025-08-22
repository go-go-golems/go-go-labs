package main

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	_ "modernc.org/sqlite"
)

func getDBPath() string {
    if p := os.Getenv("TOOLEXEC_DB"); p != "" {
        return p
    }
    return "./build_times.db"
}

func openDB(ctx context.Context) (*sql.DB, error) {
    dsn := getDBPath() + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout=5000"
    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, errors.WithStack(err)
    }
    // Quick ping to validate
    if err := db.PingContext(ctx); err != nil {
        _ = db.Close()
        return nil, errors.WithStack(err)
    }
    if err := ensureSchema(db); err != nil {
        _ = db.Close()
        return nil, errors.WithStack(err)
    }
    return db, nil
}

func ensureSchema(db *sql.DB) error {
    _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS runs (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  ts_unix   INTEGER NOT NULL,
  comment   TEXT    NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_runs_ts ON runs(ts_unix);

CREATE TABLE IF NOT EXISTS invocations (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id       INTEGER,
  ts_unix      INTEGER NOT NULL,
  tool         TEXT    NOT NULL,
  tool_path    TEXT    NOT NULL DEFAULT '',
  pkg          TEXT    NOT NULL,
  status       INTEGER NOT NULL,
  elapsed_ms   INTEGER NOT NULL,
  args         TEXT    NOT NULL,
  os           TEXT    NOT NULL DEFAULT '',
  arch         TEXT    NOT NULL DEFAULT '',
  cwd          TEXT    NOT NULL DEFAULT '',
  out          TEXT    NOT NULL DEFAULT '',
  importcfg    TEXT    NOT NULL DEFAULT '',
  embedcfg     TEXT    NOT NULL DEFAULT '',
  buildid      TEXT    NOT NULL DEFAULT '',
  goversion    TEXT    NOT NULL DEFAULT '',
  lang         TEXT    NOT NULL DEFAULT '',
  concurrency  INTEGER NOT NULL DEFAULT 0,
  complete     INTEGER NOT NULL DEFAULT 0,
  pack         INTEGER NOT NULL DEFAULT 0,
  source_count INTEGER NOT NULL DEFAULT 0,
  flags_json   TEXT    NOT NULL DEFAULT '',
  FOREIGN KEY (run_id) REFERENCES runs(id)
);
CREATE INDEX IF NOT EXISTS idx_inv_tool_ts ON invocations(tool, ts_unix);
CREATE INDEX IF NOT EXISTS idx_inv_pkg_ts  ON invocations(pkg, ts_unix);
CREATE INDEX IF NOT EXISTS idx_inv_run_ts  ON invocations(run_id, ts_unix);
`)
    if err != nil {
        return errors.WithStack(err)
    }

    // Additive migrations for existing DBs (best-effort)
    // Each ALTER is idempotent-ish: ignore errors if the column already exists
    alters := []string{
        `ALTER TABLE invocations ADD COLUMN tool_path TEXT NOT NULL DEFAULT ''`,
        `ALTER TABLE invocations ADD COLUMN os TEXT NOT NULL DEFAULT ''`,
        `ALTER TABLE invocations ADD COLUMN arch TEXT NOT NULL DEFAULT ''`,
        `ALTER TABLE invocations ADD COLUMN cwd TEXT NOT NULL DEFAULT ''`,
        `ALTER TABLE invocations ADD COLUMN out TEXT NOT NULL DEFAULT ''`,
        `ALTER TABLE invocations ADD COLUMN importcfg TEXT NOT NULL DEFAULT ''`,
        `ALTER TABLE invocations ADD COLUMN embedcfg TEXT NOT NULL DEFAULT ''`,
        `ALTER TABLE invocations ADD COLUMN buildid TEXT NOT NULL DEFAULT ''`,
        `ALTER TABLE invocations ADD COLUMN goversion TEXT NOT NULL DEFAULT ''`,
        `ALTER TABLE invocations ADD COLUMN lang TEXT NOT NULL DEFAULT ''`,
        `ALTER TABLE invocations ADD COLUMN concurrency INTEGER NOT NULL DEFAULT 0`,
        `ALTER TABLE invocations ADD COLUMN complete INTEGER NOT NULL DEFAULT 0`,
        `ALTER TABLE invocations ADD COLUMN pack INTEGER NOT NULL DEFAULT 0`,
        `ALTER TABLE invocations ADD COLUMN source_count INTEGER NOT NULL DEFAULT 0`,
        `ALTER TABLE invocations ADD COLUMN flags_json TEXT NOT NULL DEFAULT ''`,
    }
    for _, q := range alters {
        if _, e := db.Exec(q); e != nil {
            // ignore; likely exists
        }
    }
    return nil
}

func insertRun(ctx context.Context, db *sql.DB, comment string) (int64, error) {
    res, err := db.ExecContext(ctx, `INSERT INTO runs (ts_unix, comment) VALUES (?, ?)`, time.Now().Unix(), comment)
    if err != nil {
        return 0, errors.WithStack(err)
    }
    id, err := res.LastInsertId()
    if err != nil {
        return 0, errors.WithStack(err)
    }
    return id, nil
}

func currentRunID() (int64, bool) {
    // Prefer TOOLEXEC_RUN_ID; fall back to GO_BUILD_ANALYZER_RUN_ID
    if s := os.Getenv("TOOLEXEC_RUN_ID"); s != "" {
        if v, err := strconv.ParseInt(s, 10, 64); err == nil && v > 0 {
            return v, true
        }
    }
    if s := os.Getenv("GO_BUILD_ANALYZER_RUN_ID"); s != "" {
        if v, err := strconv.ParseInt(s, 10, 64); err == nil && v > 0 {
            return v, true
        }
    }
    return 0, false
}


