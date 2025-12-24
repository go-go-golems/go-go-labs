#!/usr/bin/env python3

"""
Extract a readable transcript from workbench.panel.aichat.view.aichat.chatdata.

This key stores a JSON object with tabs and bubbles. AI bubbles look like:
  { "type": "ai", "rawText": "...", "modelType": "...", "requestId": "..." }
User bubbles are more complex and often store the entered text in delegate.a / delegate.b.
"""

from __future__ import annotations

import argparse
import json
import sqlite3
from typing import Any, Dict, List, Optional


def _get_text(cur: sqlite3.Cursor, key: str) -> Optional[str]:
    cur.execute("SELECT value FROM ItemTable WHERE key = ?", (key,))
    row = cur.fetchone()
    if not row:
        return None
    val = row[0]
    if val is None:
        return None
    if isinstance(val, bytes):
        return val.decode("utf-8", "replace")
    return str(val)


def _user_text(bubble: Dict[str, Any]) -> str:
    # Prefer delegate.a if present (plain string)
    delegate = bubble.get("delegate") or {}
    a = delegate.get("a")
    if isinstance(a, str) and a:
        return a
    # rawText is often empty for user bubbles; keep as fallback
    rt = bubble.get("rawText")
    if isinstance(rt, str) and rt:
        return rt
    return ""


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--db", required=True, help="Path to state.vscdb")
    p.add_argument("--tab-id", default=None, help="Optional: only print this tabId")
    p.add_argument("--max-chars", type=int, default=2000, help="Max chars per message")
    args = p.parse_args()

    conn = sqlite3.connect(args.db)
    cur = conn.cursor()
    raw = _get_text(cur, "workbench.panel.aichat.view.aichat.chatdata")
    if not raw:
        raise SystemExit("chatdata key not found")
    data = json.loads(raw)

    tabs: List[Dict[str, Any]] = data.get("tabs", [])
    if args.tab_id:
        tabs = [t for t in tabs if t.get("tabId") == args.tab_id]

    for t in tabs:
        print(f"=== tabId: {t.get('tabId')} title: {t.get('chatTitle','')} ===")
        bubbles = t.get("bubbles", [])
        for b in bubbles:
            btype = b.get("type")
            if btype == "ai":
                txt = b.get("rawText", "") or ""
                txt = str(txt)[: args.max_chars]
                print(f"assistant: {txt}")
            elif btype == "user":
                txt = _user_text(b)[: args.max_chars]
                print(f"user: {txt}")
            else:
                # other bubble types exist; keep lightweight
                print(f"{btype}: <omitted>")
        print()

    return 0


if __name__ == "__main__":
    raise SystemExit(main())


