#!/usr/bin/env python3
"""
Build a SQLite database with git history metadata and surface TypeScript insights.

The database captures:
  * commits with parent relationships
  * per-commit file level stats (status, additions, deletions, rename metadata)
  * file catalog (extension, TypeScript/doc flags)
  * extracted TypeScript symbols from the current HEAD
  * markdown document section index

After populating the database, the script prints a short TypeScript-focused analysis
to help with targeted reviews.
"""

from __future__ import annotations

import argparse
import re
import sqlite3
import subprocess
import sys
import textwrap
from collections import defaultdict
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, Iterable, Iterator, List, Optional, Tuple

REPO_ROOT = Path(__file__).resolve().parents[1]
DB_PATH = REPO_ROOT / "sqleton" / "mento" / "git_history.db"


class GitError(RuntimeError):
    pass


def run_git(args: List[str]) -> str:
    """Run a git command from the repository root and return stdout."""
    completed = subprocess.run(
        ["git", *args],
        cwd=REPO_ROOT,
        check=False,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    if completed.returncode != 0:
        raise GitError(f"git {' '.join(args)} failed: {completed.stderr.strip()}")
    return completed.stdout


def ensure_db_directory() -> None:
    DB_PATH.parent.mkdir(parents=True, exist_ok=True)


def init_db(conn: sqlite3.Connection, rebuild: bool = True) -> None:
    """Create (or recreate) the schema."""
    conn.execute("PRAGMA journal_mode=WAL;")
    conn.execute("PRAGMA foreign_keys=ON;")
    if rebuild:
        conn.executescript(
            """
            DROP TABLE IF EXISTS doc_sections;
            DROP TABLE IF EXISTS file_symbols;
            DROP TABLE IF EXISTS commit_files;
            DROP TABLE IF EXISTS commit_parents;
            DROP TABLE IF EXISTS commits;
            DROP TABLE IF EXISTS files;
            """
        )
    conn.executescript(
        """
        CREATE TABLE IF NOT EXISTS commits (
            hash TEXT PRIMARY KEY,
            tree TEXT,
            author_name TEXT,
            author_email TEXT,
            author_date TEXT,
            committer_name TEXT,
            committer_email TEXT,
            committer_date TEXT,
            subject TEXT,
            body TEXT
        );

        CREATE TABLE IF NOT EXISTS commit_parents (
            commit_hash TEXT,
            parent_hash TEXT,
            PRIMARY KEY (commit_hash, parent_hash),
            FOREIGN KEY (commit_hash) REFERENCES commits(hash) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS files (
            id INTEGER PRIMARY KEY,
            path TEXT UNIQUE,
            extension TEXT,
            is_typescript INTEGER DEFAULT 0,
            is_documentation INTEGER DEFAULT 0
        );

        CREATE TABLE IF NOT EXISTS commit_files (
            commit_hash TEXT,
            file_id INTEGER,
            path TEXT,
            status TEXT,
            additions INTEGER,
            deletions INTEGER,
            is_binary INTEGER DEFAULT 0,
            old_path TEXT,
            PRIMARY KEY (commit_hash, path),
            FOREIGN KEY (commit_hash) REFERENCES commits(hash) ON DELETE CASCADE,
            FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS file_symbols (
            file_id INTEGER,
            symbol_name TEXT,
            symbol_type TEXT,
            line_start INTEGER,
            doc TEXT,
            PRIMARY KEY (file_id, symbol_name, line_start),
            FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS doc_sections (
            file_id INTEGER,
            heading TEXT,
            level INTEGER,
            line_start INTEGER,
            PRIMARY KEY (file_id, line_start),
            FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
        );

        CREATE INDEX IF NOT EXISTS idx_commit_files_file_id ON commit_files(file_id);
        CREATE INDEX IF NOT EXISTS idx_commit_files_commit_hash ON commit_files(commit_hash);
        CREATE INDEX IF NOT EXISTS idx_files_extension ON files(extension);
        CREATE INDEX IF NOT EXISTS idx_file_symbols_file_id ON file_symbols(file_id);
        CREATE INDEX IF NOT EXISTS idx_doc_sections_file_id ON doc_sections(file_id);
        """
    )


@dataclass
class CommitRecord:
    hash: str
    tree: str
    parents: List[str]
    author_name: str
    author_email: str
    author_date: str
    committer_name: str
    committer_email: str
    committer_date: str
    subject: str
    body: str
    numstat: List[Tuple[str, str, str]]  # additions, deletions, path field


def parse_commit_chunk(chunk: str) -> Optional[CommitRecord]:
    text = chunk.strip()
    if not text:
        return None
    if "\x1d" in text:
        header_part, numstat_part = text.split("\x1d", 1)
    else:
        header_part, numstat_part = text, ""
    parts = header_part.split("\x1f")
    if len(parts) < 10:
        parts.extend([""] * (10 - len(parts)))
    commit_hash = parts[0]
    tree_hash = parts[1]
    parents_str = parts[2]
    author_name = parts[3]
    author_email = parts[4]
    author_date = parts[5]
    committer_name = parts[6]
    committer_email = parts[7]
    committer_date = parts[8]
    subject = parts[9]
    body = parts[10] if len(parts) > 10 else ""
    numstat_entries: List[Tuple[str, str, str]] = []
    for line in numstat_part.splitlines():
        if not line.strip():
            continue
        fields = line.split("\t")
        if len(fields) < 3:
            continue
        numstat_entries.append((fields[0], fields[1], fields[2]))
    parents = [p for p in parents_str.split() if p]
    return CommitRecord(
        hash=commit_hash,
        tree=tree_hash,
        parents=parents,
        author_name=author_name,
        author_email=author_email,
        author_date=author_date,
        committer_name=committer_name,
        committer_email=committer_email,
        committer_date=committer_date,
        subject=subject,
        body=body,
        numstat=numstat_entries,
    )


def iter_commits() -> Iterator[CommitRecord]:
    fmt = "%x1e%H%x1f%T%x1f%P%x1f%an%x1f%ae%x1f%ad%x1f%cn%x1f%ce%x1f%cd%x1f%s%x1f%b%x1d"
    raw = run_git(["log", "--date=iso-strict", "--pretty=format:" + fmt, "--numstat"])
    for chunk in raw.split("\x1e"):
        rec = parse_commit_chunk(chunk)
        if rec:
            yield rec


def get_name_status(commit_hash: str) -> Dict[str, Tuple[str, Optional[str]]]:
    """
    Return a mapping of file path at commit -> (status, old_path).
    Status uses the one-letter code (A/M/D/R/C/T).
    """
    result: Dict[str, Tuple[str, Optional[str]]] = {}
    out = run_git(["show", "--name-status", "--format=", commit_hash])
    for line in out.splitlines():
        if not line.strip():
            continue
        parts = line.split("\t")
        if not parts:
            continue
        status_code = parts[0]
        if status_code.startswith(("R", "C")) and len(parts) >= 3:
            status = status_code[0]
            old_path = parts[1]
            new_path = parts[2]
            result[new_path] = (status, old_path)
        elif len(parts) >= 2:
            status = status_code[0]
            path = parts[1]
            result[path] = (status, None)
    return result


def normalize_path_field(path_field: str) -> Tuple[str, Optional[str]]:
    """
    Handle rename markers in numstat output, e.g. src/foo.ts => src/bar.ts.
    Returns (path_at_commit, old_path_if_available).
    """
    arrow_tokens = (" => ", " -> ")
    for token in arrow_tokens:
        if token in path_field:
            left, right = path_field.split(token, 1)
            return right.strip(), left.strip()
    return path_field, None


class FileCatalog:
    """Track known file IDs to avoid redundant lookups."""

    def __init__(self, conn: sqlite3.Connection):
        self.conn = conn
        self.cache: Dict[str, int] = {}

    @staticmethod
    def extension_for(path: str) -> Optional[str]:
        return Path(path).suffix.lower() or None

    def get_id(self, path: str) -> int:
        if path in self.cache:
            return self.cache[path]
        cur = self.conn.execute("SELECT id FROM files WHERE path = ?", (path,))
        row = cur.fetchone()
        if row:
            file_id = row[0]
        else:
            ext = self.extension_for(path)
            cur = self.conn.execute(
                "INSERT OR IGNORE INTO files(path, extension) VALUES(?, ?)",
                (path, ext),
            )
            if cur.lastrowid:
                file_id = cur.lastrowid
            else:
                # Row already existed but we lost the id; fetch again.
                file_id = self.conn.execute("SELECT id FROM files WHERE path = ?", (path,)).fetchone()[0]
        self.cache[path] = file_id
        return file_id

    def ensure_flag(self, path: str, column: str, value: int = 1) -> None:
        self.conn.execute(f"UPDATE files SET {column} = ? WHERE path = ?", (value, path))


INTERESTING_SUFFIXES = {".ts", ".tsx", ".md", ".mdx"}


def is_interesting_path(path: Optional[str]) -> bool:
    if not path:
        return False
    suffix = Path(path).suffix.lower()
    return suffix in INTERESTING_SUFFIXES


def import_history(conn: sqlite3.Connection) -> None:
    catalog = FileCatalog(conn)
    cur = conn.cursor()
    for commit in iter_commits():
        cur.execute(
            """
            INSERT OR REPLACE INTO commits(hash, tree, author_name, author_email, author_date,
                                           committer_name, committer_email, committer_date,
                                           subject, body)
            VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """,
            (
                commit.hash,
                commit.tree,
                commit.author_name,
                commit.author_email,
                commit.author_date,
                commit.committer_name,
                commit.committer_email,
                commit.committer_date,
                commit.subject,
                commit.body,
            ),
        )
        cur.executemany(
            "INSERT OR REPLACE INTO commit_parents(commit_hash, parent_hash) VALUES(?, ?)",
            [(commit.hash, parent) for parent in commit.parents],
        )
        status_map = get_name_status(commit.hash)
        seen_paths: set[str] = set()
        for add_str, del_str, path_field in commit.numstat:
            path, inferred_old = normalize_path_field(path_field)
            if not is_interesting_path(path):
                continue
            additions = None if add_str.strip() == "-" else int(add_str)
            deletions = None if del_str.strip() == "-" else int(del_str)
            is_binary = int(add_str.strip() == "-" or del_str.strip() == "-")
            status, old_path = status_map.get(path, (None, None))
            if status is None and path_field in status_map:
                status, old_path = status_map[path_field]
            if status is None:
                if additions is None or deletions is None:
                    status = "M"
                elif additions == 0 and deletions > 0:
                    status = "D"
                elif additions > 0 and deletions == 0:
                    status = "A"
                else:
                    status = "M"
            if old_path is None:
                old_path = inferred_old
            file_id = catalog.get_id(path)
            cur.execute(
                """
                INSERT OR REPLACE INTO commit_files
                    (commit_hash, file_id, path, status, additions, deletions, is_binary, old_path)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    commit.hash,
                    file_id,
                    path,
                    status,
                    additions,
                    deletions,
                    is_binary,
                    old_path,
                ),
            )
            seen_paths.add(path)
        # Capture files that only appear in the name-status output (e.g., deletions with no numstat)
        for path, (status, old_path) in status_map.items():
            if not is_interesting_path(path):
                continue
            if path in seen_paths:
                continue
            file_id = catalog.get_id(path)
            cur.execute(
                """
                INSERT OR REPLACE INTO commit_files
                    (commit_hash, file_id, path, status, additions, deletions, is_binary, old_path)
                VALUES (?, ?, ?, ?, NULL, NULL, 0, ?)
                """,
                (
                    commit.hash,
                    file_id,
                    path,
                    status,
                    old_path,
                ),
            )
    conn.commit()


SYMBOL_PATTERNS: List[Tuple[re.Pattern[str], str]] = [
    (re.compile(r"export\s+default\s+class\s+(?P<name>[A-Za-z0-9_$]+)"), "class"),
    (re.compile(r"export\s+default\s+function\s+(?P<name>[A-Za-z0-9_$]+)"), "function"),
    (re.compile(r"export\s+(?:abstract\s+)?class\s+(?P<name>[A-Za-z0-9_$]+)"), "class"),
    (re.compile(r"export\s+interface\s+(?P<name>[A-Za-z0-9_$]+)"), "interface"),
    (re.compile(r"export\s+type\s+(?P<name>[A-Za-z0-9_$]+)"), "type"),
    (re.compile(r"export\s+enum\s+(?P<name>[A-Za-z0-9_$]+)"), "enum"),
    (re.compile(r"export\s+(?:async\s+)?function\s+(?P<name>[A-Za-z0-9_$]+)"), "function"),
    (re.compile(r"export\s+const\s+(?P<name>[A-Za-z0-9_$]+)"), "const"),
    (re.compile(r"export\s+let\s+(?P<name>[A-Za-z0-9_$]+)"), "const"),
    (re.compile(r"export\s+var\s+(?P<name>[A-Za-z0-9_$]+)"), "const"),
]


def collapse_doc_lines(lines: List[str]) -> str:
    cleaned: List[str] = []
    for line in lines:
        stripped = line.strip()
        if stripped.startswith("*"):
            stripped = stripped.lstrip("*").strip()
        if stripped:
            cleaned.append(stripped)
    return "\n".join(cleaned).strip()


def extract_symbols(path: Path, text: str) -> List[Tuple[str, str, int, Optional[str]]]:
    """Return tuples of (symbol_name, symbol_type, line_number, doc)."""
    results: List[Tuple[str, str, int, Optional[str]]] = []
    pending_doc: Optional[str] = None
    in_doc = False
    doc_lines: List[str] = []
    for lineno, line in enumerate(text.splitlines(), start=1):
        stripped = line.strip()

        if in_doc:
            end_idx = stripped.find("*/")
            if end_idx != -1:
                before = stripped[:end_idx].strip()
                if before:
                    doc_lines.append(before)
                pending_doc = collapse_doc_lines(doc_lines)
                doc_lines = []
                in_doc = False
                remainder = stripped[end_idx + 2 :].strip()
                if remainder:
                    stripped = remainder
                else:
                    continue
            else:
                if stripped:
                    doc_lines.append(stripped)
                continue

        if stripped.startswith("/**"):
            in_doc = True
            doc_lines = []
            content = stripped[3:].strip()
            if content.endswith("*/"):
                content = content[:-2].strip()
                if content:
                    doc_lines.append(content)
                pending_doc = collapse_doc_lines(doc_lines)
                doc_lines = []
                in_doc = False
            else:
                if content:
                    doc_lines.append(content)
            continue

        if not stripped:
            continue

        doc_text = pending_doc
        recorded = False
        for pattern, symbol_type in SYMBOL_PATTERNS:
            match = pattern.search(line)
            if match:
                name = match.group("name")
                results.append((name, symbol_type, lineno, doc_text))
                pending_doc = None
                recorded = True
                break
        if not recorded and doc_text is not None:
            pending_doc = None
    return results


def populate_typescript_metadata(conn: sqlite3.Connection, catalog: FileCatalog) -> None:
    ts_files = [
        line.strip()
        for line in run_git(["ls-files", "--", "*.ts", "*.tsx"]).splitlines()
        if line.strip()
    ]
    symbol_rows: List[Tuple[int, str, str, int, Optional[str]]] = []
    for rel_path in ts_files:
        file_path = REPO_ROOT / rel_path
        if not file_path.exists():
            continue
        try:
            text = file_path.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            continue
        symbols = extract_symbols(file_path, text)
        file_id = catalog.get_id(rel_path)
        catalog.ensure_flag(rel_path, "is_typescript", 1)
        for name, symbol_type, line_no, doc in symbols:
            symbol_rows.append((file_id, name, symbol_type, line_no, doc))
    conn.execute("DELETE FROM file_symbols")
    conn.executemany(
        "INSERT INTO file_symbols(file_id, symbol_name, symbol_type, line_start, doc) VALUES(?, ?, ?, ?, ?)",
        symbol_rows,
    )
    conn.commit()


def populate_doc_sections(conn: sqlite3.Connection, catalog: FileCatalog) -> None:
    markdown_files = [
        line.strip()
        for line in run_git(["ls-files", "--", "*.md"]).splitlines()
        if line.strip()
    ]
    rows: List[Tuple[int, str, int, int]] = []
    for rel_path in markdown_files:
        file_path = REPO_ROOT / rel_path
        if not file_path.exists():
            continue
        try:
            text = file_path.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            continue
        file_id = catalog.get_id(rel_path)
        catalog.ensure_flag(rel_path, "is_documentation", 1)
        for lineno, line in enumerate(text.splitlines(), start=1):
            stripped = line.strip()
            if not stripped.startswith("#"):
                continue
            level = len(stripped) - len(stripped.lstrip("#"))
            heading = stripped.lstrip("#").strip()
            if not heading:
                continue
            rows.append((file_id, heading, level, lineno))
    conn.execute("DELETE FROM doc_sections")
    conn.executemany(
        "INSERT INTO doc_sections(file_id, heading, level, line_start) VALUES(?, ?, ?, ?)",
        rows,
    )
    conn.commit()


def analyze_typescript(conn: sqlite3.Connection) -> str:
    cur = conn.cursor()
    total_ts = cur.execute(
        "SELECT COUNT(*) FROM files WHERE is_typescript = 1"
    ).fetchone()[0]
    symbols = cur.execute(
        "SELECT symbol_type, COUNT(*) FROM file_symbols GROUP BY symbol_type ORDER BY count(*) DESC"
    ).fetchall()
    symbol_docs = cur.execute(
        "SELECT COUNT(*) FROM file_symbols WHERE doc IS NOT NULL AND TRIM(doc) <> ''"
    ).fetchone()[0]
    most_changed = cur.execute(
        """
        SELECT f.path,
               COUNT(*) AS commits,
               COALESCE(SUM(COALESCE(cf.additions, 0)), 0) AS additions,
               COALESCE(SUM(COALESCE(cf.deletions, 0)), 0) AS deletions
        FROM commit_files cf
        JOIN files f ON f.id = cf.file_id
        WHERE f.is_typescript = 1
        GROUP BY f.id
        ORDER BY commits DESC, f.path
        LIMIT 10
        """
    ).fetchall()
    recent_commits = cur.execute(
        """
        SELECT c.hash,
               c.subject,
               c.author_name,
               c.author_date,
               COUNT(*) AS files_touched
        FROM commit_files cf
        JOIN files f ON f.id = cf.file_id
        JOIN commits c ON c.hash = cf.commit_hash
        WHERE f.is_typescript = 1
        GROUP BY c.hash
        ORDER BY c.author_date DESC
        LIMIT 5
        """
    ).fetchall()
    lines = [
        f"Total TypeScript files tracked: {total_ts}",
        f"Documented symbols: {symbol_docs}/{max(1, sum(count for _, count in symbols))}",
    ]
    if symbols:
        sym_desc = ", ".join(f"{stype}={count}" for stype, count in symbols)
        lines.append(f"Symbols extracted ({sym_desc})")
    if most_changed:
        lines.append("Most edited TypeScript paths (commit count, +, -):")
        for path, commits, adds, dels in most_changed:
            lines.append(f"  - {path}: {commits} commits, +{adds}, -{dels}")
    if recent_commits:
        lines.append("Recent commits touching TypeScript:")
        for commit_hash, subject, author, date, file_count in recent_commits:
            lines.append(f"  - {commit_hash[:10]} {date} {author}: {subject} ({file_count} files)")
    return "\n".join(lines)


def build_database(rebuild: bool = True) -> str:
    ensure_db_directory()
    conn = sqlite3.connect(DB_PATH)
    try:
        init_db(conn, rebuild=rebuild)
        conn.execute("BEGIN")
        import_history(conn)
        catalog = FileCatalog(conn)
        populate_typescript_metadata(conn, catalog)
        populate_doc_sections(conn, catalog)
        analysis = analyze_typescript(conn)
        return analysis
    finally:
        conn.close()


def main(argv: Optional[List[str]] = None) -> int:
    parser = argparse.ArgumentParser(description="Index git history into SQLite for targeted reviews.")
    parser.add_argument(
        "--no-rebuild",
        action="store_true",
        help="Reuse the existing schema (keep prior data)",
    )
    args = parser.parse_args(argv)
    try:
        analysis = build_database(rebuild=not args.no_rebuild)
    except GitError as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1
    except sqlite3.DatabaseError as exc:
        print(f"sqlite error: {exc}", file=sys.stderr)
        return 1
    print(
        textwrap.dedent(
            f"""
            SQLite database created at {DB_PATH}

            {analysis}
            """
        ).strip()
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
