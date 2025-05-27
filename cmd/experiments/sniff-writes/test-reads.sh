#!/bin/bash

echo "Testing read operations..."

# Create a test file
echo "Hello, World!" > /tmp/test-read-file.txt

echo ""
echo "=== Test 1: Default behavior (should NOT capture reads) ==="
echo "Running: ./sniff-writes monitor -d /tmp -t 3s --glob 'test-read-*'"
sudo timeout 3s ./sniff-writes monitor -d /tmp -t 3s --glob 'test-read-*' &
MONITOR_PID=$!

sleep 1
echo "Reading from test file..."
cat /tmp/test-read-file.txt > /dev/null
sleep 2

wait $MONITOR_PID
echo ""

echo "=== Test 2: With reads enabled (should capture reads) ==="
echo "Running: ./sniff-writes monitor -d /tmp -t 3s --glob 'test-read-*' -o 'open,read,write,close'"
sudo timeout 3s ./sniff-writes monitor -d /tmp -t 3s --glob 'test-read-*' -o 'open,read,write,close' &
MONITOR_PID=$!

sleep 1
echo "Reading from test file..."
cat /tmp/test-read-file.txt > /dev/null
sleep 2

wait $MONITOR_PID

# Cleanup
rm -f /tmp/test-read-file.txt
echo ""
echo "Test complete."