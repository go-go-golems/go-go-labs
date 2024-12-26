# Makefile for Cross-Browser Extension

# Default target
all: chrome firefox

# Create Chrome extension
chrome:
	mkdir -p dist/chrome
	cp manifest.json dist/chrome/
	cp *.js dist/chrome/
	cp *.html dist/chrome/
	cd dist && zip -r chrome.zip chrome/

# Create Firefox extension
firefox: clean-firefox
	mkdir -p dist/firefox
	cp manifest.json dist/firefox/
	cp *.js dist/firefox/
	cp *.html dist/firefox/
	cd dist/firefox && zip -r ../firefox.xpi .

# Clean targets
clean: clean-chrome clean-firefox
	rm -rf dist

clean-chrome:
	rm -rf dist/chrome
	rm -f dist/chrome.zip

clean-firefox:
	rm -rf dist/firefox
	rm -f dist/firefox.xpi

# Phony targets
.PHONY: all chrome firefox clean clean-chrome clean-firefox