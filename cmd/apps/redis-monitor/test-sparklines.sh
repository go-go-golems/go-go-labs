#!/bin/bash
# Simple test to verify sparkline behavior step by step

echo "🧪 Testing Sparkline Behavior"
echo "============================="

# Check if monitor is built
if [ ! -f "./redis-monitor" ]; then
    echo "🔨 Building redis-monitor..."
    go build .
fi

# Check Redis
if ! redis-cli ping > /dev/null 2>&1; then
    echo "❌ Redis is not running!"
    echo "Please start Redis first: redis-server --daemonize yes"
    exit 1
fi

echo "✅ Redis is running"

# Clean start
echo "🧹 Cleaning Redis..."
redis-cli FLUSHALL > /dev/null

# Create test stream
echo "📊 Creating test stream..."
redis-cli XADD test-stream "*" message "initial" > /dev/null

echo ""
echo "🎯 Test Plan:"
echo "1. Start monitor in background"
echo "2. Add messages in controlled bursts"
echo "3. Capture screenshots to show sparkline changes"
echo ""

# Start monitor in background tmux session
echo "🖥️  Starting monitor..."
tmux new-session -d -s test-monitor "cd $(pwd) && ./redis-monitor tui --refresh-rate 2s"
sleep 3

# Capture initial state
echo "📸 Initial state (should show 1 entry, zero rates):"
tmux capture-pane -t test-monitor -p | grep -A 10 "Stream"

echo ""
echo "⏳ Adding burst of 5 messages..."
for i in {1..5}; do
    redis-cli XADD test-stream "*" message "burst_message_$i" timestamp $(date +%s) > /dev/null
    echo "  Added message $i"
    sleep 0.5
done

echo ""
echo "⏸️  Waiting 3 seconds for monitor to refresh..."
sleep 3

echo "📸 After burst (should show increased rates):"
tmux capture-pane -t test-monitor -p | grep -A 10 "Stream"

echo ""
echo "⏳ Adding another burst of 3 messages..."
for i in {6..8}; do
    redis-cli XADD test-stream "*" message "second_burst_$i" timestamp $(date +%s) > /dev/null
    echo "  Added message $i"
    sleep 1
done

echo ""
echo "⏸️  Waiting 3 seconds for monitor to refresh..."
sleep 3

echo "📸 After second burst (sparkline should show new pattern):"
tmux capture-pane -t test-monitor -p | grep -A 10 "Stream"

echo ""
echo "⏸️  Waiting 10 seconds to see rates drop to zero..."
sleep 10

echo "📸 After idle period (rates should return to zero):"
tmux capture-pane -t test-monitor -p | grep -A 10 "Stream"

echo ""
echo "🧪 Test complete!"
echo ""
echo "✅ What you should have seen:"
echo "  1. Initial sparkline: all zeros (▁▁▁▁▁▁▁▁▁▁)"
echo "  2. After first burst: some bars showing activity"
echo "  3. After second burst: different pattern"
echo "  4. After idle: back to zeros"
echo ""
echo "🔍 The sparklines show message rates (new messages per refresh period)"
echo "📊 Pattern should slide from right to left as time progresses"

# Cleanup
tmux kill-session -t test-monitor > /dev/null 2>&1

echo ""
echo "💡 To test manually:"
echo "   Terminal 1: ./start-monitor.sh"
echo "   Terminal 2: ./generate-data.sh"
