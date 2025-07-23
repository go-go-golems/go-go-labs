#!/bin/bash
# Setup script for testing Redis monitor with real data

echo "🚀 Setting up Redis Monitor Test Environment"
echo "============================================"

# Check if Redis is running
if ! redis-cli ping > /dev/null 2>&1; then
    echo "❌ Redis is not running!"
    echo "Please start Redis first:"
    echo "  redis-server --daemonize yes"
    echo "  # or"
    echo "  redis-server"
    exit 1
fi

echo "✅ Redis is running"

# Clean up any existing test data
echo "🧹 Cleaning up old test data..."
redis-cli FLUSHALL > /dev/null

# Create initial test streams with some baseline data
echo "📊 Creating test streams..."

redis-cli XADD orders "*" user_id 100 product laptop price 1200 > /dev/null
redis-cli XADD orders "*" user_id 101 product mouse price 25 > /dev/null
redis-cli XADD orders "*" user_id 102 product keyboard price 75 > /dev/null

redis-cli XADD events "*" type login user_id 100 > /dev/null
redis-cli XADD events "*" type logout user_id 99 > /dev/null

redis-cli XADD logs "*" level info message "Server started" > /dev/null
redis-cli XADD logs "*" level debug message "Connection established" > /dev/null

# Create consumer groups for testing
echo "👥 Creating consumer groups..."
redis-cli XGROUP CREATE orders order-processors 0 MKSTREAM > /dev/null 2>&1
redis-cli XGROUP CREATE events analytics 0 MKSTREAM > /dev/null 2>&1
redis-cli XGROUP CREATE logs log-processors 0 MKSTREAM > /dev/null 2>&1

# Show current state
echo ""
echo "📈 Current Redis streams:"
echo "  orders: $(redis-cli XLEN orders) entries"
echo "  events: $(redis-cli XLEN events) entries" 
echo "  logs: $(redis-cli XLEN logs) entries"

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "1. Run: ./start-monitor.sh    (to start the TUI)"
echo "2. Run: ./generate-data.sh    (to add real-time data)"
echo "3. Watch the sparklines update in real-time!"
