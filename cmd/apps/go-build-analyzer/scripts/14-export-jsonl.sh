#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
BIN="$(realpath ./go-build-analyzer)"
RUN_ID="$($BIN runs-list --output json | jq -r '.[0].run_id')"

if [[ -z "$RUN_ID" || "$RUN_ID" == "null" ]]; then
  echo "No runs found. Create one first (scripts/02-new-run-and-export.sh) and build (scripts/03-instrumented-build.sh)." >&2
  exit 1
fi

OUT="invocations-run-$RUN_ID.jsonl"
$BIN invocations-list --run-id "$RUN_ID" --limit 100000 --output json | jq -c '.[]' > "$OUT"
echo "Wrote $OUT"
