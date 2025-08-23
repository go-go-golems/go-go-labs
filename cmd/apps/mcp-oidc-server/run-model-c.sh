#!/usr/bin/env bash
set -euo pipefail

# Run Model C (local users) in tmux, with a fresh DB and sample users
SESSION="mcp-oidc-model-c"
MODE="capture" # pass --attach to attach instead
if [[ "${1:-}" == "--attach" ]]; then MODE="attach"; fi

cd "$(dirname "$0")/../../.."

echo "Building mcp-oidc-server..." >&2
go build -o ./cmd/apps/mcp-oidc-server/mcp-oidc-server ./cmd/apps/mcp-oidc-server

if tmux has-session -t "$SESSION" 2>/dev/null; then
  tmux kill-session -t "$SESSION" || true
fi

# Config
ADDR="${ADDR:-:8080}"
ISSUER="${ISSUER:-http://localhost:8080}"
LOG_FORMAT="${LOG_FORMAT:-console}"
LOG_LEVEL="${LOG_LEVEL:-debug}"
BASE=./cmd/apps/mcp-oidc-server
DB="$BASE/modelc.db"

# Fresh DB
rm -f "$DB"

# Create a few users (Model C)
./cmd/apps/mcp-oidc-server/mcp-oidc-server users add --db "$DB" --username admin --password admin123 --email admin@example.com
./cmd/apps/mcp-oidc-server/mcp-oidc-server users add --db "$DB" --username alice --password alice123 --email alice@example.com
./cmd/apps/mcp-oidc-server/mcp-oidc-server users add --db "$DB" --username bob --password bob123 --email bob@example.com

# Start tmux session with server
TMUX_SERVER_CMD="./cmd/apps/mcp-oidc-server/mcp-oidc-server --addr $ADDR --issuer $ISSUER --log-format $LOG_FORMAT --log-level $LOG_LEVEL --db $DB --local-users --session-ttl 12h --dev-token-fallback=false"
tmux new-session -d -s "$SESSION" -n server "$TMUX_SERVER_CMD"

# Second pane: helpful curls
tmux split-window -h -t "$SESSION:server" \
  "sleep 1 && echo '=== Discovery ===' && curl -s $ISSUER/.well-known/openid-configuration | jq . && \
   echo '\n=== AS Metadata ===' && curl -s $ISSUER/.well-known/oauth-authorization-server | jq . && \
   echo '\n=== JWKS ===' && curl -s $ISSUER/jwks.json | jq . && \
   echo '\nOpen $ISSUER/login in your browser. Try: admin/admin123' && \
   echo 'After login via /oauth2/auth, exchange code at /oauth2/token and call /mcp.'"

tmux select-pane -t "$SESSION:server.0"

if [[ "$MODE" == "attach" ]]; then
  tmux attach -t "$SESSION"
  exit 0
fi

echo "Started tmux session: $SESSION" >&2
echo "Server: $ISSUER (ADDR=$ADDR)" >&2
echo "DB: $DB" >&2


