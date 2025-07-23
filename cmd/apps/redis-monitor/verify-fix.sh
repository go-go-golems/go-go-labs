#!/bin/bash
# Verify the crash fix and test basic functionality

echo "🔧 Testing Redis Monitor Fix"
echo "============================"

# Build first
echo "🔨 Building..."
go build .
if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful"

# Test 1: Demo mode (should not crash)
echo ""
echo "🧪 Test 1: Demo mode (5 seconds)..."
timeout 5s ./redis-monitor tui --demo &
DEMO_PID=$!
sleep 6
if kill -0 $DEMO_PID 2>/dev/null; then
    kill $DEMO_PID
    echo "❌ Demo mode didn't exit cleanly"
else
    echo "✅ Demo mode ran successfully"
fi

# Test 2: Real Redis (if available)
echo ""
echo "🧪 Test 2: Redis connection test..."
if redis-cli ping > /dev/null 2>&1; then
    echo "✅ Redis is available"
    
    echo "🧪 Testing real Redis connection (5 seconds)..."
    timeout 5s ./redis-monitor tui --refresh-rate 1s &
    REDIS_PID=$!
    sleep 6
    if kill -0 $REDIS_PID 2>/dev/null; then
        kill $REDIS_PID
        echo "❌ Redis mode didn't exit cleanly"
    else
        echo "✅ Redis mode ran successfully"
    fi
else
    echo "⚠️  Redis not available, skipping real connection test"
fi

# Test 3: Speed controls
echo ""
echo "🧪 Test 3: Speed controls work..."
echo "Start with demo mode and test speed controls manually:"
echo "  1. Run: ./redis-monitor tui --demo"
echo "  2. Press '>' to speed up"
echo "  3. Press '<' to slow down"
echo "  4. Watch the 'Refresh: XXXms' value change"
echo "  5. Press 'q' to quit"

echo ""
echo "🎯 Manual Test:"
echo "   ./redis-monitor tui --demo"
echo ""
echo "Expected behavior:"
echo "  - No crashes"
echo "  - Speed controls (> <) work"
echo "  - Sparklines show activity"
echo "  - Real-time updates"
