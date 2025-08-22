#!/usr/bin/env bash
set -euo pipefail

# Resolve important paths (repo root contains both glazed and go-go-labs)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
APP_DIR="$(realpath "$SCRIPT_DIR/..")"
REPO_ROOT="$(realpath "$APP_DIR/../../../..")"
BIN_PATH="$(realpath "$APP_DIR/go-build-analyzer")"

if [[ ! -x "$BIN_PATH" ]]; then
  echo "Analyzer binary not found at $BIN_PATH. Run scripts/01-build-binary.sh first." >&2
  exit 1
fi

pushd "$REPO_ROOT" >/dev/null

go clean -cache || true

pushd glazed >/dev/null
  go build -a -toolexec="$BIN_PATH" ./... || true
popd >/dev/null

pushd go-go-labs >/dev/null
  go build -a -toolexec="$BIN_PATH" ./... || true
popd >/dev/null

popd >/dev/null

echo "Instrumented builds completed (some packages may have failed, which is ok for logging)."
