#!/usr/bin/env python3

"""
Locate raw byte offsets of a needle inside a SQLite DB file (state.vscdb).

Outputs:
- absolute byte offsets
- computed page number (1-based) given page_size
- offset within page

This helps reconcile "grep: binary file matches" with the owning sqlite object
via dbstat (page ownership).
"""

from __future__ import annotations

import argparse
from pathlib import Path
from typing import Iterable, List, Tuple


def iter_matches(path: Path, needle: bytes, chunk_size: int) -> Iterable[int]:
    with path.open("rb") as f:
        overlap = b""
        offset_base = 0
        while True:
            chunk = f.read(chunk_size)
            if not chunk:
                break
            data = overlap + chunk
            start = 0
            while True:
                idx = data.find(needle, start)
                if idx == -1:
                    break
                yield offset_base - len(overlap) + idx
                start = idx + 1
            # keep overlap so we can match across chunk boundaries
            if len(needle) > 1:
                overlap = data[-(len(needle) - 1) :]
            else:
                overlap = b""
            offset_base += len(chunk)


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--file", required=True, help="Path to sqlite file (e.g. state.vscdb)")
    p.add_argument("--needle", required=True, help="Needle string to search for (utf-8)")
    p.add_argument("--page-size", type=int, required=True, help="SQLite page size in bytes (PRAGMA page_size)")
    p.add_argument("--chunk-size", type=int, default=8 * 1024 * 1024, help="Read chunk size")
    p.add_argument("--max-hits", type=int, default=50, help="Stop after N matches")
    p.add_argument("--context", type=int, default=120, help="Show +/- context bytes around each match (utf-8 lossy)")
    args = p.parse_args()

    file_path = Path(args.file)
    needle = args.needle.encode("utf-8")

    hits: List[int] = []
    for off in iter_matches(file_path, needle, args.chunk_size):
        hits.append(off)
        if len(hits) >= args.max_hits:
            break

    print(f"file\t{file_path}")
    print(f"needle\t{args.needle!r}")
    print(f"page_size\t{args.page_size}")
    print(f"hits\t{len(hits)}")

    if not hits:
        return 0

    # Avoid loading whole file for huge DBs: if >256MB, read per-hit windows instead.
    file_size = file_path.stat().st_size
    data = b""
    if file_size <= 256 * 1024 * 1024:
        data = file_path.read_bytes()

    for i, off in enumerate(hits):
        pgno = off // args.page_size + 1
        in_pg = off % args.page_size
        print(f"\n== hit {i+1} ==")
        print(f"offset\t{off}")
        print(f"page\t{pgno}")
        print(f"offset_in_page\t{in_pg}")

        if args.context > 0:
            if data:
                start = max(0, off - args.context)
                end = min(len(data), off + len(needle) + args.context)
                snippet = data[start:end]
            else:
                with file_path.open("rb") as f:
                    start = max(0, off - args.context)
                    f.seek(start)
                    snippet = f.read(len(needle) + 2 * args.context)
            print("context_utf8")
            print(snippet.decode("utf-8", "replace"))

    return 0


if __name__ == "__main__":
    raise SystemExit(main())


