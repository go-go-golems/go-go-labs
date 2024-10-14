---
Title: Basic Usage of Snakemake Viewer CLI
Slug: basic-usage
Short: Learn how to use the basic features of Snakemake Viewer CLI
Topics:
  - snakemake
  - cli
Commands:
  - view
Flags:
  - logfile
  - verbose
IsTopLevel: true
ShowPerDefault: true
SectionType: Example
---

# Basic Usage of Snakemake Viewer CLI

To use the Snakemake Viewer CLI, follow these basic steps:

1. Run the `view` command with the required `logfile` parameter:

   ```
   snakemake-viewer-cli view --logfile path/to/your/snakemake.log
   ```

   This will display a summary of the Snakemake log, including total jobs, completed jobs, and in-progress jobs.

2. For more detailed information, use the `verbose` flag:

   ```
   snakemake-viewer-cli view --logfile path/to/your/snakemake.log --verbose
   ```

   This will show additional information about each job, including start time, end time, duration, and resource usage.

3. By default, the output is in a structured format. To use the legacy text output, add the `legacy` flag:

   ```
   snakemake-viewer-cli view --logfile path/to/your/snakemake.log --legacy
   ```

These examples cover the basic usage of the Snakemake Viewer CLI. Explore more options and flags to customize the output according to your needs.
