#!/usr/bin/env python3

"""
Extract and filter Cursor globalStorage `aiCodeTrackingLines`.

We found `aiCodeTrackingLines` is a large JSON array (~3.8MB) containing objects like:
  {
    "hash": "...",
    "metadata": {
      "source": "composer",
      "composerId": "...",
      "fileExtension": "md",
      "fileName": "...",
      "invocationID": "...",
      "timestamp": 1765808745868
    }
  }

This script filters by composerId and prints matching metadata.
"""

from __future__ import annotations

import argparse
import json
import sqlite3
from typing import Any, Dict, List


def _get_text(cur: sqlite3.Cursor, key: str) -> str:
    cur.execute("SELECT value FROM ItemTable WHERE key = ?", (key,))
    row = cur.fetchone()
    if not row:
        raise RuntimeError(f"Key not found: {key}")
    val = row[0]
    if isinstance(val, bytes):
        return val.decode("utf-8", "replace")
    return str(val)


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--global-db", required=True, help="Path to globalStorage/state.vscdb")
    p.add_argument("--composer-id", required=True, help="composerId / conversation_id UUID")
    p.add_argument("--limit", type=int, default=50, help="Max matches to print (0 = unlimited)")
    p.add_argument("--out-json", help="Optional path to write full filtered JSON list")
    args = p.parse_args()

    conn = sqlite3.connect(args.global_db)
    cur = conn.cursor()

    raw = _get_text(cur, "aiCodeTrackingLines")
    arr: List[Dict[str, Any]] = json.loads(raw)

    sel = [x for x in arr if x.get("metadata", {}).get("composerId") == args.composer_id]

    if args.out_json:
        with open(args.out_json, "w", encoding="utf-8") as f:
            json.dump(sel, f, indent=2, ensure_ascii=False)

    lim = sel if args.limit == 0 else sel[: args.limit]
    print(json.dumps({"composerId": args.composer_id, "matched": len(sel), "sample": lim}, indent=2, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())


