#!/bin/bash

# Script to test the functionality of external analysis tools:
# - Magika
# - ExifTool
# - jdupes

DOWNLOADS_DIR="/home/manuel/Downloads"
SAMPLE_FILE_COUNT=5
JDUPE_TEST_DIR="tmp_jdupes_test_$$" # Use PID for uniqueness

# Function to print a formatted header
print_header() {
    echo
    echo "=================================================="
    echo "  TESTING: $1"
    echo "=================================================="
}

# Function to check for a tool and print status
check_tool() {
    local tool_name=$1
    local install_cmd=$2
    echo -n "Checking for $tool_name... "
    if command -v "$tool_name" &> /dev/null; then
        echo "FOUND ($(command -v "$tool_name"))"
        return 0
    else
        echo "NOT FOUND. Install with: $install_cmd"
        return 1
    fi
}

# Function to run and print command/output
run_test_command() {
    local description=$1
    local cmd=$2
    echo "---"
    echo "Test: $description"
    echo "Command: $cmd"
    echo "Output:"
    echo "-------"
    eval "$cmd" # Use eval to handle complex commands with quotes/pipes properly
    echo "-------"
    echo "[Return Code: $?]" 
}

# --- Tool Availability Check ---
print_header "Tool Availability"
MAGIKA_OK=0
EXIFTOOL_OK=0
JDUPES_OK=0
check_tool "magika" "pip install magika" && MAGIKA_OK=1
check_tool "exiftool" "sudo apt install libimage-exiftool-perl" && EXIFTOOL_OK=1
check_tool "jdupes" "sudo apt install jdupes" && JDUPES_OK=1

# --- Find Sample Files ---
print_header "Sample Files"
SAMPLE_FILES=()
while IFS= read -r -d $'\0' file; do
    SAMPLE_FILES+=("$file")
done < <(find "$DOWNLOADS_DIR" -maxdepth 1 -type f -print0 | head -z -n $SAMPLE_FILE_COUNT)

if [ ${#SAMPLE_FILES[@]} -eq 0 ]; then
    echo "No files found in $DOWNLOADS_DIR to use as samples. Exiting."
    exit 1
fi
echo "Using the following sample files:"
printf -- '- %s\n' "${SAMPLE_FILES[@]}"

# --- Test Magika ---
if [ $MAGIKA_OK -eq 1 ]; then
    print_header "Magika"
    for file in "${SAMPLE_FILES[@]}"; do
        run_test_command "Magika plain output for '$(basename "$file")'" "magika \"$file\""
    done
    # Test JSON output on the first file
    if [ ${#SAMPLE_FILES[@]} -gt 0 ]; then
        run_test_command "Magika JSON output for '$(basename "${SAMPLE_FILES[0]}")'" "magika --json \"${SAMPLE_FILES[0]}\""
    fi
    # Test batch JSON output
    TEMP_MAGIKA_INPUT_LIST=$(mktemp)
    printf "%s\0" "${SAMPLE_FILES[@]}" > "$TEMP_MAGIKA_INPUT_LIST"
    
    echo "---"
    echo "Test: Magika batch JSON output (${#SAMPLE_FILES[@]} files) using xargs"
    echo "Command: xargs -0 -I {} magika --json {} < \"$TEMP_MAGIKA_INPUT_LIST\""
    echo "Output:"
    echo "-------"
    xargs -0 -I {} magika --json {} < "$TEMP_MAGIKA_INPUT_LIST"
    batch_rc=$?
    echo "-------"
    echo "[Return Code: $batch_rc]"
    rm -f "$TEMP_MAGIKA_INPUT_LIST"
else
    echo "[SKIPPING Magika tests - tool not found]"
fi

# --- Test ExifTool ---
if [ $EXIFTOOL_OK -eq 1 ]; then
    print_header "ExifTool"
    if [ ${#SAMPLE_FILES[@]} -gt 0 ]; then
        FIRST_FILE="${SAMPLE_FILES[0]}"
        run_test_command "ExifTool basic tags for '$(basename "$FIRST_FILE")'" "exiftool -s3 -FileType -ImageSize -Duration -CreateDate -MIMEType \"$FIRST_FILE\""
    else
        echo "No sample files found to test ExifTool."
    fi
else
    echo "[SKIPPING ExifTool tests - tool not found]"
fi

# --- Test jdupes ---
if [ $JDUPES_OK -eq 1 ]; then
    print_header "jdupes"
    echo "Creating temporary directory structure for jdupes test..."
    mkdir -p "$JDUPE_TEST_DIR/subdir"
    echo "duplicate content" > "$JDUPE_TEST_DIR/fileA.txt"
    echo "duplicate content" > "$JDUPE_TEST_DIR/subdir/fileB.txt"
    echo "unique content" > "$JDUPE_TEST_DIR/fileC.txt"
    echo "Created:"
    find "$JDUPE_TEST_DIR" -type f -print
    echo

    run_test_command "jdupes recursive list" "jdupes -r "$JDUPE_TEST_DIR""
    run_test_command "jdupes recursive summary" "jdupes -rS "$JDUPE_TEST_DIR""
    run_test_command "jdupes recursive list (quiet)" "jdupes -rq "$JDUPE_TEST_DIR""

    echo "Cleaning up temporary jdupes test directory..."
    rm -rf "$JDUPE_TEST_DIR"
    echo "Done."
else
    echo "[SKIPPING jdupes tests - tool not found]"
fi

print_header "Testing Complete" 