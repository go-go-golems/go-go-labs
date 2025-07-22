#!/bin/bash
# Start the Redis monitor TUI

echo "🖥️  Starting Redis Monitor TUI"
echo "=============================="

# Build the application if needed
if [ ! -f "./redis-monitor" ]; then
    echo "🔨 Building redis-monitor..."
    go build .
    if [ $? -ne 0 ]; then
        echo "❌ Build failed!"
        exit 1
    fi
fi

echo "✅ Starting TUI with 1-second refresh rate"
echo ""
echo "📊 What to watch for:"
echo "  - Sparklines start at all zeros: ▁▁▁▁▁▁▁▁▁▁"
echo "  - As data is added, they show message rates"
echo "  - New data slides in from right, old data shifts left"
echo ""
echo "🎮 Controls:"
echo "  - r: Manual refresh"
echo "  - g: Switch to groups view" 
echo "  - s: Switch to streams view"
echo "  - q: Quit"
echo ""
echo "Press any key to start..."
read -n 1

./redis-monitor tui --refresh-rate 1s
