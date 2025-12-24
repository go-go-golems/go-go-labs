#!/usr/bin/env bash
set -euo pipefail

DB="${1:-}"
if [[ -z "${DB}" ]]; then
  echo "usage: $0 /abs/path/to/workspace/state.vscdb" >&2
  exit 2
fi

sqlite3 "${DB}" "
PRAGMA busy_timeout=2000;
SELECT key, LENGTH(value) AS size
FROM ItemTable
WHERE key LIKE '%aiService%' OR key LIKE '%composer%' OR key LIKE '%aichat%' OR key LIKE '%chat%'
ORDER BY size DESC, key ASC;
"


