#!/usr/bin/env python3

"""
Dump a single ItemTable key from a state.vscdb (workspace or global).

Includes conservative redaction for JSON objects that look like secrets.
"""

from __future__ import annotations

import argparse
import json
import sqlite3
from typing import Any


SENSITIVE_SUBSTRINGS = ("secret", "token", "apikey", "api_key", "password", "key")


def _redact(obj: Any) -> Any:
    if isinstance(obj, dict):
        out = {}
        for k, v in obj.items():
            lk = k.lower()
            if any(s in lk for s in SENSITIVE_SUBSTRINGS):
                out[k] = "<redacted>"
            else:
                out[k] = _redact(v)
        return out
    if isinstance(obj, list):
        return [_redact(x) for x in obj]
    return obj


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--db", required=True, help="Path to state.vscdb")
    p.add_argument("--key", required=True, help="ItemTable.key to dump")
    p.add_argument("--max-chars", type=int, default=20000, help="Max chars to print for raw (non-JSON) values")
    args = p.parse_args()

    conn = sqlite3.connect(args.db)
    cur = conn.cursor()
    cur.execute("SELECT value, typeof(value), LENGTH(value) FROM ItemTable WHERE key = ?", (args.key,))
    row = cur.fetchone()
    if not row:
        raise SystemExit(f"Key not found: {args.key}")

    val, t, size = row
    if isinstance(val, bytes):
        text = val.decode("utf-8", "replace")
    else:
        text = str(val)

    print(f"key: {args.key}")
    print(f"type: {t}")
    print(f"size: {size}")

    # Try JSON parse
    try:
        data = json.loads(text)
        data = _redact(data)
        print(json.dumps(data, indent=2, ensure_ascii=False)[: args.max_chars])
        return 0
    except Exception:
        pass

    # Fallback to raw
    print(text[: args.max_chars])
    return 0


if __name__ == "__main__":
    raise SystemExit(main())


