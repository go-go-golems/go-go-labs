#!/usr/bin/env python3

"""
Export a Composer/Agent conversation from Cursor globalStorage/state.vscdb.

Key discovery: globalStorage/state.vscdb contains a table:
  cursorDiskKV(key TEXT UNIQUE ON CONFLICT REPLACE, value BLOB)

For a given composerId (conversation UUID), it stores:
  - composerData:<composerId>
  - bubbleId:<composerId>:<bubbleUUID>
  - checkpointId:<composerId>:<checkpointUUID>
  - codeBlockDiff:<composerId>:<uuid>
  - (possibly others)

This script exports:
  1) composerData summary (and headers-only bubble list)
  2) selected bubble payloads (all or limited), including “interesting” large string paths
  3) optional checkpoint payload list

Read-only.
"""

from __future__ import annotations

import argparse
import json
import sqlite3
from datetime import datetime
from typing import Any, Dict, Iterable, List, Optional, Tuple


SENSITIVE_SUBSTRINGS = (
    "secret",
    "token",
    "apikey",
    "api_key",
    "password",
    "encryptionkey",
    "encryption_key",
    "privatekey",
    "private_key",
)


def _redact(obj: Any) -> Any:
    if isinstance(obj, dict):
        out: Dict[str, Any] = {}
        for k, v in obj.items():
            lk = str(k).lower()
            if any(s in lk for s in SENSITIVE_SUBSTRINGS):
                out[k] = "<redacted>"
            else:
                out[k] = _redact(v)
        return out
    if isinstance(obj, list):
        return [_redact(x) for x in obj]
    return obj


def _to_text(val: Any) -> str:
    if val is None:
        return ""
    if isinstance(val, (bytes, bytearray)):
        return val.decode("utf-8", "replace")
    return str(val)


def _get_kv(cur: sqlite3.Cursor, key: str) -> Optional[str]:
    cur.execute("SELECT value FROM cursorDiskKV WHERE key = ?", (key,))
    row = cur.fetchone()
    if not row:
        return None
    return _to_text(row[0])


def _iter_keys(cur: sqlite3.Cursor, like: str) -> Iterable[Tuple[str, int]]:
    cur.execute("SELECT key, LENGTH(value) FROM cursorDiskKV WHERE key LIKE ? ORDER BY key ASC", (like,))
    yield from cur.fetchall()


def _walk_large_strings(x: Any, prefix: str = "", min_len: int = 200) -> List[Tuple[str, int, str]]:
    paths: List[Tuple[str, int, str]] = []

    def rec(v: Any, p: str) -> None:
        if isinstance(v, dict):
            for kk, vv in v.items():
                rec(vv, f"{p}.{kk}" if p else str(kk))
        elif isinstance(v, list):
            for i, vv in enumerate(v):
                # keep bounded
                if i >= 4000:
                    break
                rec(vv, f"{p}[{i}]")
        elif isinstance(v, str) and v.strip():
            if len(v) >= min_len:
                paths.append((p, len(v), v))

    rec(x, prefix)
    paths.sort(key=lambda t: -t[1])
    return paths


def _md_codeblock(f, lang: str, s: str) -> None:
    f.write(f"```{lang}\n")
    f.write(s)
    if not s.endswith("\n"):
        f.write("\n")
    f.write("```\n")


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--db", required=True, help="Path to globalStorage/state.vscdb")
    p.add_argument("--composer-id", required=True, help="composerId / conversation UUID")
    p.add_argument("--out", required=True, help="Output markdown path")
    p.add_argument("--max-chars", type=int, default=20000, help="Max chars per large value dump")
    p.add_argument("--max-bubbles", type=int, default=120, help="Max bubbles to export (0=all)")
    p.add_argument("--include-checkpoints", action="store_true", help="Also export checkpointId payloads (index only by default)")
    p.add_argument("--max-checkpoints", type=int, default=30, help="Max checkpoints to export (0=all)")
    p.add_argument("--min-large-string-len", type=int, default=400, help="Min length to include a string path dump")
    args = p.parse_args()

    conn = sqlite3.connect(args.db)
    cur = conn.cursor()

    composer_key = f"composerData:{args.composer_id}"
    composer_raw = _get_kv(cur, composer_key)
    if composer_raw is None:
        raise SystemExit(f"composerData not found for composerId: {args.composer_id}")

    composer_obj = _redact(json.loads(composer_raw))

    # Bubble headers (fast index)
    headers = composer_obj.get("fullConversationHeadersOnly", [])
    if not isinstance(headers, list):
        headers = []

    bubble_ids: List[str] = []
    for h in headers:
        if isinstance(h, dict) and isinstance(h.get("bubbleId"), str):
            bubble_ids.append(h["bubbleId"])

    # Fall back: enumerate bubbleId keys if headers absent
    if not bubble_ids:
        bubble_ids = [k.split(":")[-1] for (k, _sz) in _iter_keys(cur, f"bubbleId:{args.composer_id}:%")]

    # Bound export
    if args.max_bubbles != 0:
        bubble_ids = bubble_ids[: args.max_bubbles]

    with open(args.out, "w", encoding="utf-8") as f:
        f.write("---\n")
        f.write("Title: Cursor cursorDiskKV Composer Export\n")
        f.write("DocType: reference\n")
        f.write(f"GeneratedAt: {datetime.now().isoformat()}\n")
        f.write(f"SourceDB: {args.db}\n")
        f.write(f"ComposerId: {args.composer_id}\n")
        f.write("---\n\n")

        f.write("# Cursor `cursorDiskKV` Composer Export\n\n")
        f.write(f"- **DB**: `{args.db}`\n")
        f.write(f"- **composerId**: `{args.composer_id}`\n")
        f.write(f"- **bubble headers (fullConversationHeadersOnly)**: {len(headers)}\n")
        f.write(f"- **bubbleIds exported**: {len(bubble_ids)}\n\n")

        f.write("## composerData (summary)\n\n")
        # Keep it readable: omit the huge header list in this section
        composer_summary = dict(composer_obj)
        if "fullConversationHeadersOnly" in composer_summary:
            composer_summary["fullConversationHeadersOnly"] = f"<omitted: {len(headers)} headers>"
        _md_codeblock(f, "json", json.dumps(composer_summary, indent=2, ensure_ascii=False)[: args.max_chars])

        f.write("\n## Bubble headers (first 40)\n\n")
        _md_codeblock(f, "json", json.dumps(headers[:40], indent=2, ensure_ascii=False))

        f.write("\n## Bubbles\n\n")
        for i, bid in enumerate(bubble_ids):
            key = f"bubbleId:{args.composer_id}:{bid}"
            raw = _get_kv(cur, key)
            if raw is None:
                continue

            try:
                obj = _redact(json.loads(raw))
            except Exception:
                # dump raw
                f.write(f"### {i+1}. `{key}` (non-JSON)\n\n")
                _md_codeblock(f, "text", raw[: args.max_chars])
                continue

            f.write(f"### {i+1}. `{key}`\n\n")
            f.write(f"- **type**: `{obj.get('type')}`\n")
            if isinstance(obj.get("createdAt"), str):
                f.write(f"- **createdAt**: `{obj.get('createdAt')}`\n")
            if isinstance(obj.get("requestId"), str):
                f.write(f"- **requestId**: `{obj.get('requestId')}`\n")
            if isinstance(obj.get("checkpointId"), str):
                f.write(f"- **checkpointId**: `{obj.get('checkpointId')}`\n")
            f.write("\n")

            # Prefer "text" if present
            text = obj.get("text")
            if isinstance(text, str) and text.strip():
                f.write("#### text\n\n")
                _md_codeblock(f, "text", text[: args.max_chars])

            # Extract “large string paths” (tool results, code blocks, etc.)
            large_paths = _walk_large_strings(obj, min_len=args.min_large_string_len)
            if large_paths:
                f.write("#### large string paths (top 15)\n\n")
                for pth, ln, s in large_paths[:15]:
                    f.write(f"- `{pth}` (len={ln})\n")
                f.write("\n")

                # dump the top few large strings verbatim (truncated)
                f.write("#### large string dumps (top 5)\n\n")
                for pth, ln, s in large_paths[:5]:
                    f.write(f"##### `{pth}` (len={ln})\n\n")
                    _md_codeblock(f, "text", s[: args.max_chars])

            # Always include a small JSON slice for structure
            f.write("#### JSON (structure)\n\n")
            _md_codeblock(f, "json", json.dumps(obj, indent=2, ensure_ascii=False)[: args.max_chars])

        if args.include_checkpoints:
            f.write("\n## Checkpoints\n\n")
            ck_keys = list(_iter_keys(cur, f"checkpointId:{args.composer_id}:%"))
            if args.max_checkpoints != 0:
                ck_keys = ck_keys[: args.max_checkpoints]

            f.write(f"- checkpoint keys: {len(ck_keys)} (exported)\n\n")
            for k, size in ck_keys:
                f.write(f"- `{k}` (size={size})\n")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())


