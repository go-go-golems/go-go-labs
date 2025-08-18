#!/usr/bin/env bash
set -euo pipefail

# Self-contained tmux runner for mcp-oidc-server
SESSION="mcp-oidc-server"
MODE="capture" # default non-interactive; pass --attach to attach instead
if [[ "${1:-}" == "--attach" ]]; then MODE="attach"; fi

cd "$(dirname "$0")/../../.."

# Build
echo "Building mcp-oidc-server..." >&2
go build -o ./cmd/apps/mcp-oidc-server/mcp-oidc-server ./cmd/apps/mcp-oidc-server

# Ensure a clean session name
if tmux has-session -t "$SESSION" 2>/dev/null; then
  tmux kill-session -t "$SESSION" || true
fi

ADDR="${ADDR:-:8080}"
ISSUER="${ISSUER:-http://localhost:8080}"
LOG_FORMAT="${LOG_FORMAT:-console}"
LOG_LEVEL="${LOG_LEVEL:-debug}"

# Start tmux session detached with two panes
TMUX_SERVER_CMD="./cmd/apps/mcp-oidc-server/mcp-oidc-server --addr $ADDR --issuer $ISSUER --log-format $LOG_FORMAT --log-level $LOG_LEVEL"

tmux new-session -d -s "$SESSION" -n server "$TMUX_SERVER_CMD"

tmux split-window -h -t "$SESSION:server" \
  'sleep 1 && echo "=== Discovery ===" && curl -s http://localhost:8080/.well-known/openid-configuration | jq . && \
   echo "\n=== AS Metadata ===" && curl -s http://localhost:8080/.well-known/oauth-authorization-server | jq . && \
   echo "\n=== JWKS ===" && curl -s http://localhost:8080/jwks.json | jq . && \
   echo "\n=== MCP (should be 501) ===" && curl -i -s -X POST http://localhost:8080/mcp -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\"}"'

tmux select-pane -t "$SESSION:server.0"

if [[ "$MODE" == "attach" ]]; then
  tmux attach -t "$SESSION"
  exit 0
fi

# Headless: wait a bit for curls to complete, then capture all panes and kill session
echo "Letting curls run for a moment..." >&2
sleep 3

ts=$(date +%Y%m%d-%H%M%S)
base=./cmd/apps/mcp-oidc-server
srv_log="$base/out-server-$ts.log"
curl_log="$base/out-curl-$ts.log"

# Capture all panes of the window, place larger output into server log heuristically
pane_info=$(tmux list-panes -t "$SESSION:server" -F '#{pane_index} #{pane_id}')
srv_captured=false
curl_captured=false
while read -r idx pid; do
  tmpfile=$(mktemp)
  tmux capture-pane -p -t "$pid" > "$tmpfile" || true
  if [[ "$srv_captured" == false ]]; then
    cp "$tmpfile" "$srv_log"
    srv_captured=true
  else
    cp "$tmpfile" "$curl_log"
    curl_captured=true
  fi
  rm -f "$tmpfile"
done <<< "$pane_info"

tmux kill-session -t "$SESSION" || true

echo "Saved logs:" >&2
echo "  server: $srv_log" >&2
echo "  curls : $curl_log" >&2

# Print a brief summary to stdout for quick inspection
echo "--- CURL OUTPUT (tail) ---"
tail -n 100 "$curl_log" 2>/dev/null || true
echo "--- SERVER OUTPUT (tail) ---"
tail -n 100 "$srv_log" 2>/dev/null || true


