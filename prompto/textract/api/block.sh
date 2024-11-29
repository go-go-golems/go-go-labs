#!/bin/bash

echo "Block Interface Key Concepts:"
echo "- Basic unit of all detected items"
echo "- Has unique identifier"
echo "- Contains confidence scores"
echo "- Has relationships (parent/child)"
echo "- Contains geometry information"
echo ""
echo "Relevant Documentation Files:"

pinocchio catter print "cmd/apps/textractor/ttmp/2024-11-29/textract-docs/02-text-detection.txt" \
                      "cmd/apps/textractor/ttmp/2024-11-29/textract-api-docs/api-Block.txt" \
                      "cmd/apps/textractor/ttmp/2024-11-29/textract-api-docs/api-Relationship.txt" 