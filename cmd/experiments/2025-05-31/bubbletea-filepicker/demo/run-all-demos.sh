#!/bin/bash

# Script to build the app, setup test environment, and run all VHS demos

set -e

echo "🎬 Setting up Bubbletea File Manager demos..."

# Navigate to the project directory
cd "$(dirname "$0")/.."

# Build the application
echo "🔨 Building application..."
go build -o filepicker .

# Setup test environment
echo "📁 Setting up test files..."
./demo/setup-test-env.sh

# Run all VHS demos
echo "🎥 Recording demos..."

demos=(
    "basic-navigation"
    "file-icons" 
    "multi-selection"
    "copy-paste"
    "cut-paste"
    "delete-confirm"
    "create-files"
    "rename-file"
    "help-system"
    "overview"
)

for demo in "${demos[@]}"; do
    if [ -f "demo/${demo}.tape" ]; then
        echo "Recording ${demo}..."
        vhs demo/${demo}.tape
    else
        echo "⚠️  Script demo/${demo}.tape not found"
    fi
done

echo "✅ All demos recorded! GIFs are in the demo/ directory."
echo ""
echo "Generated files:"
ls -la demo/*.gif 2>/dev/null || echo "No GIF files found - make sure VHS is installed"
