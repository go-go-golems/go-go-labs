#!/bin/bash

# Script to render still frames every 30 frames for each animation in Root.tsx
# Usage: ./render-stills.sh

set -e

# Configuration
ENTRY_POINT="src/index.ts"
OUTPUT_BASE="/tmp/remotion-stills"
FRAME_INTERVAL=30

# Create base output directory
mkdir -p "$OUTPUT_BASE"

# Extract entry point name for directory structure
ENTRY_NAME=$(basename "$ENTRY_POINT" .ts)
OUTPUT_DIR="$OUTPUT_BASE/$ENTRY_NAME"
mkdir -p "$OUTPUT_DIR"

echo "🎬 Starting still frame rendering..."
echo "📁 Output directory: $OUTPUT_DIR"
echo "⚡ Frame interval: every $FRAME_INTERVAL frames"
echo ""

# Composition data extracted from Root.tsx
declare -A compositions=(
    ["ToolCallingAnimationNew"]=780
    ["CRMQueryAnimationNew"]=630
    ["SQLiteQueryAnimationNew"]=1440
    ["SQLiteViewOptimizationAnimationNew"]=1280
    ["ContextBuildupAnimation"]=300
    ["PostResponseEditingAnimation"]=400
    ["AdaptiveSystemPromptAnimation"]=630
    ["AssistantDiscussionAnimation"]=360
    ["UserControlledToolsAnimation"]=480
    ["LLMGeneratedUIAnimation"]=460
)

# Function to render stills for a composition
render_composition_stills() {
    local comp_id="$1"
    local duration="$2"
    
    echo "🎯 Processing: $comp_id (${duration} frames)"
    
    # Create composition-specific directory
    local comp_dir="$OUTPUT_DIR/$comp_id"
    mkdir -p "$comp_dir"
    
    # Calculate frames to render (every 30 frames)
    local frame=0
    local frame_count=0
    
    while [ $frame -lt $duration ]; do
        local frame_padded=$(printf "%03d" $frame_count)
        local output_file="$comp_dir/$comp_id-$frame_padded.png"
        
        echo "  📸 Rendering frame $frame → $output_file"
        
        # Render the still frame
        npx remotion still "$ENTRY_POINT" "$comp_id" "$output_file" --frame="$frame"
        
        if [ $? -eq 0 ]; then
            echo "  ✅ Success: $output_file"
        else
            echo "  ❌ Failed: $output_file"
        fi
        
        frame=$((frame + FRAME_INTERVAL))
        frame_count=$((frame_count + 1))
    done
    
    echo "  🏁 Completed $comp_id: $frame_count stills rendered"
    echo ""
}

# Render stills for each composition
total_compositions=${#compositions[@]}
current_comp=0

for comp_id in "${!compositions[@]}"; do
    current_comp=$((current_comp + 1))
    duration=${compositions[$comp_id]}
    
    echo "[$current_comp/$total_compositions] Starting $comp_id"
    render_composition_stills "$comp_id" "$duration"
done

echo "🎉 All compositions processed!"
echo "📁 Stills saved to: $OUTPUT_DIR"
echo ""
echo "📊 Summary:"
for comp_id in "${!compositions[@]}"; do
    duration=${compositions[$comp_id]}
    still_count=$(( (duration + FRAME_INTERVAL - 1) / FRAME_INTERVAL ))
    echo "  $comp_id: $still_count stills"
done 