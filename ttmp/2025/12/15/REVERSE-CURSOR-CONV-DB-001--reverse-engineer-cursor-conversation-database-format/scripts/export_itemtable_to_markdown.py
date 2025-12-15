#!/usr/bin/env python3

"""
Export a state.vscdb ItemTable to Markdown for manual inspection.

Designed for small workspace state.vscdb files (tens/hundreds of keys).
For huge global state.vscdb, use --key-like to narrow output.

Read-only.
"""

from __future__ import annotations

import argparse
import json
import sqlite3
from datetime import datetime
from typing import Any, Optional, Tuple


SENSITIVE_SUBSTRINGS = ("secret", "token", "apikey", "api_key", "password", "encryptionkey", "encryption_key", "key")


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


def _fetch_value(cur: sqlite3.Cursor, key: str) -> Tuple[Optional[str], str, int]:
    cur.execute("SELECT value, typeof(value), LENGTH(value) FROM ItemTable WHERE key = ?", (key,))
    row = cur.fetchone()
    if not row:
        return None, "null", 0
    val, t, size = row
    if val is None:
        return None, t, int(size or 0)
    if isinstance(val, bytes):
        return val.decode("utf-8", "replace"), t, int(size or 0)
    return str(val), t, int(size or 0)


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--db", required=True, help="Path to state.vscdb (workspace or global)")
    p.add_argument("--out", required=True, help="Output markdown file path")
    p.add_argument("--key-like", default=None, help="Optional SQL LIKE filter for keys, e.g. 'aiService.%'")
    p.add_argument("--max-chars", type=int, default=20000, help="Max chars per value section")
    p.add_argument("--include-raw-json", action="store_true", help="If set, also include the raw JSON string (truncated)")
    args = p.parse_args()

    conn = sqlite3.connect(args.db)
    cur = conn.cursor()

    if args.key_like:
        cur.execute("SELECT key, typeof(value), LENGTH(value) FROM ItemTable WHERE key LIKE ? ORDER BY key ASC", (args.key_like,))
    else:
        cur.execute("SELECT key, typeof(value), LENGTH(value) FROM ItemTable ORDER BY key ASC")

    keys = cur.fetchall()

    with open(args.out, "w", encoding="utf-8") as f:
        f.write("---\n")
        f.write(f"Title: SQLite ItemTable Export\n")
        f.write(f"GeneratedAt: {datetime.now().isoformat()}\n")
        f.write(f"SourceDB: {args.db}\n")
        if args.key_like:
            f.write(f"KeyLike: {args.key_like}\n")
        f.write("DocType: reference\n")
        f.write("---\n\n")

        f.write("# SQLite ItemTable Export\n\n")
        f.write(f"- **DB**: `{args.db}`\n")
        f.write(f"- **Keys exported**: **{len(keys)}**\n")
        if args.key_like:
            f.write(f"- **Filter**: `key LIKE {args.key_like!r}`\n")
        f.write("\n")

        f.write("## Key index\n\n")
        f.write("| key | typeof(value) | length |\n")
        f.write("|---|---:|---:|\n")
        for key, t, size in keys:
            f.write(f"| `{key}` | `{t}` | {size} |\n")

        f.write("\n## Key dumps\n\n")
        for key, _t, _size in keys:
            text, t, size = _fetch_value(cur, key)
            f.write(f"### `{key}`\n\n")
            f.write(f"- **type**: `{t}`\n")
            f.write(f"- **size**: `{size}`\n\n")

            if text is None:
                f.write("_NULL_\n\n")
                continue

            # Attempt JSON pretty-print
            pretty: Optional[str] = None
            try:
                parsed = json.loads(text)
                parsed = _redact(parsed)
                pretty = json.dumps(parsed, indent=2, ensure_ascii=False)
            except Exception:
                pretty = None

            if pretty is not None:
                f.write("#### Parsed JSON\n\n")
                f.write("```json\n")
                f.write(pretty[: args.max_chars])
                if len(pretty) > args.max_chars:
                    f.write("\n... (truncated)\n")
                f.write("\n```\n\n")
                if args.include_raw_json:
                    f.write("#### Raw JSON (as stored)\n\n")
                    f.write("```text\n")
                    f.write(text[: args.max_chars])
                    if len(text) > args.max_chars:
                        f.write("\n... (truncated)\n")
                    f.write("\n```\n\n")
            else:
                f.write("#### Raw value\n\n")
                f.write("```text\n")
                f.write(text[: args.max_chars])
                if len(text) > args.max_chars:
                    f.write("\n... (truncated)\n")
                f.write("\n```\n\n")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())


