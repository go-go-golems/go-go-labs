#!/usr/bin/env bash

print_help() {
    cat << EOF
Usage: $(basename "$0") [options]

This script is used to generate emrichen related code.

Options:
  --help                 Display this help text and exit.
  Any other flags passed will be forwarded to the pinocchio command.

EOF
}

# Parse options and forward them to pinocchio command
options=()
while [[ $# -gt 0 ]]; do
    case "$1" in
        --help)
            print_help
            exit 0
            ;;
        *)
            # Add the option to the options array
	   option="$1"
	   options+=("$option")
	   echo -- $option
            ;;
    esac
    shift
done

# Define a function to run the pinocchio command
# The function takes a boolean flag to determine whether to include --print-prompt
run_pinocchio() {
    local print_prompt=$1
    local prompt_flag=""

    # If print_prompt is true, add the --print-prompt flag
    if [[ $print_prompt == true ]]; then
        prompt_flag="--print-prompt"
    fi

    # Run the pinocchio command with or without the --print-prompt flag
    pinocchio code go \
	--context emrichen-readme.md \
	--context emrichen/examples/peano.yml \
	--context /home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/experiments/yaml-custom-tags/emrichen.go \
	$prompt_flag --bracket --ai-max-response-tokens 3000 \
	"${options[@]}" 
}

# Run the command with the --print-prompt flag and capture the output
output=$(run_pinocchio true)

# Use pinocchio tokens count to count the tokens and print the count
token_count=$(echo "$output" | pinocchio tokens count -)
echo "Token count: $token_count"

# Ask the user for confirmation before moving on
read -p "Are you sure you want to proceed? View token count file? Copy to clipboard? [y/N/v/c] " -n 1 -r
echo    # move to a new line

if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Run the command again without the --print-prompt flag
    run_pinocchio false
elif [[ $REPLY =~ ^[Cc]$ ]]; then
    # Copy to clipboard
    echo "$output" | xsel -b
    echo "Output copied to clipboard"
elif [[ $REPLY =~ ^[Vv]$ ]]; then
    # View the token count file using the terminal pager
    echo "$output" | less
else
    echo "Operation canceled."
    exit 1
fi

