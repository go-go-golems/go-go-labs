#!/bin/bash

# Script to analyze the contents of the Downloads folder
# Uses modern tools for file identification and organization:
# - Magika: Google's AI-powered file type detection
# - ExifTool: Advanced metadata extraction
# - jdupes: Fast file deduplication

DOWNLOADS_DIR="/home/manuel/Downloads"
OUTPUT_FILE="downloads_analysis.txt"
DEBUG_LOG="downloads_analysis_debug.log"

# --- Verbose Option --- Default to 0 (off)
VERBOSE=0
# --- Sampling Option --- Default to 0 (off)
SAMPLE_PER_DIR=0

# --- Bash Version Check ---
# Associative arrays require Bash 4.0+
if (( BASH_VERSINFO[0] < 4 )); then
    echo "Error: This script requires Bash version 4.0 or higher." >&2
    exit 1
fi

# --- Setup ---
# Enable debugging and status updates
set -o pipefail
exec 3>"$DEBUG_LOG" # File descriptor 3 for debug log

# --- Logging Functions ---
# Status display function - updates in place
show_status() {
    # If verbose, print status on a new line, otherwise overwrite
    if [ "$VERBOSE" -eq 1 ]; then
        echo "[STATUS] $1" >&2
    else
        echo -en "\r\033[K[STATUS] $1" >&2
    fi
}

# Debug function that logs to file and stderr (ensure it doesn't leak to stdout)
debug() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp DEBUG] $1" >&3 # Write to debug log file (FD 3)
    echo "[$timestamp DEBUG] $1" >&2 # Always write primary debug to stderr
}

# Verbose debug function - only prints to stderr if VERBOSE=1
vdebug() {
    if [ "$VERBOSE" -eq 1 ]; then
        local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
        echo "[$timestamp VDEBUG] $1" >&2 # Write to stderr only if VERBOSE
    fi
}

# Function to safely get file size
get_file_size() {
    du -b "$1" 2>/dev/null | cut -f1 || echo "0"
}

# Function to convert bytes to human-readable format
human_readable_size() {
    local size=$1
    numfmt --to=iec-i --suffix=B --format="%.1f" "$size" 2>/dev/null || echo "${size} B"
}

debug "Script started at $(date)"
debug "Analyzing directory: $DOWNLOADS_DIR"

# Check if Downloads directory exists
if [ ! -d "$DOWNLOADS_DIR" ]; then
    echo "Error: Downloads directory not found at $DOWNLOADS_DIR" >&2
    debug "Downloads directory not found at $DOWNLOADS_DIR"
    exit 1
fi

# --- Argument Parsing ---
TEMP=$(getopt -o vs: --long verbose,sample-per-dir: -n '$0' -- "$@")
if [ $? != 0 ] ; then echo "Terminating..." >&2 ; exit 1 ; fi

# Note the quotes around '$TEMP': they are essential!
eval set -- "$TEMP"

while true; do
  case "$1" in
    -v | --verbose ) VERBOSE=1; shift ;; # Set verbose flag
    -s | --sample-per-dir ) SAMPLE_PER_DIR="$2"; shift 2 ;; # Set sample limit
    -- ) shift; break ;; # End of options
    * ) break ;; # Unexpected option
  esac
done

if [ "$VERBOSE" -eq 1 ]; then
    debug "Verbose mode enabled."
fi

if ! [[ "$SAMPLE_PER_DIR" =~ ^[0-9]+$ ]]; then
    echo "Error: --sample-per-dir value must be a non-negative integer." >&2
    debug "Invalid --sample-per-dir value: $SAMPLE_PER_DIR"
    exit 1
elif [ "$SAMPLE_PER_DIR" -gt 0 ]; then
    debug "Directory sampling enabled: Max $SAMPLE_PER_DIR files per directory for type analysis."
fi

# --- Tool Detection ---
# Check for required/optional tools
check_tool() {
    show_status "Checking for tool: $1"
    if ! command -v "$1" &> /dev/null; then
        # Clear status line before printing warning
        echo -e "\r\033[K[WARNING] $1 not found. Install with: $2" >&2
        debug "Tool not found: $1"
        return 1
    fi
    debug "Tool found: $1"
    return 0
}

MAGIKA_AVAILABLE=0
if check_tool "magika" "pip install magika"; then
    MAGIKA_AVAILABLE=1
fi

EXIFTOOL_AVAILABLE=0
if check_tool "exiftool" "sudo apt install libimage-exiftool-perl"; then
    EXIFTOOL_AVAILABLE=1
fi

JDUPES_AVAILABLE=0
if check_tool "jdupes" "sudo apt install jdupes"; then
    JDUPES_AVAILABLE=1
fi

# Clear status line after checks
echo -e "\r\033[K" >&2
echo "Analyzing Downloads folder: $DOWNLOADS_DIR"
echo "Output will be saved to: $OUTPUT_FILE"
echo "Debug log: $DEBUG_LOG"

# --- Initialization ---
# Create output file with header
show_status "Creating output file: $OUTPUT_FILE"
{
    echo "=== DOWNLOADS FOLDER ANALYSIS ==="
    echo "Generated on: $(date)"
    echo "===================================="
    echo ""
} > "$OUTPUT_FILE"

# --- Basic Statistics (Top Level Only) ---
show_status "Collecting basic top-level statistics"
debug "Collecting basic top-level statistics for $DOWNLOADS_DIR"
{
    echo "BASIC STATISTICS (Top Level Only):"
    TOTAL_ITEMS_TOP=$(find "$DOWNLOADS_DIR" -maxdepth 1 -mindepth 1 -printf '.' | wc -c)
    TOTAL_FILES_TOP=$(find "$DOWNLOADS_DIR" -maxdepth 1 -type f -printf '.' | wc -c)
    TOTAL_DIRS_TOP=$(find "$DOWNLOADS_DIR" -maxdepth 1 -type d -mindepth 1 -printf '.' | wc -c)
    TOTAL_SIZE=$(du -sh "$DOWNLOADS_DIR" | cut -f1)
    echo "Total items (top level): $TOTAL_ITEMS_TOP"
    echo "Total files (top level): $TOTAL_FILES_TOP"
    echo "Total directories (top level): $TOTAL_DIRS_TOP"
    echo "Total size (recursive): $TOTAL_SIZE"
    echo ""
} >> "$OUTPUT_FILE"

# --- Pre-computation: File List, Sizes, and Types (Recursive) ---
show_status "Finding all files recursively..."
TEMP_ALL_FILES_LIST=$(mktemp)
# Exclude __MACOSX directories and find all regular files, null-terminated
find "$DOWNLOADS_DIR" -path '*/__MACOSX/*' -prune -o -type f -print0 > "$TEMP_ALL_FILES_LIST"
TOTAL_FILES_RECURSIVE=$(grep -zc $'\0' "$TEMP_ALL_FILES_LIST")
debug "Found $TOTAL_FILES_RECURSIVE files recursively (excluding __MACOSX). List stored in $TEMP_ALL_FILES_LIST"

declare -A file_paths # Store paths (might be redundant but useful)
declare -A file_sizes # Store sizes: path -> size_in_bytes
declare -A file_types # Store types: path -> ct_label string
declare -A file_groups # Store types: path -> group string
declare -A file_mod_times # Store modification times: path -> YYYY-MM-DD HH:MM:SS
declare -A file_mod_epochs # Store modification epoch seconds: path -> epoch

FILE_COUNT=0
show_status "Collecting file sizes and modification times (0/$TOTAL_FILES_RECURSIVE)..."
debug "Starting collection of sizes and modification times."
while IFS= read -r -d $'\0' filepath; do
    FILE_COUNT=$((FILE_COUNT + 1))
    if (( FILE_COUNT % 100 == 0 )); then # Update status periodically
         show_status "Collecting file sizes and modification times ($FILE_COUNT/$TOTAL_FILES_RECURSIVE)..."
    fi
    vdebug "Processing file $FILE_COUNT/$TOTAL_FILES_RECURSIVE: $filepath"
    file_paths["$filepath"]=1
    file_sizes["$filepath"]=$(get_file_size "$filepath")
    mod_epoch=$(stat -c '%Y' "$filepath" 2>/dev/null || echo "0")
    file_mod_epochs["$filepath"]=$mod_epoch
    file_mod_times["$filepath"]=$(date -d "@$mod_epoch" '+%Y-%m-%d %H:%M:%S' 2>/dev/null || echo "Unknown")
done < "$TEMP_ALL_FILES_LIST"
debug "Finished collecting sizes and modification times for $TOTAL_FILES_RECURSIVE files."


show_status "Determining file types (0/$TOTAL_FILES_RECURSIVE)..."
if [ "$MAGIKA_AVAILABLE" -eq 1 ]; then
    debug "Using Magika for recursive file type identification"
    TEMP_MAGIKA_OUTPUT=$(mktemp)
    # Use xargs to process files; output is one JSON object per line

    # Determine the input list for Magika (potentially sampled)
    local MAGIKA_INPUT_LIST="$TEMP_ALL_FILES_LIST"
    local TOTAL_FILES_FOR_MAGIKA=$TOTAL_FILES_RECURSIVE
    local TEMP_SAMPLED_FOR_MAGIKA=""

    if [ "$SAMPLE_PER_DIR" -gt 0 ]; then
        debug "Applying sampling (max $SAMPLE_PER_DIR per directory)..."
        TEMP_SAMPLED_FOR_MAGIKA=$(mktemp)
        # Use awk to filter the list based on directory counts
        # Reads null-terminated input, outputs null-terminated paths
        awk -v limit="$SAMPLE_PER_DIR" 'BEGIN{FS=OFS=RS="\0"; ORS="\0"} {
            dir = $0; sub(/\/[^\/]*$/, "", dir); # Get dirname
            if (count[dir] < limit) {
                print $0;
                count[dir]++;
            }
        }' "$TEMP_ALL_FILES_LIST" > "$TEMP_SAMPLED_FOR_MAGIKA"

        TOTAL_FILES_FOR_MAGIKA=$(grep -zc $'\0' "$TEMP_SAMPLED_FOR_MAGIKA")
        debug "Created sampled list for Magika: $TOTAL_FILES_FOR_MAGIKA files in $TEMP_SAMPLED_FOR_MAGIKA"
        MAGIKA_INPUT_LIST="$TEMP_SAMPLED_FOR_MAGIKA"
    else
        debug "Sampling disabled. Preparing to analyze all $TOTAL_FILES_RECURSIVE files with Magika."
    fi

    vdebug "Starting xargs magika command on $TOTAL_FILES_FOR_MAGIKA files (this might take a while)..."
    xargs -0 -I {} magika --json {} < "$MAGIKA_INPUT_LIST" > "$TEMP_MAGIKA_OUTPUT" 2>>"$DEBUG_LOG"
    local magika_rc=$? # Store return code (local is valid inside functions)
    debug "Magika xargs process completed with rc=$magika_rc. Output in $TEMP_MAGIKA_OUTPUT"
    vdebug "Finished xargs magika command."

    # Cleanup sampled list file if it was created
    [ -n "$TEMP_SAMPLED_FOR_MAGIKA" ] && rm -f "$TEMP_SAMPLED_FOR_MAGIKA"

    # Process Magika JSON output (one object per line)
    FILE_COUNT=0
    while IFS= read -r line; do
        FILE_COUNT=$((FILE_COUNT + 1))
         if (( FILE_COUNT % 100 == 0 )); then # Update status periodically
            vdebug "Processing Magika output line $FILE_COUNT/$TOTAL_FILES_FOR_MAGIKA..."
            show_status "Processing Magika types ($FILE_COUNT/$TOTAL_FILES_FOR_MAGIKA)..."
        fi
        # Extract path, type (label), and group using jq or fallback
        if command -v jq &> /dev/null; then
            path=$(echo "$line" | jq -r '.path' 2>/dev/null)
            type=$(echo "$line" | jq -r '.result.value.output.ct_label' 2>/dev/null) # Use ct_label for readable type
            group=$(echo "$line" | jq -r '.result.value.output.group' 2>/dev/null)
        else
            # Basic parsing, less robust
            path=$(echo "$line" | cut -d '"' -f 4) # Assuming "path": "...",
            type=$(echo "$line" | grep -o '"ct_label": "[^"]*"' | cut -d '"' -f 4)
            group=$(echo "$line" | grep -o '"group": "[^"]*"' | cut -d '"' -f 4)
        fi

        if [[ -n "$path" && -n "$type" ]]; then
            file_types["$path"]="$type"
            file_groups["$path"]="${group:-unknown}" # Store group, default to unknown
            # debug "Magika type for '$path': $type" # Too verbose for log
        else
             # Handle cases where Magika might fail or parsing might fail
             file_types["$path"]="Unknown (Magika Error)"
             file_groups["$path"]="unknown"
             debug "Magika failed to determine type for: $path (Line: $line)"
        fi
    done < "$TEMP_MAGIKA_OUTPUT" # Read line by line from the output file

    rm -f "$TEMP_MAGIKA_OUTPUT"
    debug "Finished processing Magika types."
else
    debug "Using 'file' command for recursive file type identification (Magika not available)"
    vdebug "Starting fallback 'file' command processing..."
    FILE_COUNT=0
    while IFS= read -r -d $'\0' filepath; do
        FILE_COUNT=$((FILE_COUNT + 1))
        if (( FILE_COUNT % 100 == 0 )); then # Update status periodically
            vdebug "Processing file $FILE_COUNT/$TOTAL_FILES_RECURSIVE with 'file' command..."
            show_status "Determining file types with 'file' ($FILE_COUNT/$TOTAL_FILES_RECURSIVE)..."
        fi
        file_types["$filepath"]=$(file -b "$filepath" 2>/dev/null || echo "Unknown (file error)")
    done < "$TEMP_ALL_FILES_LIST"
    debug "Finished determining types with 'file' command."
fi

# --- File Types Analysis (Recursive) ---
show_status "Aggregating file types..."
debug "Starting aggregation of recursive file types."
{
    echo "FILE TYPES ANALYSIS (Recursive):"
    if [ "$MAGIKA_AVAILABLE" -eq 1 ]; then
        echo "Using Magika (AI-powered identification)"
    else
        echo "Using file command (Magika not available)"
    fi
    if [ "$SAMPLE_PER_DIR" -gt 0 ]; then
        echo "(Sampled: Max $SAMPLE_PER_DIR files per directory analyzed for types)"
    fi
    echo "--------------------------------------------------"
    printf "%-40s | %8s | %12s\n" "Type" "Count" "Total Size"
    echo "--------------------------------------------------"
} >> "$OUTPUT_FILE"

declare -A type_counts
declare -A type_sizes

for filepath in "${!file_types[@]}"; do
    type="${file_types[$filepath]}"
    size="${file_sizes[$filepath]}"
    
    ((type_counts["$type"]++))
    ((type_sizes["$type"]+=size))
done

# Prepare sorted output for file types
TEMP_SORTED_TYPES=$(mktemp)
for type in "${!type_counts[@]}"; do
    count="${type_counts[$type]}"
    size="${type_sizes[$type]}"
    human_size=$(human_readable_size "$size")
    printf "%s|%d|%s|%s\n" "$type" "$count" "$size" "$human_size" >> "$TEMP_SORTED_TYPES"
done

# Sort by count descending and append to output file
sort -t'|' -k2 -nr "$TEMP_SORTED_TYPES" | while IFS='|' read -r type count _ human_size; do
     printf "%-40s | %8d | %12s\n" "${type:0:40}" "$count" "$human_size" >> "$OUTPUT_FILE"
done

rm -f "$TEMP_SORTED_TYPES"
echo "" >> "$OUTPUT_FILE"
debug "Finished aggregating file types."


# --- Media File Analysis (Recursive, Top 10 by size) ---
if [ "$EXIFTOOL_AVAILABLE" -eq 1 ]; then
    show_status "Analyzing media file metadata..."
    debug "Starting ExifTool media analysis (Top 10 largest media files)."
    {
        echo "MEDIA FILE METADATA (Top 10 Largest Media Files):"
        echo "--------------------------------------------------"
        printf "%-30s | %-20s | %-25s | %s\n" "Filename" "Type (Magika)" "Resolution/Duration (Exif)" "Created Date (Exif)"
        echo "--------------------------------------------------"
    } >> "$OUTPUT_FILE"

    declare -A media_files # path -> size
    for filepath in "${!file_types[@]}"; do
        # Use the group identified by Magika
        group="${file_groups[$filepath]}"
        if [[ "$group" == "image" || "$group" == "video" || "$group" == "audio" ]]; then
             media_files["$filepath"]=${file_sizes["$filepath"]}
        fi
    done
    debug "Identified ${#media_files[@]} potential media files based on type."

    # Sort media files by size descending and take top 10
    MEDIA_COUNT=0
    while IFS= read -r filepath; do
        ((MEDIA_COUNT++))
        filename=$(basename "$filepath")
        show_status "Analyzing media file $MEDIA_COUNT/10: $filename"
        debug "Processing media file with ExifTool: $filepath"

        # Get metadata using ExifTool
        meta_type=$(exiftool -s3 -FileType "$filepath" 2>/dev/null || echo "N/A")
        resolution=$(exiftool -s3 -ImageSize "$filepath" 2>/dev/null)
        duration=$(exiftool -s3 -Duration "$filepath" 2>/dev/null)
        res_dur="N/A"
        if [[ -n "$resolution" ]]; then
            res_dur="$resolution"
        elif [[ -n "$duration" ]]; then
             # Attempt to format duration nicely if possible (basic example)
             res_dur=$(echo "$duration" | awk '{printf "%.2fs", $1}' || echo "$duration")
        fi

        created=$(exiftool -s3 -CreateDate "$filepath" 2>/dev/null || \
                  exiftool -s3 -MediaCreateDate "$filepath" 2>/dev/null || echo "Unknown")

        magika_type="${file_types[$filepath]:-Unknown}"

        printf "%-30s | %-20s | %-25s | %s\n" "${filename:0:30}" "${magika_type:0:20}" "${res_dur:0:25}" "$created" >> "$OUTPUT_FILE"

    done < <(for path in "${!media_files[@]}"; do printf "%s\t%s\n" "${media_files[$path]}" "$path"; done | sort -nr -k1 | head -n 10 | cut -f2)


    echo "" >> "$OUTPUT_FILE"
    debug "Completed ExifTool media analysis."
fi


# --- Duplicate Files Analysis (Recursive) ---
show_status "Starting duplicate file analysis..."
debug "Starting duplicate file analysis (recursive)."
{
    echo "DUPLICATE FILES ANALYSIS (Recursive):"
} >> "$OUTPUT_FILE"

if [ "$JDUPES_AVAILABLE" -eq 1 ]; then
    show_status "Analyzing duplicates with jdupes..."
    debug "Using jdupes for duplicate analysis."
    {
        echo "Using jdupes (fast and accurate)"
        echo "--------------------------------------------------"
    } >> "$OUTPUT_FILE"

    TEMP_JDUPES_RAW=$(mktemp)
    TEMP_JDUPES_SUMMARY=$(mktemp)
    debug "Running jdupes scan, raw output: $TEMP_JDUPES_RAW, summary: $TEMP_JDUPES_SUMMARY"

    # Run jdupes non-interactively to get the list of duplicates
    jdupes -r -q -o name "$DOWNLOADS_DIR" > "$TEMP_JDUPES_RAW" 2>>"$DEBUG_LOG"
    # Run jdupes with summary to get counts and wasted space
    jdupes -r -q -S "$DOWNLOADS_DIR" > "$TEMP_JDUPES_SUMMARY" 2>>"$DEBUG_LOG"
    debug "jdupes scans completed."

    # Parse summary
    DUPLICATE_SETS=$(grep "sets of duplicate files" "$TEMP_JDUPES_SUMMARY" | awk '{print $1}' || echo "0")
    TOTAL_DUPLICATE_FILES=$(grep "duplicate files" "$TEMP_JDUPES_SUMMARY" | head -n 1 | awk '{print $1}' || echo "0") # First line usually has total files
    WASTED_SPACE_BYTES=$(grep "would be freed" "$TEMP_JDUPES_SUMMARY" | awk '{print $1}' || echo "0")
    HUMAN_WASTED=$(human_readable_size "$WASTED_SPACE_BYTES")

    debug "jdupes results: $DUPLICATE_SETS sets, $TOTAL_DUPLICATE_FILES duplicate files, $WASTED_SPACE_BYTES bytes wasted."

    {
        echo "Duplicate sets found: $DUPLICATE_SETS"
        echo "Total duplicate files: $TOTAL_DUPLICATE_FILES (excluding originals)"
        echo "Wasted space: $HUMAN_WASTED"
    } >> "$OUTPUT_FILE"

    # Show the first few sets of duplicates from the raw output
    if [ "$DUPLICATE_SETS" -gt 0 ]; then
        echo "" >> "$OUTPUT_FILE"
        echo "Sample duplicate sets (first 5 sets):" >> "$OUTPUT_FILE"
        # Process the raw output (groups separated by blank lines)
        awk 'BEGIN{RS=""; ORS="\n\n"; set_count=0} {if (NF > 1 && set_count < 5) {print $0; set_count++}}' "$TEMP_JDUPES_RAW" >> "$OUTPUT_FILE"

        if [ "$DUPLICATE_SETS" -gt 5 ]; then
            echo "... and $(($DUPLICATE_SETS - 5)) more sets." >> "$OUTPUT_FILE"
        fi
    fi

    rm -f "$TEMP_JDUPES_RAW" "$TEMP_JDUPES_SUMMARY"
else
     # Fallback to basic MD5 hash comparison
    show_status "Analyzing duplicates with MD5 (slower)..."
    debug "Using MD5 for duplicate analysis (jdupes not available)."
    {
        echo "Using basic MD5 hash comparison (jdupes not available, slower and less reliable)"
        echo "--------------------------------------------------"
    } >> "$OUTPUT_FILE"

    TEMP_MD5_HASHLIST=$(mktemp)
    TEMP_MD5_DUPLICATES=$(mktemp)
    debug "Calculating MD5 hashes for all files..."

    FILE_COUNT=0
    while IFS= read -r -d $'\0' filepath; do
        FILE_COUNT=$((FILE_COUNT + 1))
        if (( FILE_COUNT % 100 == 0 )); then
            show_status "Calculating MD5 ($FILE_COUNT/$TOTAL_FILES_RECURSIVE)..."
        fi
        md5sum -b "$filepath" >> "$TEMP_MD5_HASHLIST" 2>>"$DEBUG_LOG" # Use binary mode
    done < "$TEMP_ALL_FILES_LIST"

    show_status "Finding duplicate MD5 hashes..."
    # Group by hash, identify hashes with count > 1
    cut -d ' ' -f 1 "$TEMP_MD5_HASHLIST" | sort | uniq -d > "$TEMP_MD5_DUPLICATES"
    DUPLICATE_HASH_COUNT=$(wc -l < "$TEMP_MD5_DUPLICATES")
    debug "Found $DUPLICATE_HASH_COUNT potentially duplicate hashes."

    if [ "$DUPLICATE_HASH_COUNT" -gt 0 ]; then
         echo "Found $DUPLICATE_HASH_COUNT potential duplicate content hashes." >> "$OUTPUT_FILE"
         echo "Sample duplicate sets (based on MD5, first 5 hashes):" >> "$OUTPUT_FILE"
         HASH_COUNT=0
         while read -r hash && [ "$HASH_COUNT" -lt 5 ]; do
             ((HASH_COUNT++))
             echo "" >> "$OUTPUT_FILE"
             echo "Files with MD5 hash: $hash" >> "$OUTPUT_FILE"
             # Grep for the hash and print the corresponding filenames
             grep "^$hash " "$TEMP_MD5_HASHLIST" | while read -r line; do
                 # Extract filename safely after hash and space+asterisk
                 filename=$(echo "$line" | sed -E "s/^$hash \\\*//")
                 echo "  - $filename" >> "$OUTPUT_FILE"
             done
         done < "$TEMP_MD5_DUPLICATES"
         if [ "$DUPLICATE_HASH_COUNT" -gt 5 ]; then
             echo "" >> "$OUTPUT_FILE"
             echo "... and $(($DUPLICATE_HASH_COUNT - 5)) more hashes with duplicates." >> "$OUTPUT_FILE"
         fi
    else
        echo "No duplicate files found based on MD5 hash." >> "$OUTPUT_FILE"
    fi

    rm -f "$TEMP_MD5_HASHLIST" "$TEMP_MD5_DUPLICATES"
fi
echo "" >> "$OUTPUT_FILE"


# --- Recent Files Analysis (Recursive, Last 30 Days) ---
show_status "Analyzing recent files (last 30 days)..."
debug "Starting recent files analysis (recursive)."
{
    echo "RECENT FILES (Last 30 Days, Recursive):"
    echo "----------------------------------------------------------------------------------------------"
    printf "%-40s | %-35s | %7s | %s\n" "Filename (Relative Path)" "Type" "Size" "Date Modified"
    echo "----------------------------------------------------------------------------------------------"
} >> "$OUTPUT_FILE"

thirty_days_ago_epoch=$(date -d '30 days ago' +%s)
TEMP_RECENT_FILES_SORTED=$(mktemp)
RECENT_COUNT=0

# Iterate through pre-collected data
for filepath in "${!file_mod_epochs[@]}"; do
    mod_epoch="${file_mod_epochs[$filepath]}"
    if [[ "$mod_epoch" -gt "$thirty_days_ago_epoch" ]]; then
        ((RECENT_COUNT++))
        filename_rel=$(realpath --relative-to="$DOWNLOADS_DIR" "$filepath" 2>/dev/null || basename "$filepath") # Relative path if possible
        type="${file_types[$filepath]:-Unknown}"
        size_bytes="${file_sizes[$filepath]:-0}"
        human_size=$(human_readable_size "$size_bytes")
        date_modified="${file_mod_times[$filepath]:-Unknown}"
        # Store epoch for sorting
        printf "%s|%s|%s|%s|%s\n" "$mod_epoch" "${filename_rel:0:40}" "${type:0:35}" "$human_size" "$date_modified" >> "$TEMP_RECENT_FILES_SORTED"
    fi
done
debug "Found $RECENT_COUNT recent files."

# Sort by modification time (epoch) descending and print
sort -t'|' -k1 -nr "$TEMP_RECENT_FILES_SORTED" | cut -d'|' -f2- | while IFS='|' read -r filename_rel type human_size date_modified; do
     printf "%-40s | %-35s | %7s | %s\n" "$filename_rel" "$type" "$human_size" "$date_modified" >> "$OUTPUT_FILE"
done

rm -f "$TEMP_RECENT_FILES_SORTED"
echo "" >> "$OUTPUT_FILE"
debug "Finished recent files analysis."


# --- Large Files Analysis (Recursive, > 100MB) ---
show_status "Finding large files (> 100MB)..."
debug "Starting large files analysis (recursive)."
{
    echo "LARGE FILES (Over 100MB, Recursive):"
    echo "------------------------------------------------------------------------------------"
    printf "%-40s | %-35s | %7s | %s\n" "Filename (Relative Path)" "Type" "Size" "Full Path"
    echo "------------------------------------------------------------------------------------"
} >> "$OUTPUT_FILE"

LARGE_FILE_THRESHOLD=$((100 * 1024 * 1024)) # 100 MB in bytes
TEMP_LARGE_FILES_SORTED=$(mktemp)
LARGE_COUNT=0

# Iterate through pre-collected sizes
for filepath in "${!file_sizes[@]}"; do
    size_bytes="${file_sizes[$filepath]}"
    if [[ "$size_bytes" -gt "$LARGE_FILE_THRESHOLD" ]]; then
         ((LARGE_COUNT++))
         filename_rel=$(realpath --relative-to="$DOWNLOADS_DIR" "$filepath" 2>/dev/null || basename "$filepath")
         type="${file_types[$filepath]:-Unknown}"
         human_size=$(human_readable_size "$size_bytes")
         # Store size for sorting
         printf "%s|%s|%s|%s|%s\n" "$size_bytes" "${filename_rel:0:40}" "${type:0:35}" "$human_size" "$filepath" >> "$TEMP_LARGE_FILES_SORTED"
    fi
done
debug "Found $LARGE_COUNT large files."

# Sort by size descending and print
sort -t'|' -k1 -nr "$TEMP_LARGE_FILES_SORTED" | cut -d'|' -f2- | while IFS='|' read -r filename_rel type human_size filepath; do
     printf "%-40s | %-35s | %7s | %s\n" "$filename_rel" "$type" "$human_size" "$filepath" >> "$OUTPUT_FILE"
done

rm -f "$TEMP_LARGE_FILES_SORTED"
echo "" >> "$OUTPUT_FILE"
debug "Finished large files analysis."


# --- Files by Year/Month Analysis (Recursive) ---
show_status "Organizing files by year/month..."
debug "Starting files by month analysis (recursive)."
{
    echo "FILES BY YEAR/MONTH (Modification Time, Recursive):"
    echo "------------------------------------"
    printf "%-10s | %8s | %12s\n" "Year-Month" "Count" "Total Size"
    echo "------------------------------------"
} >> "$OUTPUT_FILE"

declare -A month_counts
declare -A month_sizes

# Aggregate counts and sizes by month from pre-collected data
for filepath in "${!file_mod_times[@]}"; do
    mod_time="${file_mod_times[$filepath]}"
    size_bytes="${file_sizes[$filepath]}"
    # Extract YYYY-MM from timestamp
    year_month=$(echo "$mod_time" | cut -d' ' -f1 | cut -d'-' -f1,2)
    if [[ "$year_month" =~ ^[0-9]{4}-[0-9]{2}$ ]]; then
        ((month_counts["$year_month"]++))
        ((month_sizes["$year_month"]+=size_bytes))
    fi
done

# Prepare sorted output
TEMP_MONTH_SORTED=$(mktemp)
for month in "${!month_counts[@]}"; do
    count="${month_counts[$month]}"
    size="${month_sizes[$month]}"
    human_size=$(human_readable_size "$size")
    printf "%s|%d|%s\n" "$month" "$count" "$human_size" >> "$TEMP_MONTH_SORTED"
done

# Sort chronologically by year-month and print
sort -t'|' -k1 "$TEMP_MONTH_SORTED" | while IFS='|' read -r month count human_size; do
    printf "%-10s | %8d | %12s\n" "$month" "$count" "$human_size" >> "$OUTPUT_FILE"
done

rm -f "$TEMP_MONTH_SORTED"
echo "" >> "$OUTPUT_FILE"
debug "Finished files by month analysis."


# --- Recommendations ---
show_status "Generating recommendations..."
{
    echo "RECOMMENDATIONS BASED ON ANALYSIS:"
    echo "1. Review 'FILE TYPES' to understand the distribution and consider organizing by category."
    echo "2. Examine 'LARGE FILES' for items that could potentially be archived or deleted."
    echo "3. Check 'DUPLICATE FILES' to recover wasted disk space."
    if [ "$JDUPES_AVAILABLE" -eq 1 ] && [ "$DUPLICATE_SETS" -gt 0 ]; then
        echo "   - To list duplicates: jdupes -r \"$DOWNLOADS_DIR\""
        echo "   - To interactively delete duplicates: jdupes -rdN \"$DOWNLOADS_DIR\" (Use with caution!)"
    elif [ "$DUPLICATE_HASH_COUNT" -gt 0 ]; then
        echo "   - Manual review needed for MD5-based duplicates."
    fi
    echo "4. Review 'MEDIA FILE METADATA' for potential organization by date or content type."
    echo "5. Use 'FILES BY YEAR/MONTH' to identify and potentially archive very old files."
    echo "6. Consider strategies like subfolders (e.g., 'Installers', 'Documents', 'Images') or date-based folders."
} >> "$OUTPUT_FILE"
debug "Recommendations generated."


# --- Finalization ---
# Clean up main file list temp file
rm -f "$TEMP_ALL_FILES_LIST"
debug "Cleaned up main temporary file list: $TEMP_ALL_FILES_LIST"

debug "Analysis completed successfully at $(date)"
show_status "Analysis complete!"
echo -e "\r\033[K" >&2 # Clear final status line
echo ""
echo "✅ Analysis complete! Results saved to $OUTPUT_FILE"
echo "🔍 Debug log saved to $DEBUG_LOG"
echo "📄 Run 'less $OUTPUT_FILE' or 'cat $OUTPUT_FILE' to view the analysis."
echo ""
echo "Tool status:"
echo "- Magika (AI file detection): $([ "$MAGIKA_AVAILABLE" -eq 1 ] && echo "✓ Available" || echo "✗ Not installed (pip install magika)")" >&2
echo "- ExifTool (metadata extraction): $([ "$EXIFTOOL_AVAILABLE" -eq 1 ] && echo "✓ Available" || echo "✗ Not installed (sudo apt install libimage-exiftool-perl)")" >&2
echo "- jdupes (deduplication): $([ "$JDUPES_AVAILABLE" -eq 1 ] && echo "✓ Available" || echo "✗ Not installed (sudo apt install jdupes)")" >&2

exit 0 