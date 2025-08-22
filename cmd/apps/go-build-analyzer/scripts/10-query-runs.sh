#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
BIN="$(realpath ./go-build-analyzer)"

echo "# Runs (table)"
$BIN runs-list --output table

echo

echo "# Runs (json)"
$BIN runs-list --output json
