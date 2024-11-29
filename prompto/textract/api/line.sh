#!/bin/bash

echo "Line Interface Key Concepts:"
echo "- String of tab-delimited contiguous words"
echo "- Contains WORD blocks as children"
echo "- Has confidence scores"
echo "- Has geometry information"
echo ""
echo "Relevant Documentation Files:"

pinocchio catter print "cmd/apps/textractor/ttmp/2024-11-29/textract-docs/04-lines-words.txt" \
                      "cmd/apps/textractor/ttmp/2024-11-29/textract-api-docs/api-Block.txt"