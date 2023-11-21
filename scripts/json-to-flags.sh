#!/bin/bash

# This script takes a JSON file as input and transforms each key-value pair
# into a shell script argument format --key 'value', skipping any pairs where
# the value is an empty string, array, or object. The resulting arguments can
# be passed to other shell scripts.

usage() {
  echo "Usage: $0 <filename.json>"
  echo "Transform JSON key-value pairs into shell script arguments."
  echo
  echo "Options:"
  echo "  -h, --help    Show this help message and exit."
}

if [[ "$1" == "-h" || "$1" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -ne 1 ]]; then
  echo "Error: Invalid number of arguments."
  usage
  exit 1
fi

json_file="$1"

if [[ ! -f "$json_file" ]]; then
  echo "Error: File '$json_file' not found."
  exit 1
fi

json_input=$(<"$json_file")

echo "$json_input" | jq -r 'to_entries[] | select(.value != "" and .value != [] and .value != {}) | "--\(.key) \\\"\(.value)\\\""' | xargs echo
