#!/usr/bin/env python3

"""
Scan Cursor workspaceStorage/*/state.vscdb for presence of interesting keys.

Goal: find which workspaces store chat/composer data in a way that includes assistant messages.
"""

from __future__ import annotations

import argparse
import sqlite3
from pathlib import Path
from typing import List, Tuple


def has_key_like(db_path: Path, like: str) -> int:
    conn = sqlite3.connect(str(db_path))
    cur = conn.cursor()
    cur.execute("SELECT COUNT(*) FROM ItemTable WHERE key LIKE ?", (like,))
    (n,) = cur.fetchone()
    conn.close()
    return int(n)


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--workspace-storage-root", required=True, help="Path to ~/.config/Cursor/User/workspaceStorage")
    p.add_argument("--like", action="append", required=True, help="SQL LIKE filter (repeatable), e.g. '%aiService%'")
    p.add_argument("--limit", type=int, default=0, help="Limit number of workspace DBs to scan (0=all)")
    args = p.parse_args()

    root = Path(args.workspace_storage_root)
    dbs = sorted(root.glob("*/state.vscdb"))
    if args.limit and args.limit > 0:
        dbs = dbs[: args.limit]

    print("workspaceDb\t" + "\t".join(args.like))
    for db in dbs:
        counts: List[str] = []
        ok = True
        for like in args.like:
            try:
                n = has_key_like(db, like)
                counts.append(str(n))
            except Exception:
                ok = False
                counts.append("ERR")
        if ok and all(c == "0" for c in counts):
            continue
        print(str(db) + "\t" + "\t".join(counts))

    return 0


if __name__ == "__main__":
    raise SystemExit(main())


