#!/bin/bash

# Redis Streams Tutorial Runner
# Usage: ./run-tutorial.sh [section]

set -e

echo "🚀 Redis Streams Tutorial"
echo "======================="

# Check if redis-cli is available
if ! command -v redis-cli &> /dev/null; then
    echo "❌ redis-cli not found. Please install Redis."
    exit 1
fi

# Check if Redis is running
if ! redis-cli PING &> /dev/null; then
    echo "❌ Redis server not running. Start with: redis-server"
    exit 1
fi

echo "✅ Redis is running"
echo ""

# Clean up any existing data
echo "🧹 Cleaning up existing data..."
redis-cli FLUSHALL > /dev/null

echo ""
echo "📚 Tutorial Sections:"
echo "1. Basic Streams   - 01-basic-streams.md"
echo "2. Consumer Groups  - 02-consumer-groups.md"
echo "3. Pending Messages - 03-pending-messages.md"
echo "4. Real Patterns   - 04-patterns.md"
echo ""
echo "Start with: redis-cli"
echo "Then follow along with the .md files!"