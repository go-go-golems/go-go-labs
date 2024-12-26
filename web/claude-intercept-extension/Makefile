# Makefile for Chrome Extension

# Name of the output zip file
ZIP_NAME = claude_conversation_downloader.zip

# List of files to include in the zip
FILES = manifest.json \
        background.js \
        popup.html \
        popup.js \
        README.md \
        jszip.min.js

# Default target
all: zip

# Create the zip file
zip:
	@echo "Creating $(ZIP_NAME)..."
	@zip -r $(ZIP_NAME) $(FILES)
	@echo "$(ZIP_NAME) created successfully."

# Clean up (remove the zip file)
clean:
	@echo "Removing $(ZIP_NAME)..."
	@rm -f $(ZIP_NAME)
	@echo "$(ZIP_NAME) removed."

# Phony targets
.PHONY: all zip clean