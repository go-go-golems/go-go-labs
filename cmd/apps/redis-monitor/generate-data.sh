#!/bin/bash
# Generate real-time data for testing sparklines

echo "📈 Redis Real-time Data Generator"
echo "================================="

# Check if Redis is running
if ! redis-cli ping > /dev/null 2>&1; then
    echo "❌ Redis is not running!"
    exit 1
fi

echo "🔄 Starting data generation..."
echo "Press Ctrl+C to stop"
echo ""

# Function to generate random data
generate_burst() {
    local stream=$1
    local count=$2
    local delay=$3
    
    echo "⚡ Burst: Adding $count messages to $stream (delay: ${delay}s)"
    for i in $(seq 1 $count); do
        case $stream in
            "orders")
                redis-cli XADD orders "*" \
                    user_id $((RANDOM % 1000)) \
                    product "item_$((RANDOM % 50))" \
                    price $((100 + RANDOM % 1900)) \
                    timestamp $(date +%s) > /dev/null
                ;;
            "events")
                redis-cli XADD events "*" \
                    type "event_$((RANDOM % 10))" \
                    user_id $((RANDOM % 1000)) \
                    timestamp $(date +%s) > /dev/null
                ;;
            "logs")
                redis-cli XADD logs "*" \
                    level "info" \
                    message "Generated log entry $i" \
                    timestamp $(date +%s) > /dev/null
                ;;
        esac
        sleep $delay
    done
}

# Main data generation loop
counter=0
while true; do
    counter=$((counter + 1))
    
    case $((counter % 4)) in
        0)
            # Heavy orders burst
            generate_burst "orders" 5 0.5
            ;;
        1)
            # Light events
            generate_burst "events" 2 1
            ;;
        2)
            # Medium logs
            generate_burst "logs" 3 0.8
            ;;
        3)
            # Mixed activity
            redis-cli XADD orders "*" user_id $((RANDOM % 1000)) product "special_item" price 999 > /dev/null
            sleep 1
            redis-cli XADD events "*" type "special_event" user_id 999 > /dev/null
            sleep 1
            redis-cli XADD logs "*" level "warn" message "Special activity detected" > /dev/null
            ;;
    esac
    
    # Show current counts
    echo "📊 Current: orders=$(redis-cli XLEN orders), events=$(redis-cli XLEN events), logs=$(redis-cli XLEN logs)"
    
    # Random pause between cycles
    sleep $((2 + RANDOM % 3))
done
