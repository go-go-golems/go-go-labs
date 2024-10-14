#!/bin/bash

# Default values
num_files=5
num_lines=10
output_format="yaml"

# Function to display usage information
usage() {
    echo "Usage: $0 [-f num_files] [-l num_lines] [-t] [-h]"
    echo "  -f num_files  Number of files to process (default: 5)"
    echo "  -l num_lines  Number of lines to display per file (default: 10)"
    echo "  -t            Output in text format (default is YAML)"
    echo "  -h            Display this help message"
    exit 1
}

# Parse command line arguments
while getopts "f:l:th" opt; do
  case $opt in
    f) num_files=$OPTARG ;;
    l) num_lines=$OPTARG ;;
    t) output_format="text" ;;
    h) usage ;;
    \?) echo "Invalid option -$OPTARG" >&2; usage ;;
  esac
done

# Define patterns to search for
patterns=("*.log" "*.RnaSeqMetrics.txt" "*.flagstat.concord.txt" "*.star.duplic" "*.p2.Log.final.out")

# Output YAML header
if [ "$output_format" = "yaml" ]; then
    echo "---"
    echo "log_files:"
fi

# Loop through each pattern
for pattern in "${patterns[@]}"; do
    if [ "$output_format" = "text" ]; then
        echo "Processing pattern: $pattern"
    elif [ "$output_format" = "yaml" ]; then
        echo "- pattern: $pattern"
        echo "  files:"
    fi
    
    # Find the first N files matching the pattern
    files=$(find . -type f -name "$pattern" | head -n $num_files)
    
    # Loop through each found file
    for file in $files; do
        if [ "$output_format" = "text" ]; then
            echo "File: $file"
            # Display the top N lines of the file
            head -n $num_lines "$file"
            echo "-----------------------------"
        else
            content=$(head -n $num_lines "$file")
            echo "  - name: $file"
            echo "    content: |"
            echo "$content" | sed 's/^/      /'
        fi
    done
done

# Output YAML footer
if [ "$output_format" = "yaml" ]; then
    echo "..."
fi
