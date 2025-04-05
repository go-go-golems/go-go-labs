#!/bin/bash

# Script to analyze the contents of the Downloads folder
# Outputs information useful for organization (file types, names, dates, sizes, etc.)

DOWNLOADS_DIR="/home/manuel/Downloads"
OUTPUT_FILE="downloads_analysis.txt"

# Check if Downloads directory exists
if [ ! -d "$DOWNLOADS_DIR" ]; then
    echo "Error: Downloads directory not found at $DOWNLOADS_DIR"
    exit 1
fi

echo "Analyzing Downloads folder: $DOWNLOADS_DIR"
echo "Output will be saved to: $OUTPUT_FILE"

# Create output file with header
echo "=== DOWNLOADS FOLDER ANALYSIS ===" > "$OUTPUT_FILE"
echo "Generated on: $(date)" >> "$OUTPUT_FILE"
echo "====================================" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Basic folder statistics
echo "BASIC STATISTICS:" >> "$OUTPUT_FILE"
TOTAL_FILES=$(find "$DOWNLOADS_DIR" -maxdepth 1 -type f | wc -l)
TOTAL_DIRS=$(find "$DOWNLOADS_DIR" -maxdepth 1 -type d | wc -l) # Count only top-level dirs
TOTAL_SIZE=$(du -sh "$DOWNLOADS_DIR" | cut -f1)
echo "Total items (top level): $(ls -1A "$DOWNLOADS_DIR" | wc -l)" >> "$OUTPUT_FILE"
echo "Total files (top level): $TOTAL_FILES" >> "$OUTPUT_FILE"
echo "Total directories (top level): $TOTAL_DIRS" >> "$OUTPUT_FILE" # Report top-level count
echo "Total size: $TOTAL_SIZE" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# File types analysis using 'file' command
echo "FILE TYPES (using file command):" >> "$OUTPUT_FILE"
echo "Type Description | Count | Total Size" >> "$OUTPUT_FILE"
echo "--------------------------------------------------" >> "$OUTPUT_FILE"
# Use process substitution and awk for aggregation
find "$DOWNLOADS_DIR" -maxdepth 1 -type f -print0 | \
  xargs -0 -I {} bash -c 'file -b "{}" && du -b "{}"' | \
  awk -F'\t' '
  BEGIN { OFS=" | "; ORS="\\n"; }
  NR % 2 == 1 { current_type=$0; }
  NR % 2 == 0 {
      size=$1;
      type_counts[current_type]++;
      type_sizes[current_type]+=size;
  }
  END {
      PROCINFO["sorted_in"] = "@val_num_desc"; # Sort by count descending
      for (type in type_counts) {
          # Convert size to human-readable format
          split("B KB MB GB TB PB", units, " ");
          s = type_sizes[type];
          u = 1;
          while (s >= 1024 && u < 6) {
              s /= 1024;
              u++;
          }
          human_size = sprintf("%.1f %s", s, units[u]);
          printf "%-40s | %5d | %12s", substr(type, 0, 40), type_counts[type], human_size;
      }
  }' | sort -k2 -nr >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"


# Files with no extension (still useful info)
echo "FILES WITH NO EXTENSION:" >> "$OUTPUT_FILE"
count=$(find "$DOWNLOADS_DIR" -maxdepth 1 -type f -not -name "*.*" | wc -l)
size_bytes=$(find "$DOWNLOADS_DIR" -maxdepth 1 -type f -not -name "*.*" -print0 | xargs -0 du -cb | grep total$ | cut -f1 2>/dev/null || echo "0")
# Convert size to human-readable format
human_size=$(numfmt --to=iec-i --suffix=B --format="%.1f" $size_bytes)
echo "Count: $count, Total Size: $human_size" >> "$OUTPUT_FILE"
find "$DOWNLOADS_DIR" -maxdepth 1 -type f -not -name "*.*" -exec ls -lh {} \; >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Recent files (last 30 days) - including 'file' type
echo "RECENT FILES (LAST 30 DAYS):" >> "$OUTPUT_FILE"
echo "Filename                     | Type (from file)                     | Size    | Date Modified" >> "$OUTPUT_FILE"
echo "----------------------------------------------------------------------------------------------" >> "$OUTPUT_FILE"
find "$DOWNLOADS_DIR" -maxdepth 1 -type f -mtime -30 -print0 | while IFS= read -r -d $'\0' file; do
    # Get details using ls and file
    details=$(ls -lh --full-time "$file")
    size=$(echo "$details" | awk '{print $5}')
    # Use --full-time for consistent date format
    date_modified=$(echo "$details" | awk '{print $6, $7}')
    filename=$(basename "$file")
    file_type=$(file -b "$file")
    printf "%-30s | %-35s | %7s | %s\\n" "${filename:0:30}" "${file_type:0:35}" "$size" "$date_modified" >> "$OUTPUT_FILE"
done | sort -k4 -r >> "$OUTPUT_FILE" # Sort by date, most recent first
echo "" >> "$OUTPUT_FILE"

# Large files (over 100MB) - including 'file' type
echo "LARGE FILES (OVER 100MB):" >> "$OUTPUT_FILE"
echo "Filename                     | Type (from file)                     | Size    | Path" >> "$OUTPUT_FILE"
echo "------------------------------------------------------------------------------------" >> "$OUTPUT_FILE"
find "$DOWNLOADS_DIR" -maxdepth 1 -type f -size +100M -print0 | while IFS= read -r -d $'\0' file; do
     details=$(ls -lh "$file")
     size=$(echo "$details" | awk '{print $5}')
     filename=$(basename "$file")
     file_type=$(file -b "$file")
     printf "%-30s | %-35s | %7s | %s\\n" "${filename:0:30}" "${file_type:0:35}" "$size" "$file" >> "$OUTPUT_FILE"
done | sort -hr -k5 >> "$OUTPUT_FILE" # Sort by size, largest first
echo "" >> "$OUTPUT_FILE"

# Old files (more than 180 days) - including 'file' type
echo "OLD FILES (MORE THAN 180 DAYS):" >> "$OUTPUT_FILE"
echo "Filename                     | Type (from file)                     | Size    | Date Modified" >> "$OUTPUT_FILE"
echo "----------------------------------------------------------------------------------------------" >> "$OUTPUT_FILE"
find "$DOWNLOADS_DIR" -maxdepth 1 -type f -mtime +180 -print0 | while IFS= read -r -d $'\0' file; do
    details=$(ls -lh --full-time "$file")
    size=$(echo "$details" | awk '{print $5}')
    date_modified=$(echo "$details" | awk '{print $6, $7}')
    filename=$(basename "$file")
    file_type=$(file -b "$file")
    printf "%-30s | %-35s | %7s | %s\\n" "${filename:0:30}" "${file_type:0:35}" "$size" "$date_modified" >> "$OUTPUT_FILE"
done | sort -k4 >> "$OUTPUT_FILE" # Sort by date, oldest first
echo "" >> "$OUTPUT_FILE"

# Duplicate files (based on MD5 hash) - Consider performance on large folders
echo "POTENTIAL DUPLICATE FILES (Top 10 groups by count):" >> "$OUTPUT_FILE"
echo "MD5 Hash                         | Count | Files" >> "$OUTPUT_FILE"
echo "-------------------------------------------------------------------" >> "$OUTPUT_FILE"
# Limit processing for performance; consider fdupes or similar for thorough checks
find "$DOWNLOADS_DIR" -maxdepth 1 -type f -print0 | xargs -0 md5sum | \
  sort | awk '{print $1}' | uniq -c | sort -rn | head -n 10 | while read count hash; do
    if [ "$count" -gt 1 ]; then
        echo "$hash | $count |" >> "$OUTPUT_FILE"
        find "$DOWNLOADS_DIR" -maxdepth 1 -type f -print0 | xargs -0 md5sum | grep "^$hash" | awk '{print "    - " $2}' >> "$OUTPUT_FILE"
    fi
done
echo "Note: Duplicate check limited to top-level files and top 10 hash groups for performance." >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Organize files by year/month for timeline perspective
echo "FILES BY YEAR/MONTH (Modification Time):" >> "$OUTPUT_FILE"
echo "Year-Month | Count | Total Size" >> "$OUTPUT_FILE"
echo "------------------------------------" >> "$OUTPUT_FILE"
find "$DOWNLOADS_DIR" -maxdepth 1 -type f -printf "%TY-%Tm\\t%s\\n" | sort | awk -F'\t' '{
    month_counts[$1]++;
    month_sizes[$1]+=$2;
}
END {
    PROCINFO["sorted_in"] = "@ind_str_asc"; # Sort by Year-Month ascending
    for (month in month_counts) {
        # Convert size to human-readable format
        split("B KB MB GB TB PB", units, " ");
        s = month_sizes[month];
        u = 1;
        while (s >= 1024 && u < 6) {
            s /= 1024;
            u++;
        }
        human_size = sprintf("%.1f %s", s, units[u]);
        printf "%-10s | %5d | %12s\\n", month, month_counts[month], human_size;
    }
}' >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Add recommendations section
echo "RECOMMENDATIONS BASED ON ANALYSIS:" >> "$OUTPUT_FILE"
echo "1. Review the 'FILE TYPES' section. Consider creating folders for dominant types (e.g., Documents, Images, Archives, Videos)." >> "$OUTPUT_FILE"
echo "2. Examine 'LARGE FILES'. Delete or move any large files that are no longer needed in Downloads." >> "$OUTPUT_FILE"
echo "3. Check 'OLD FILES'. Archive or delete files you haven't accessed in over 6 months." >> "$OUTPUT_FILE"
echo "4. Investigate 'POTENTIAL DUPLICATE FILES'. Use a dedicated tool like 'fdupes' or 'rmlint' for a thorough check and cleanup if needed." >> "$OUTPUT_FILE"
echo "5. Look at 'FILES BY YEAR/MONTH'. This can help identify periods of high download activity and locate older forgotten files." >> "$OUTPUT_FILE"
echo "6. Address 'FILES WITH NO EXTENSION'. Rename them if you know their type, or use 'file' command individually to investigate." >> "$OUTPUT_FILE"
echo "7. Consider automating cleanup: Set up scripts or use tools to move files based on type or age periodically." >> "$OUTPUT_FILE"


echo ""
echo "Analysis complete! Results saved to $OUTPUT_FILE"
echo "Run 'less $OUTPUT_FILE' or 'cat $OUTPUT_FILE' to view the analysis."

# Note: chmod +x is removed as it's better practice to set permissions outside the script
# The user should run `chmod +x 01-inspect-downloads-folder.sh` once. 