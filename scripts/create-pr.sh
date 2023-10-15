#!/bin/bash

set -e

# base branch name is the first argument, defaults to origin/main
base_branch=${1:-origin/main}
if [ "$base_branch" = "origin/main" ]; then
    echo "Using default base branch origin/main"
fi

# File names with fixed names similar to Git's COMMIT_EDITMSG
description_file=".description_editmsg"
issue_file=".issue_editmsg"

# If the description or issue file doesn't exist, get input from user
if [ ! -f "$description_file" ]; then
    gum write --header "Enter description of the pull request (Ctrl-D for exit)" > "$description_file"
fi

if [ ! -f "$issue_file" ]; then
    gum write --header "Enter github issue (Ctrl-D for exit)" --width 80 --header.foreground=40 > "$issue_file"
fi

# Create temporary files for diff and commits
diff_file=$(mktemp)
commits_file=$(mktemp)

# Ensure the temporary files are deleted when the script exits
trap "rm -f $diff_file $commits_file" EXIT

# Run the git and pinocchio commands
git diff "${base_branch}" HEAD > "$diff_file"
git log --stat "${base_branch}..HEAD" > "$commits_file"

# Attempt the pinocchio command; delete description and issue files on success
if pinocchio code create-pull-request \
    --description "$(cat $description_file)" \
    --commits "$commits_file" \
    --issue "$issue_file" \
   ; then
    echo "Deleting $description_file and $issue_file"
    rm -f "$description_file" "$issue_file"
fi

