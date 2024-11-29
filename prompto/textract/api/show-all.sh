#!/bin/bash

# Make all scripts executable
chmod +x *.sh

echo "=== Document Interface ==="
prompto get textract/api/document.sh

echo "=== Page Interface ==="
prompto get textract/api/page.sh

echo "=== Block Interface ==="
prompto get textract/api/block.sh

echo "=== Line Interface ==="
prompto get textract/api/line.sh

echo "=== Table Interface ==="
prompto get textract/api/table.sh

echo "=== Form Interface ==="
prompto get textract/api/form.sh

echo "=== KeyValue Interface ==="
prompto get textract/api/keyvalue.sh

echo "=== SelectionElement Interface ==="
prompto get textract/api/selection.sh

echo "=== Query Interface ==="
prompto get textract/api/query.sh

echo "=== Geometry Types ==="
prompto get textract/api/geometry.sh

echo "=== Support Types ==="
prompto get textract/api/support.sh