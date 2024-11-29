#!/bin/bash

echo "Query Interface Key Concepts:"
echo "- Contains question and alias"
echo "- Has relationship to answers"
echo "- Contains confidence scores"
echo "- Can specify page ranges"
echo ""
echo "Relevant Documentation Files:"

pinocchio catter print "cmd/apps/textractor/ttmp/2024-11-29/textract-docs/08-queries.txt" \
                      "cmd/apps/textractor/ttmp/2024-11-29/textract-api-docs/api-Query.txt" \
                      "cmd/apps/textractor/ttmp/2024-11-29/textract-api-docs/api-Block.txt" 