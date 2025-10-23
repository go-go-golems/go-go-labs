#!/bin/bash
# Production build script

set -e

echo "Building PR History & Code Browser..."

# Build frontend
echo "Building frontend..."
cd frontend
npm install
npm run build
cd ..

# Build Go binary
echo "Building Go binary..."
go build -o pr-history-code-browser main.go

echo ""
echo "Build complete!"
echo "Binary: ./pr-history-code-browser"
echo ""
echo "Run with:"
echo "  ./pr-history-code-browser --db /path/to/database.db --port 8080"

