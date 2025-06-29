#!/bin/bash
set -e

echo "🎬 Testing Add Research VHS Demos"
echo "=================================="

# Ensure we're in the right directory
cd "$(dirname "$0")/.."

# Check prerequisites
echo "🔍 Checking prerequisites..."

if ! command -v vhs &> /dev/null; then
    echo "❌ VHS not found. Install with: go install github.com/charmbracelet/vhs@latest"
    exit 1
fi

if [ ! -f "add-research" ]; then
    echo "🔨 Building add-research tool..."
    go build -o add-research .
fi

echo "✅ Prerequisites met"

# Test each demo tape file
demos=(
    "demo-basic"
    "demo-files" 
    "demo-content"
    "demo-search"
    "demo-types"
)

echo ""
echo "🎭 Running VHS demos..."

for demo in "${demos[@]}"; do
    echo "▶️  Testing $demo..."
    
    # Check if tape file exists
    if [ ! -f "demos/$demo.tape" ]; then
        echo "❌ demos/$demo.tape not found"
        continue
    fi
    
    # Run VHS (this will create the GIF)
    if vhs "demos/$demo.tape" 2>/dev/null; then
        echo "✅ $demo.gif created successfully"
        
        # Check file size
        if [ -f "demos/$demo.gif" ]; then
            size=$(du -h "demos/$demo.gif" | cut -f1)
            echo "   📊 Size: $size"
        fi
        
        # Check if screenshot was created
        if [ -f "demos/$demo.txt" ]; then
            echo "   📸 Screenshot available: demos/$demo.txt"
        fi
    else
        echo "❌ Failed to create $demo.gif"
    fi
    
    echo ""
done

echo "📈 Demo Statistics:"
echo "=================="

total_size=0
gif_count=0

for demo in "${demos[@]}"; do
    if [ -f "demos/$demo.gif" ]; then
        size_bytes=$(stat -c%s "demos/$demo.gif" 2>/dev/null || stat -f%z "demos/$demo.gif" 2>/dev/null || echo "0")
        size_human=$(du -h "demos/$demo.gif" | cut -f1)
        echo "$demo.gif: $size_human"
        total_size=$((total_size + size_bytes))
        gif_count=$((gif_count + 1))
    fi
done

if [ $gif_count -gt 0 ]; then
    total_human=$(echo $total_size | awk '
        function human(x) {
            if (x<1000) {return x " B"}
            x/=1024
            if (x<1000) {return int(x) " KB"}
            x/=1024
            if (x<1000) {return int(x) " MB"}
            x/=1024
            return int(x) " GB"
        }
        {print human($1)}
    ')
    echo ""
    echo "Total: $gif_count demos, $total_human"
fi

echo ""
echo "🎯 Next Steps:"
echo "- Review GIF files in demos/ directory"  
echo "- Check TXT screenshots for validation"
echo "- Embed GIFs in documentation"
echo "- Test demos work on different environments"

echo ""
echo "✅ Demo testing completed!"
