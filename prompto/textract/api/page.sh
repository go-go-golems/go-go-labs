#!/bin/bash

echo "Page Interface Key Concepts:"
echo "- Each page contains child blocks for detected items"
echo "- Can contain: lines, tables, forms, key-value pairs, queries"
echo "- Has geometry information (bounding box)"
echo "- Returns items in implied reading order"
echo ""
echo "Relevant Documentation Files:"

pinocchio catter print "cmd/apps/textractor/ttmp/2024-11-29/textract-docs/03-pages.txt" \
                      "cmd/apps/textractor/ttmp/2024-11-29/textract-api-docs/api-Block.txt" \
                      "cmd/apps/textractor/ttmp/2024-11-29/textract-docs/09-layout-response.txt" 