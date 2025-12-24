#!/usr/bin/env python3

"""
Map sqlite page numbers to owning objects using dbstat virtual table.

Requires SQLite compiled with SQLITE_ENABLE_DBSTAT_VTAB (commonly true).
If dbstat is unavailable, this script exits with a clear error.
"""

from __future__ import annotations

import argparse
import sqlite3
from typing import List, Tuple


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--db", required=True, help="Path to sqlite DB")
    p.add_argument("--page", action="append", required=True, help="Page number (1-based). Repeatable.")
    args = p.parse_args()

    pages = [int(x) for x in args.page]
    placeholders = ",".join(["?"] * len(pages))

    conn = sqlite3.connect(args.db)
    cur = conn.cursor()

    # quick probe
    try:
        cur.execute("SELECT name FROM sqlite_master WHERE name='dbstat' AND type='table'")
        # dbstat is virtual, not in sqlite_master; probe directly:
        cur.execute("SELECT name, path, pageno, pagetype, ncell, payload, unused, mx_payload FROM dbstat LIMIT 1")
        _ = cur.fetchone()
    except Exception as e:
        raise SystemExit(f"dbstat not available in this sqlite build: {e}")

    cur.execute(
        f"SELECT name, path, pageno, pagetype, ncell, payload, unused, mx_payload "
        f"FROM dbstat WHERE pageno IN ({placeholders}) ORDER BY pageno ASC",
        pages,
    )
    rows: List[Tuple] = cur.fetchall()

    print("name\tpath\tpageno\tpagetype\tncell\tpayload\tunused\tmx_payload")
    for r in rows:
        print("\t".join(str(x) for x in r))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())


