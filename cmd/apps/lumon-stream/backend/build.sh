#!/bin/bash

# This script builds the React frontend and embeds it in the Golang binary

# Build the React frontend
cd ../frontend
npm run build

# Copy the build files to the backend embed directory
cp -r build/* ../backend/embed/

# Build the Golang binary with embedded files
cd ../backend
go build -tags embed -o lumonstream

echo "Build completed successfully!"
echo "The binary is located at: $(pwd)/lumonstream"
