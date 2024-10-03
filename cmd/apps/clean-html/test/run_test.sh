#!/bin/bash

# Run the HTML extractor
cat test/sample.html | python extract_html.py --config test/config.yaml > test/output.yaml

# Display the output
echo "Extraction complete. Output:"
cat test/output.yaml