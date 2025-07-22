#!/bin/bash
# Test script to demonstrate real-time sparkline updates in Redis monitor

echo "🔄 Testing Redis Monitor with Real-time Sparklines"
echo "=================================================="

# Start Redis if not running
if ! redis-cli ping > /dev/null 2>&1; then
    echo "⚠️  Redis not running. Please start Redis first:"
    echo "   redis-server --daemonize yes"
    exit 1
fi

echo "✅ Redis is running"

# Create a clean test stream
redis-cli DEL test-activity > /dev/null 2>&1

echo "📊 Starting Redis Monitor TUI (demo mode for now)..."
echo "   In a real scenario, this would show live message rate sparklines"
echo "   that update as new messages are added to Redis streams."
echo ""
echo "🔧 To test with real data:"
echo "   1. Start: ./redis-monitor tui --refresh-rate 2s"
echo "   2. In another terminal: redis-cli XADD test-stream * data value"
echo "   3. Watch sparklines update in real-time!"
echo ""
echo "Starting demo mode (Press 'q' to quit)..."
sleep 2

./redis-monitor tui --demo --refresh-rate 1s
