#!/bin/bash

echo "Geometry Types Key Concepts:"
echo "- BoundingBox uses ratios of page dimensions"
echo "- Points are coordinate pairs"
echo "- Polygon provides fine-grained boundary"
echo "- Coordinates relative to top-left origin"
echo ""
echo "Relevant Documentation Files:"

pinocchio catter print "cmd/apps/textractor/ttmp/2024-11-29/textract-api-docs/api-Geometry.txt" \
                      "cmd/apps/textractor/ttmp/2024-11-29/textract-api-docs/api-BoundingBox.txt" \
                      "cmd/apps/textractor/ttmp/2024-11-29/textract-api-docs/api-Point.txt" 