#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
BIN="$(realpath ./go-build-analyzer)"
DB_PATH="${TOOLEXEC_DB:-$(realpath ../../../../build_times.db)}"

export TOOLEXEC_DB="$DB_PATH"
RUN_JSON="$($BIN runs-new --comment "scripted run $(date -u +%FT%TZ)" --output json)"
RUN_ID="$(printf '%s' "$RUN_JSON" | sed -n 's/.*"run_id":[ ]*\([0-9]*\).*/\1/p')"
export TOOLEXEC_RUN_ID="$RUN_ID"

echo "TOOLEXEC_DB=$TOOLEXEC_DB"
echo "TOOLEXEC_RUN_ID=$TOOLEXEC_RUN_ID"
