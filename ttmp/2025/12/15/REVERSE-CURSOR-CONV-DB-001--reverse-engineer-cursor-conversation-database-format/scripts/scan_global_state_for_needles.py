#!/usr/bin/env python3

"""
Byte-level needle scan for VS Code/Cursor global storage DB (state.vscdb).

Why: SQLite `LIKE '%...%'` won't match BLOB values. This script scans values as bytes
and reports which keys contain the needle.
"""

from __future__ import annotations

import argparse
import sqlite3
from typing import List, Tuple


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--db", required=True, help="Path to global state.vscdb (SQLite)")
    p.add_argument("--needle", action="append", required=True, help="Needle string (repeatable)")
    p.add_argument("--max-hits", type=int, default=50, help="Max hits per needle")
    args = p.parse_args()

    needles = [(n, n.encode("utf-8")) for n in args.needle]
    hits = {n: [] for n, _ in needles}  # type: ignore[var-annotated]

    conn = sqlite3.connect(args.db)
    cur = conn.cursor()
    cur.execute("SELECT key, value FROM ItemTable")

    scanned = 0
    for key, val in cur:
        scanned += 1
        if val is None:
            continue
        b = val.encode("utf-8", "ignore") if isinstance(val, str) else val
        for n_str, n_bytes in needles:
            if len(hits[n_str]) >= args.max_hits:
                continue
            if n_bytes in b:
                hits[n_str].append((key, len(b)))

    print(f"scanned_keys\t{scanned}")
    for n_str, _ in needles:
        print(f"\n== needle: {n_str} ==")
        if not hits[n_str]:
            print("NO HITS")
            continue
        for key, size in hits[n_str]:
            print(f"{key}\t{size}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())


