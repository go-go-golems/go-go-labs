#!/usr/bin/env bash
set -euo pipefail

LOGS_ROOT=""
CONV_ID=""
GEN_ID=""
PHRASE=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --logs-root) LOGS_ROOT="$2"; shift 2 ;;
    --conversation-id) CONV_ID="$2"; shift 2 ;;
    --generation-id) GEN_ID="$2"; shift 2 ;;
    --phrase) PHRASE="$2"; shift 2 ;;
    *) echo "unknown arg: $1" >&2; exit 2 ;;
  esac
done

if [[ -z "${LOGS_ROOT}" ]]; then
  echo "usage: $0 --logs-root /home/manuel/.config/Cursor/logs [--conversation-id UUID] [--generation-id UUID] [--phrase TEXT]" >&2
  exit 2
fi

PATTERNS=()
if [[ -n "${CONV_ID}" ]]; then PATTERNS+=("${CONV_ID}"); fi
if [[ -n "${GEN_ID}" ]]; then PATTERNS+=("${GEN_ID}"); fi
if [[ -n "${PHRASE}" ]]; then PATTERNS+=("${PHRASE}"); fi

if [[ ${#PATTERNS[@]} -eq 0 ]]; then
  echo "provide at least one of --conversation-id / --generation-id / --phrase" >&2
  exit 2
fi

for p in "${PATTERNS[@]}"; do
  echo "=== grep: ${p} ==="
  grep -R -n -- "${p}" "${LOGS_ROOT}" | head -200 || true
  echo
done


