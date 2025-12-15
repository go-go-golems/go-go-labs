#!/usr/bin/env python3

"""
List keys in Cursor/VS Code global storage state.vscdb with sizes/types.

This is useful for quickly spotting large blobs that might contain chat transcripts,
response caches, or tracking metadata.
"""

from __future__ import annotations

import argparse
import sqlite3


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--db", required=True, help="Path to globalStorage/state.vscdb")
    p.add_argument("--top", type=int, default=50, help="Show top N keys by size")
    p.add_argument("--like", default=None, help="Optional SQL LIKE filter, e.g. '%cursorai%'")
    args = p.parse_args()

    conn = sqlite3.connect(args.db)
    cur = conn.cursor()

    if args.like:
        cur.execute(
            "SELECT key, LENGTH(value) AS size, typeof(value) AS t "
            "FROM ItemTable WHERE key LIKE ? ORDER BY size DESC, key ASC LIMIT ?",
            (args.like, args.top),
        )
    else:
        cur.execute(
            "SELECT key, LENGTH(value) AS size, typeof(value) AS t "
            "FROM ItemTable ORDER BY size DESC, key ASC LIMIT ?",
            (args.top,),
        )

    for key, size, t in cur.fetchall():
        print(f"{key}\t{size}\t{t}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())


