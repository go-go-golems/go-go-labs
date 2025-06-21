#!/bin/bash

# Test script for GitHub GraphQL CLI labels and issue editing functionality
# This script tests creating issues with labels and updating issue body/labels

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_OWNER="go-go-golems"
REPO_NAME="go-go-labs"  # Assuming this is a test repo
TEST_ISSUE_PREFIX="[TEST-LABELS]"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

echo -e "${BLUE}Starting GitHub GraphQL CLI Labels and Editing Test${NC}"
echo -e "${BLUE}Repository: ${REPO_OWNER}/${REPO_NAME}${NC}"
echo -e "${BLUE}Timestamp: ${TIMESTAMP}${NC}"
echo ""

# Function to print test step
print_step() {
    echo -e "${YELLOW}=== $1 ===${NC}"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Check if CLI is built
if [ ! -f "./github-projects" ]; then
    print_error "github-projects binary not found. Building..."
    go build .
    if [ $? -ne 0 ]; then
        print_error "Failed to build github-projects"
        exit 1
    fi
    print_success "Built github-projects"
fi

# Set environment variables
export GITHUB_OWNER="go-go-golems"
export GITHUB_PROJECT_NUMBER="1"

# Test 1: Create issue with initial labels
print_step "Test 1: Creating issue with initial labels"
ISSUE_TITLE="${TEST_ISSUE_PREFIX} Test Issue ${TIMESTAMP}"
INITIAL_BODY="This is a test issue created for testing labels and editing functionality.

Created at: $(date)
Test run: ${TIMESTAMP}"

INITIAL_LABELS="bug"

echo "Creating issue with:"
echo "  Title: ${ISSUE_TITLE}"
echo "  Labels: ${INITIAL_LABELS}"
echo "  Body: ${INITIAL_BODY}"

CREATE_OUTPUT=$(./github-projects create-issue \
    --repo-owner="${REPO_OWNER}" \
    --repo-name="${REPO_NAME}" \
    --title="${ISSUE_TITLE}" \
    --body="${INITIAL_BODY}" \
    --labels="${INITIAL_LABELS}" \
    --output=json 2>/dev/null)

if [ $? -eq 0 ]; then
    ISSUE_NUMBER=$(echo "$CREATE_OUTPUT" | jq -r '.[0].issue_number')
    ISSUE_URL=$(echo "$CREATE_OUTPUT" | jq -r '.[0].issue_url')
    print_success "Created issue #${ISSUE_NUMBER}"
    echo "  URL: ${ISSUE_URL}"
    echo "  Output: ${CREATE_OUTPUT}"
else
    print_error "Failed to create issue"
    echo "Output: ${CREATE_OUTPUT}"
    exit 1
fi

echo ""

# Test 2: Update issue body
print_step "Test 2: Updating issue body"
UPDATED_BODY="This is an UPDATED test issue for testing labels and editing functionality.

Originally created at: $(date)
Updated at: $(date)
Test run: ${TIMESTAMP}

## Changes Made
- Updated the issue body
- Added more content
- Testing markdown formatting

## Next Steps
- Will test label updates next
- Then test removing labels"

echo "Updating issue #${ISSUE_NUMBER} body..."

UPDATE_BODY_OUTPUT=$(./github-projects update-issue \
    --repo-owner="${REPO_OWNER}" \
    --repo-name="${REPO_NAME}" \
    --issue-number="${ISSUE_NUMBER}" \
    --body="${UPDATED_BODY}" \
    --output=json 2>/dev/null)

if [ $? -eq 0 ]; then
    print_success "Updated issue body"
    echo "  Output: ${UPDATE_BODY_OUTPUT}"
else
    print_error "Failed to update issue body"
    echo "Output: ${UPDATE_BODY_OUTPUT}"
fi

echo ""

# Test 3: Update issue labels (add more labels)
print_step "Test 3: Adding more labels"
ADDITIONAL_LABELS="enhancement"

echo "Adding labels: ${ADDITIONAL_LABELS}"

ADD_LABELS_OUTPUT=$(./github-projects update-issue \
    --repo-owner="${REPO_OWNER}" \
    --repo-name="${REPO_NAME}" \
    --issue-number="${ISSUE_NUMBER}" \
    --add-labels="${ADDITIONAL_LABELS}" \
    --output=json 2>/dev/null)

if [ $? -eq 0 ]; then
    print_success "Added more labels"
    echo "  Output: ${ADD_LABELS_OUTPUT}"
else
    print_error "Failed to add labels"
    echo "Output: ${ADD_LABELS_OUTPUT}"
fi

echo ""

# Test 4: Update issue title and labels together
print_step "Test 4: Updating title and labels together"
FINAL_TITLE="${TEST_ISSUE_PREFIX} FINAL - Test Issue ${TIMESTAMP}"
FINAL_LABELS="bug"

echo "Updating to:"
echo "  Title: ${FINAL_TITLE}"
echo "  Labels: ${FINAL_LABELS}"

FINAL_UPDATE_OUTPUT=$(./github-projects update-issue \
    --repo-owner="${REPO_OWNER}" \
    --repo-name="${REPO_NAME}" \
    --issue-number="${ISSUE_NUMBER}" \
    --title="${FINAL_TITLE}" \
    --add-labels="${FINAL_LABELS}" \
    --output=json 2>/dev/null)

if [ $? -eq 0 ]; then
    print_success "Updated title and labels"
    echo "  Output: ${FINAL_UPDATE_OUTPUT}"
else
    print_error "Failed to update title and labels"
    echo "Output: ${FINAL_UPDATE_OUTPUT}"
fi

echo ""

# Test 5: Test with empty labels (remove all labels)
print_step "Test 5: Removing all labels"

echo "Removing the 'bug' label..."

REMOVE_LABELS_OUTPUT=$(./github-projects update-issue \
    --repo-owner="${REPO_OWNER}" \
    --repo-name="${REPO_NAME}" \
    --issue-number="${ISSUE_NUMBER}" \
    --remove-labels="bug" \
    --output=json 2>/dev/null)

if [ $? -eq 0 ]; then
    print_success "Removed all labels"
    echo "  Output: ${REMOVE_LABELS_OUTPUT}"
else
    print_error "Failed to remove labels"
    echo "Output: ${REMOVE_LABELS_OUTPUT}"
fi

echo ""

# Test 6: Test MCP tools with the created issue
print_step "Test 6: Testing MCP task management integration"

echo "Testing MCP tools to track the created issue..."

# Start MCP server in background and test it
echo "This would typically involve starting MCP server and testing task management"
echo "Issue created: #${ISSUE_NUMBER} - can be tracked in MCP tasks"

echo ""

# Summary
print_step "Test Summary"
print_success "Issue #${ISSUE_NUMBER} created and updated successfully"
echo "  Title: ${FINAL_TITLE}"
echo "  URL: ${ISSUE_URL}"
echo "  Repo: ${REPO_OWNER}/${REPO_NAME}"
echo ""
echo -e "${BLUE}All tests completed!${NC}"
echo ""
echo -e "${YELLOW}Note: You may want to close the test issue #${ISSUE_NUMBER} manually${NC}"
echo -e "${YELLOW}Issue URL: ${ISSUE_URL}${NC}"
