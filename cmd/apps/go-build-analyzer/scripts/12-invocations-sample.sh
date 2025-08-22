#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
BIN="$(realpath ./go-build-analyzer)"

RUN_ID="$($BIN runs-list --output json | jq -r '.[0].run_id')"
if [[ -z "$RUN_ID" || "$RUN_ID" == "null" ]]; then
  echo "No runs found. Create one first (scripts/02-new-run-and-export.sh) and build (scripts/03-instrumented-build.sh)." >&2
  exit 1
fi

echo "# Recent compile invocations (table)"
$BIN invocations-list --run-id "$RUN_ID" --tool compile --limit 10 --output table

echo

echo "# Recent compile invocations (json)"
$BIN invocations-list --run-id "$RUN_ID" --tool compile --limit 5 --output json
