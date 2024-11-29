#!/bin/bash

echo "Document Interface Key Concepts:"
echo "- Document is made up of Block objects"
echo "- Contains list of child IDs for lines of text, key-value pairs, tables, queries"
echo "- Metadata includes number of pages"
echo "- Document can be processed sync or async"
echo ""
echo "Relevant Documentation Files:"

pinocchio catter print "cmd/apps/textractor/ttmp/2024-11-29/textract-docs/02-text-detection.txt" \
                      "cmd/apps/textractor/ttmp/2024-11-29/textract-api-docs/api-DocumentMetadata.txt" \
                      "cmd/apps/textractor/ttmp/2024-11-29/textract-api-docs/api-Block.txt" 