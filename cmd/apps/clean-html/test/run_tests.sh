#!/bin/bash

# Function to run a single test
run_test() {
    local html_file=$1
    local config_file=$2
    local output_file=$3

    echo "Running test: $html_file with $config_file"
    cat "$html_file" | python ../extract_html.py --config "$config_file" > "$output_file"
    echo "Output saved to $output_file"
    echo "---"
}

# Create output directory if it doesn't exist
mkdir -p test_output

# Run all tests
run_test "test/simple.html" "test/simple_config.yaml" "test_output/simple_output.yaml"
run_test "test/list.html" "test/list_config.yaml" "test_output/list_output.yaml"
run_test "test/nested.html" "test/nested_config.yaml" "test_output/nested_output.yaml"
run_test "test/attribute.html" "test/attribute_config.yaml" "test_output/attribute_output.yaml"
run_test "test/sample.html" "test/config.yaml" "test_output/complex_output.yaml"

echo "All tests completed. Outputs are in the test_output directory."