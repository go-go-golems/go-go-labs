#!/bin/bash

# Create test directory structure for VHS demos
TEST_DIR="test-files"

# Clean up any existing test directory
rm -rf "$TEST_DIR"

# Create main test directory
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Create subdirectories
mkdir -p documents/{personal,work}
mkdir -p projects/{go,python,javascript}
mkdir -p media/{images,videos,audio}
mkdir -p backups

# Create sample files in documents
echo "This is my personal README file" > documents/personal/readme.md
echo "# Personal Notes" > documents/personal/notes.md
echo "Shopping list: milk, bread, eggs" > documents/personal/shopping.txt

echo "Quarterly report for Q4 2024" > documents/work/report.txt
echo "Meeting minutes from team standup" > documents/work/meeting-notes.md
echo "Budget spreadsheet data" > documents/work/budget.csv

# Create sample files in projects
echo "package main

import \"fmt\"

func main() {
    fmt.Println(\"Hello, World!\")
}" > projects/go/main.go

echo "# Go Project README" > projects/go/README.md

echo "def hello_world():
    print(\"Hello, World!\")

if __name__ == \"__main__\":
    hello_world()" > projects/python/hello.py

echo "# Python Project" > projects/python/README.md
echo "numpy==1.21.0
requests==2.25.1" > projects/python/requirements.txt

echo "console.log('Hello, World!');" > projects/javascript/app.js
echo "{
  \"name\": \"demo-app\",
  \"version\": \"1.0.0\",
  \"main\": \"app.js\"
}" > projects/javascript/package.json

# Create some dummy media files (just text files with media extensions)
echo "dummy image data" > media/images/photo1.jpg
echo "dummy image data" > media/images/photo2.png
echo "dummy image data" > media/images/diagram.svg

echo "dummy video data" > media/videos/movie.mp4
echo "dummy video data" > media/videos/presentation.avi

echo "dummy audio data" > media/audio/song.mp3
echo "dummy audio data" > media/audio/podcast.wav

# Create some archive files (dummy)
echo "dummy zip data" > backups/backup-2024.zip
echo "dummy tar data" > backups/logs.tar.gz
echo "dummy archive data" > backups/old-files.rar

# Create some executable scripts
echo "#!/bin/bash
echo 'This is a test script'" > test-script.sh
chmod +x test-script.sh

# Create some configuration files
echo "[settings]
debug=true
port=8080" > config.ini

echo "version: '3.8'
services:
  app:
    build: .
    ports:
      - '8080:8080'" > docker-compose.yml

echo "Test environment created successfully!"
echo "Directory structure:"
find . -type f | head -20
echo "..."
echo "Total files created: $(find . -type f | wc -l)"
