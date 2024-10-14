---
Title: Snakemake Viewer CLI - Usage and Examples
Slug: usage
Short: The Snakemake Viewer CLI is a tool for analyzing and visualizing Snakemake log files.
Topics:
  - snakemake
  - cli
  - data analysis
Commands:
  - view
Flags:
  - logfiles
  - verbose
  - data
  - output
IsTopLevel: true
ShowPerDefault: true
SectionType: Example
---

## Basic Usage

To view parsed Snakemake log information:

```bash
snakemake-viewer-cli view --logfiles path/to/your/snakemake.log
```

## Selecting Data to Display

Use the `--data` flag to specify which information to show:

```bash
# Show only rule summaries
snakemake-viewer-cli view --logfiles snakemake.log --data rules

# Show only job information
snakemake-viewer-cli view --logfiles snakemake.log --data jobs

# Show summary information
snakemake-viewer-cli view --logfiles snakemake.log --data summary

# Show all available information
snakemake-viewer-cli view --logfiles snakemake.log --data all
```

## Processing Multiple Log Files

Analyze multiple log files in a single command:

```bash
snakemake-viewer-cli view --logfiles file1.log,file2.log,file3.log
```

Or use shell expansion:

```bash
snakemake-viewer-cli view --logfiles $(find /path/to/logs -name "*.log" | head -5)
```

## Output Formats

The CLI supports various output formats using Glazed's output flags:

```bash
# Output as CSV
snakemake-viewer-cli view --logfiles snakemake.log --output csv

# Output as YAML
snakemake-viewer-cli view --logfiles snakemake.log --output yaml

# Output as JSON
snakemake-viewer-cli view --logfiles snakemake.log --output json
```

## Filtering and Selecting Fields

Use Glazed's field selection and filtering capabilities:

```bash
# Select specific fields
snakemake-viewer-cli view --logfiles snakemake.log --fields rule,duration,status

# Filter out specific fields
snakemake-viewer-cli view --logfiles snakemake.log --filter filename
```

## Sorting and Limiting Results

Sort and limit the output:

```bash
# Sort by duration (descending)
snakemake-viewer-cli view --logfiles snakemake.log --sort-by=-duration

# Limit to the first 10 results
snakemake-viewer-cli view --logfiles snakemake.log --glazed-limit 10
```

## Advanced Data Manipulation

Utilize Glazed's advanced data manipulation features:

```bash
# Rename fields
snakemake-viewer-cli view --logfiles snakemake.log --rename rule:task_name,duration:execution_time

# Add a constant field
snakemake-viewer-cli view --logfiles snakemake.log --add-fields pipeline_version:1.2.3

# Apply a JQ query
snakemake-viewer-cli view --logfiles snakemake.log --jq '.[] | select(.status == "Completed")'
```

## Verbose Output

For more detailed information, use the `--verbose` flag:

```bash
snakemake-viewer-cli view --logfiles snakemake.log --verbose
```

## Combining Features

Combine multiple features for powerful analysis:

```bash
snakemake-viewer-cli view \
  --logfiles $(find /path/to/logs -name "*.log") \
  --data jobs \
  --output csv \
  --fields rule,duration,status \
  --sort-by=-duration \
  --glazed-limit 20 \
  --verbose
```

This command processes all log files in a directory, shows job information, outputs as CSV, selects specific fields, sorts by duration (descending), limits to 20 results, and provides verbose output.

For more information on Glazed's output formatting and data manipulation flags, refer to the [Glazed documentation](https://github.com/go-go-golems/glazed).

Explore these features to get the most out of your Snakemake log analysis!
