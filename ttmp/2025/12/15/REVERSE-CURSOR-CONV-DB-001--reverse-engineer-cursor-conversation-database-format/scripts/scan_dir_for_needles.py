#!/usr/bin/env python3

"""
Scan files under a directory for one or more byte-string needles.

Use this for binary stores like Electron LevelDB / Session Storage / blob stores,
where grep may not be effective or may be too slow/noisy.
"""

from __future__ import annotations

import argparse
import os
from pathlib import Path
from typing import Dict, List, Tuple


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--root", required=True, help="Directory to scan")
    p.add_argument("--needle", action="append", required=True, help="Needle string (repeatable)")
    p.add_argument("--max-file-bytes", type=int, default=8 * 1024 * 1024, help="Skip files larger than this")
    p.add_argument("--max-hits", type=int, default=50, help="Stop after N total hits")
    p.add_argument("--ext", action="append", default=None, help="Only scan files with this suffix (repeatable), e.g. .ldb .log")
    args = p.parse_args()

    root = Path(args.root)
    needles: Dict[str, bytes] = {n: n.encode("utf-8") for n in args.needle}

    hits: List[Tuple[str, str, int]] = []

    for path in root.rglob("*"):
        if not path.is_file():
            continue
        try:
            st = path.stat()
        except FileNotFoundError:
            continue
        if st.st_size > args.max_file_bytes:
            continue
        if args.ext and path.suffix not in set(args.ext):
            continue
        try:
            data = path.read_bytes()
        except Exception:
            continue
        for name, bneedle in needles.items():
            idx = data.find(bneedle)
            if idx != -1:
                hits.append((name, str(path), idx))
                if len(hits) >= args.max_hits:
                    break
        if len(hits) >= args.max_hits:
            break

    print(f"root\t{args.root}")
    print(f"needles\t{len(needles)}")
    print(f"hits\t{len(hits)}")
    for name, path, idx in hits:
        print(f"{name}\t{idx}\t{path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())


