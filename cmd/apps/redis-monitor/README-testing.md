# Testing Redis Monitor Real-time Sparklines

This directory contains scripts to test the real-time sparkline functionality of the Redis monitor.

## Quick Test

```bash
# 1. Setup test environment
./setup-test.sh

# 2. Run automated test
./test-sparklines.sh
```

## Manual Testing (Recommended)

### Terminal 1: Start the Monitor
```bash
./start-monitor.sh
```

### Terminal 2: Generate Real-time Data
```bash
./generate-data.sh
```

## What to Look For

### Initial State
- Sparklines show all zeros: `▁▁▁▁▁▁▁▁▁▁`
- This is correct - no recent message activity

### During Data Generation
- Sparklines should show varying heights: `▁▂▅▃▇█▄▂▁▁`
- New activity appears on the right
- Old activity slides left and disappears
- Different streams show different patterns

### Expected Behavior
1. **Message Rate Tracking**: Sparklines show messages added per refresh period (not total count)
2. **Sliding Window**: New data slides in from right, old data shifts left
3. **Zero Baseline**: When no new messages, sparklines return to zeros
4. **Real-time Updates**: Changes visible within 1-2 seconds

## Troubleshooting

### No Activity in Sparklines
- Check if Redis is running: `redis-cli ping`
- Verify data is being added: `redis-cli XLEN orders`
- Ensure refresh rate isn't too slow

### Static Patterns
- If sparklines don't change, the rate calculation might not be working
- Check if timestamps in Redis are updating

### Redis Connection Issues
```bash
# Start Redis if needed
redis-server --daemonize yes

# Or run in foreground
redis-server
```

## File Descriptions

- `setup-test.sh` - Clean Redis and create initial test data
- `start-monitor.sh` - Launch the Redis monitor TUI
- `generate-data.sh` - Continuously add test data to streams
- `test-sparklines.sh` - Automated test with screenshots

## Understanding Sparklines

The sparklines show **message rates**, not cumulative counts:
- Height = number of new messages since last refresh
- Width = last 10 time periods
- Zero = no new activity in that period
- Pattern slides left as time progresses

This gives you a real-time view of Redis stream activity patterns.
