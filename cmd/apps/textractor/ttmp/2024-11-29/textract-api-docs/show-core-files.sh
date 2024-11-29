#!/bin/bash

# Core files needed for Textract parsing
FILES=(
  "api-Block.txt"
  "api-Document.txt"
  "api-DocumentMetadata.txt"
  "api-Geometry.txt"
  "api-BoundingBox.txt"
  "api-Point.txt"
  "api-Relationship.txt"
)

# Get the directory of this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Build the full paths
FULL_PATHS=()
for file in "${FILES[@]}"; do
  FULL_PATHS+=("$DIR/$file")
done

# Print header
echo "=== Core Textract API Documentation Files ==="
echo

# Execute pinocchio catter with all files
pinocchio catter print "${FULL_PATHS[@]}" 