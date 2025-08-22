#!/usr/bin/env bash
set -euo pipefail

# Build the analyzer binary in-place
cd "$(dirname "$0")/.."
go build -o ./go-build-analyzer .
echo "Built: $(realpath ./go-build-analyzer)"


