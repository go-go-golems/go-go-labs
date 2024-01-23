#!/usr/bin/env bash

print_help() {
    cat << EOF
Usage: $(basename "$0") [options]

This script is used to generate emrichen related unit tests.

Options:
  --help                 Display this help text and exit.
  Flags for emrichen-context.sh are forwarded and removed from the pinocchio call.

EOF
}

# Parse options and forward them to emrichen-context.sh and pinocchio command
emrichen_context_options=(--with-readme --with-example --with-utils-test --with-interpreter-spec)
pinocchio_options=()
while [[ $# -gt 0 ]]; do
    case "$1" in
        --help)
            print_help
            exit 0
            ;;
        --with-example|--with-readme|--with-utils-test|--with-utils|--with-utils-body|--function|--test)
            # Add the option to the emrichen_context_options array
            emrichen_context_options+=("$1")
            if [[ "$1" == "--function" || "$1" == "--test" ]]; then
                # Next argument is part of the current flag
                shift
                emrichen_context_options+=("$1")
            fi
            ;;
        *)
            # Add other options to the pinocchio_options array
            pinocchio_options+=("$1")
            ;;
    esac
    shift
done

# Define a function to run the emrichen-context.sh and pinocchio commands
run_pinocchio() {
    local print_prompt=$1
    local prompt_flag=""

    # If print_prompt is true, add the --print-prompt flag
    if [[ $print_prompt == true ]]; then
        prompt_flag="--print-prompt"
    fi

    # Run emrichen-context.sh with its options
    emrichen_context_output=$(./emrichen-context.sh "${emrichen_context_options[@]}" | temporizer)

    # Run the pinocchio command with the output from emrichen-context.sh
    pinocchio code unit-tests --bracket --ai-max-response-tokens 3000 \
    --code "$(echo "$emrichen_context_output")" \
    $prompt_flag "${pinocchio_options[@]}"
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
