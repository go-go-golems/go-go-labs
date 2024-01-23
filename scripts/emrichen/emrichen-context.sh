#!/bin/bash

# Default values for flags
with_example=false
function_name=""
test_name=""
with_utils=false
with_utils_body=false
with_utils_test=false
with_readme=false

# Function to display the help message
show_help() {
    echo "Usage: $0 [options]"
    echo "Options:"
    echo "  --with-example        Display the example (default: false)"
    echo "  --function NAME       Run the function with the specified name"
    echo "  --test NAME           Display the test with the specified name"
    echo "  --with-utils          Include utils (default: true)"
    echo "  --with-utils-body     Include the body of utils (default: false)"
    echo "  --with-readme         Include the README file (default: true)"
    echo "  --with-utils-test     Include the utils test (default: true)"
    echo "  --with-interpreter-spec Include the interpreter spec (default: true)"
    echo "  --help                Display this help message and exit"
}

# Function to handle the --function flag
handle_function() {
    if [[ -n "$function_name" ]]; then
        oak go definitions emrichen-interpreter/emrichen.go --with-body --function-name "$function_name"
    fi
}

# Function to handle the --test flag
handle_test() {
    if [[ -n "$test_name" ]]; then
        cat emrichen-interpreter/emrichen-"${test_name}"_test.go
    fi
}

# Function to handle the --with-example flag
handle_with_example() {
    if [[ "$with_example" == true ]]; then
        echo "# Concrete emrichen example"
        echo "This example showcases most of the emrichen feature. Use it as a guide when asked to generate emrichen YAML according to the spec."
        echo "This is the authoritative source for the emrichen syntax."
        cat emrichen/examples/peano.yml
    fi
}

# Function to handle the --with-utils flag
handle_with_utils() {
    if [[ "$with_utils" == true ]]; then
        echo "# Go utilities to deal with YAML custom tags"
        echo "This describes the utility functions we have to achieve different things with YAML custom tags."
        if [[ "$with_utils_body" == true ]]; then
            oak go definitions emrichen-interpreter/utils.go --with-body
        else
            oak go definitions emrichen-interpreter/utils.go
        fi
    fi
}
#
# Function to handle the --with-utils-test flag
handle_with_utils_test() {
    if [[ "$with_utils_test" == true ]]; then
        echo "# Go test utilities to deal with YAML custom tags"
        echo "This describes the utility functions we have to help write unit tests for the emrichen interpreter."
        oak go definitions emrichen-interpreter/utils_test.go --with-body
    fi
}


# Function to handle the --with-readme flag
handle_with_readme() {
    if [[ "$with_readme" == true ]]; then
        echo "# Emrichen README.md"
        echo "This describes the Emrichen YAML custom tags."
        echo "Use this as a reference guide for the Emrichen syntax and functionality. This is the authoritative source for the Emrichen syntax."
        echo ""
        cat emrichen-readme.md
        echo "---"
        echo
    fi
}

handle_interpreter_spec() {
    if [[ "$with_interpreter_spec" == true ]]; then
        echo "# Emrichen interpreter spec"
        echo "This describes the Emrichen interpreter."
        echo "Use this as a reference guide for the Emrichen interpreter. This is the authoritative source for the Emrichen interpreter."
        echo ""
        oak go definitions emrichen-interpreter/emrichen.go
        echo "---"
        echo
    fi
}

# Check for --help flag before processing other options
for arg in "$@"; do
    if [[ $arg == "--help" ]]; then
        show_help
        exit 0
    fi
done

# # Parse command line options
while [[ $# -gt 0 ]]; do
    case $1 in
        --with-example) with_example=true; shift;;
        --function) function_name=$2; shift 2;;
        --test) test_name=$2; shift 2;;
        --with-utils) with_utils=true; shift;;
        --with-utils-test) with_utils_test=true; shift;;
        --with-utils-body) with_utils_body=true; shift;;
        --with-interpreter-spec) with_interpreter_spec=true; shift;;
        --with-readme) with_readme=true; shift;;
        --help) show_help; exit 0;;
        *) echo "Unknown option: $1" >&2; exit 1;;
    esac
done


# Execute actions based on flags
handle_with_readme
handle_with_example
handle_interpreter_spec
handle_with_utils
handle_with_utils_test
handle_function
handle_test
