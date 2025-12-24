#!/usr/bin/env python3

"""
Read-only inspector for Cursor per-workspace conversation persistence.

Targets a VS Code-style SQLite KV database (state.vscdb) and prints:
- composer.composerData (allComposers + matching composerId)
- aiService.generations (generationUUID/type/textDescription/unixMs)
- aiService.prompts (text/commandType)

This is designed to reproduce our findings that:
  - conversation UUID == composerId
  - generation UUID == generationUUID
  - prompts live in aiService.generations + aiService.prompts
"""

from __future__ import annotations

import argparse
import json
import sqlite3
from typing import Any, Dict, List, Optional


def _get_value(cur: sqlite3.Cursor, key: str) -> Optional[str]:
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


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--workspace-db", required=True, help="Path to workspace state.vscdb (SQLite)")
    p.add_argument("--composer-id", required=False, help="Composer/conversation UUID to focus on")
    p.add_argument("--limit", type=int, default=50, help="Max generations/prompts to print (0 = unlimited)")
    args = p.parse_args()

    conn = sqlite3.connect(args.workspace_db)
    cur = conn.cursor()

    composer_raw = _get_value(cur, "composer.composerData")
    gens_raw = _get_value(cur, "aiService.generations")
    prompts_raw = _get_value(cur, "aiService.prompts")

    out: Dict[str, Any] = {"workspaceDb": args.workspace_db, "composerId": args.composer_id}

    if composer_raw:
        composer = json.loads(composer_raw)
        out["composer.composerData.keys"] = list(composer.keys())
        all_composers = composer.get("allComposers", [])
        out["composer.composerData.allComposers.count"] = len(all_composers)
        if args.composer_id:
            match = next((c for c in all_composers if c.get("composerId") == args.composer_id), None)
            out["composer.match"] = match
        else:
            out["composer.allComposers.sample"] = all_composers[:3]
    else:
        out["composer.composerData"] = None

    def _limit(arr: List[Any]) -> List[Any]:
        if args.limit == 0:
            return arr
        return arr[: args.limit]

    if gens_raw:
        gens = json.loads(gens_raw)
        out["aiService.generations.count"] = len(gens)
        out["aiService.generations.sample"] = _limit(gens)
    else:
        out["aiService.generations"] = None

    if prompts_raw:
        prompts = json.loads(prompts_raw)
        out["aiService.prompts.count"] = len(prompts)
        out["aiService.prompts.sample"] = _limit(prompts)
    else:
        out["aiService.prompts"] = None

    print(json.dumps(out, indent=2, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())


