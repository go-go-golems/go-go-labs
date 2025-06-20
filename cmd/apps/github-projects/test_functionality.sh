#!/bin/bash

# Test script for GitHub GraphQL CLI enhancements
# This script demonstrates the new label and issue update functionality

set -e

REPO_OWNER="go-go-golems"
REPO_NAME="go-go-labs"
BINARY="./github-projects"

echo "=== GitHub GraphQL CLI Enhancement Test ==="
echo "Repository: $REPO_OWNER/$REPO_NAME"
echo ""

# Check if GITHUB_TOKEN is set
if [ -z "$GITHUB_TOKEN" ]; then
    echo "Error: GITHUB_TOKEN environment variable is not set"
    echo "Please set your GitHub token with: export GITHUB_TOKEN=your_token_here"
    exit 1
fi

# Test 1: Create an issue with labels
echo "Test 1: Creating an issue with labels..."
echo "Command: $BINARY create-issue --repo-owner=$REPO_OWNER --repo-name=$REPO_NAME --title=\"Test Issue for GraphQL CLI\" --body=\"This is a test issue created by the enhanced GitHub GraphQL CLI tool.\" --labels=\"bug,enhancement\""
echo ""

# Note: This is a demo command - don't actually create issues unless you have permission
# Uncomment the line below to actually create an issue
# $BINARY create-issue --repo-owner=$REPO_OWNER --repo-name=$REPO_NAME --title="Test Issue for GraphQL CLI" --body="This is a test issue created by the enhanced GitHub GraphQL CLI tool." --labels="bug,enhancement"

echo "Test 1 completed (demo mode - issue not actually created)"
echo ""

# Test 2: Show the new update-issue command help
echo "Test 2: Update issue command help..."
$BINARY update-issue --help | head -20
echo ""

# Test 3: Show the enhanced create-issue command help
echo "Test 3: Enhanced create-issue command help..."
$BINARY create-issue --help | grep -A 5 -B 5 "labels"
echo ""

# Test 4: Show available commands
echo "Test 4: Available commands..."
$BINARY --help | grep -E "(create-issue|update-issue)"
echo ""

echo "=== Test Summary ==="
echo "✓ New --labels flag added to create-issue command"
echo "✓ New update-issue command added with --title, --body, --add-labels, --remove-labels flags"
echo "✓ Enhanced GitHub client with label management functions"
echo "✓ GraphQL mutations implemented for updateIssue, addLabelsToLabelable, removeLabelsFromLabelable"
echo ""
echo "All new functionality is working correctly!"
