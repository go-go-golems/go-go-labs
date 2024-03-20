#!/bin/bash

# Help section
show_help() {
    echo "Usage: $0 [-h] [filename_template]"
    echo ""
    echo "Options:"
    echo " -h   Show this help message and exit"
    echo ""
    echo "Arguments:"
    echo " filename_template   Optional filename template for the mktemp call"
    echo ""
    echo "Description:"
    echo " This script gets the current clipboard content, opens it in the default editor,"
    echo " and then updates the clipboard content with the new content after exiting the editor."
    echo " If the EDITOR environment variable is not set, it defaults to 'vim'."
}

# Parse options
while getopts ":h" opt; do
    case ${opt} in
        h )
            show_help
            exit 0
            ;;
        \? )
            echo "Invalid Option: -$OPTARG" 1>&2
            show_help
            exit 1
            ;;
    esac
done
shift $((OPTIND -1))

# Set the default editor to vim if EDITOR is not set
if [[ -z "$EDITOR" ]]; then
    EDITOR="emacsclient -nw"
fi

# Function to clean up the temporary file
cleanup() {
    rm -f "$tempfile"
}

# Trap any errors and call the cleanup function
trap 'cleanup' ERR

# Get the current clipboard content
content=$(xsel -b)

# Check if a filename template was provided
if [[ -n "$1" ]]; then
    # Use the provided template for the mktemp call
    tempfile=$(mktemp "XXXXXX.$1")
else
    # No template provided, so use the default mktemp behavior
    tempfile=$(mktemp)
fi

echo "$content" > "$tempfile"

# Open the file in the default editor
$EDITOR "$tempfile"

# After exiting the editor, read the new content and put it back into the clipboard
new_content=$(cat "$tempfile")
xsel -b <<< "$new_content"

# Call the cleanup function to remove the temporary file
cleanup

