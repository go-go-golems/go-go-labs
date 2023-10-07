#!/bin/bash

# Check if at least one file is provided
if [ "$#" -eq 0 ]; then
    echo "Please provide at least one file as an argument."
    exit 1
fi

# Loop through all provided files
for file in "$@"; do
    # Check if the file exists
    if [ ! -f "$file" ]; then
        echo "File $file not found!"
        continue
    fi

    # Use sed to perform the replacement
    sed -i 's/\*\*user\*\*/\*\*john\*\*/g' "$file"
    sed -i 's/\*\*assistant\*\*/\*\*billy\*\*/g' "$file"
done

echo "Replacement done for all provided files."
