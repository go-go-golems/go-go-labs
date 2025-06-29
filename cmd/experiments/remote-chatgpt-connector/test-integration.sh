#!/bin/bash

echo "🚀 MCP Remote Connector Integration Test"
echo "======================================"

BASE_URL="http://localhost:8080"

echo ""
echo "1. Testing Health Check..."
curl -s "$BASE_URL/health" | jq .

echo ""
echo "2. Testing Plugin Manifest (Discovery)..."
curl -s -H "User-Agent: ChatGPT-User-Agent" "$BASE_URL/.well-known/ai-plugin.json" | jq '{name, auth, api}'

echo ""
echo "3. Testing OAuth Configuration..."
curl -s "$BASE_URL/.well-known/oauth-authorization-server" | jq '{issuer, authorization_endpoint, token_endpoint, scopes_supported}'

echo ""
echo "4. Testing Unauthorized SSE Access..."
curl -s -w "Status: %{http_code}\n" "$BASE_URL/sse" -o /dev/null

echo ""
echo "5. Testing Invalid Token..."
curl -s -w "Status: %{http_code}\n" -H "Authorization: Bearer invalid_token" "$BASE_URL/sse" -o /dev/null

echo ""
echo "📋 Next Steps for ChatGPT Integration:"
echo "1. Make sure your server is accessible via HTTPS (use Tailscale Funnel or ngrok)"
echo "2. Update your GitHub OAuth app callback URL to ChatGPT's redirect"
echo "3. In ChatGPT, go to Settings → Data Controls → Connectors"
echo "4. Add connector with your public HTTPS URL"
echo "5. Complete the GitHub OAuth flow"
echo ""
echo "🔍 Server is running in tmux session 'mcp-server'"
echo "   View logs: tmux attach -t mcp-server"
echo "   Stop server: tmux send-keys -t mcp-server C-c"
