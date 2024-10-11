#!/bin/bash

# Check if a directory argument is provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 <input_directory>"
    exit 1
fi

input_dir="$1"
output_dir="${input_dir}/extracted"

# Create the output directory if it doesn't exist
mkdir -p "$output_dir"

# Process all jpg and png files in the input directory
for file in "$input_dir"/*.{jpg,jpeg,png}; do
    # Check if the file exists (to handle cases where no matches are found)
    [ -e "$file" ] || continue
    
    # Get the filename without the path
    filename=$(basename "$file")
    
    # Set the output file path
    output_file="${output_dir}/${filename%.*}_dewarped.${filename##*.}"
    
    echo "Processing: $file"
    python photo-dewarp.py -m morphology "$file" -o "$output_file"
done

echo "All images processed. Dewarped images saved in ${output_dir}"

