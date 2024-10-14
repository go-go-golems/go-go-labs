#!/bin/bash

# Function to display help message
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo "Play video from /dev/video0 with selected resolution."
    echo
    echo "Options:"
    echo "  -h, --help    Show this help message and exit"
}

# Check for help option
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    show_help
    exit 0
fi

# Check if required commands are available
for cmd in v4l2-ctl gum ffplay; do
    if ! command -v $cmd &> /dev/null; then
        echo "Error: $cmd is not installed or not in PATH"
        exit 1
    fi
done

# Parse resolutions from v4l2-ctl output
resolutions=$(v4l2-ctl --list-formats-ext | grep -oP 'Size: Discrete \K[0-9]+x[0-9]+' | sort -u)

# Use gum to choose a resolution
selected_resolution=$(echo "$resolutions" | gum choose)

# Run ffplay with the selected resolution
if [ -n "$selected_resolution" ]; then
    ffplay -video_size "$selected_resolution" /dev/video0
else
    echo "No resolution selected."
fi
