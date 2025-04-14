#!/bin/bash
set -e

echo "Generating templ templates..."
go generate ./...

echo "Building application..."
go build -o ocr-mistral-view .

echo "Running application with sample.json..."
./ocr-mistral-view -i sample.json 