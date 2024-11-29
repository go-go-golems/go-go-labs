#!/bin/bash

# Make all scripts executable
chmod +x *.sh

echo "=== Document Interface ==="
./document.sh

echo "=== Page Interface ==="
./page.sh

echo "=== Block Interface ==="
./block.sh

echo "=== Line Interface ==="
./line.sh

echo "=== Table Interface ==="
./table.sh

echo "=== Form Interface ==="
./form.sh

echo "=== KeyValue Interface ==="
./keyvalue.sh

echo "=== SelectionElement Interface ==="
./selection.sh

echo "=== Query Interface ==="
./query.sh

echo "=== Geometry Types ==="
./geometry.sh

echo "=== Support Types ==="
./support.sh 