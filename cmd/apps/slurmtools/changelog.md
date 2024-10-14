# Enhanced log file processing script

Added flexibility and YAML output option to improve usability and integration with other tools.

- Added command-line flags for customizing the number of files (-f) and lines (-l) to display
- Implemented a YAML output option (-y) for structured data representation
- Updated script to use getopts for parsing command-line arguments
- Added a function to handle YAML output formatting

Further improved the script's usability and made YAML the default output format for better integration with other tools.

- Added a help flag (-h) to display usage information
- Made YAML the default output format
- Changed the text output flag to -t (previously -y was used for YAML)
- Updated the script to use the new command-line arguments
- Improved error handling by displaying usage information for invalid options

# snakemake html application

- Created a web-based Snakemake log viewer application
  - Implemented main.go with log parsing and web server functionality
  - Added index.html template for displaying job overview
  - Added job_details.html template for showing individual job details
  - Integrated HTMX for dynamic content loading
  - Implemented command-line flags for log file path, port, and host using Cobra

## Features

- Display overview of all Snakemake jobs in a table format
- Show detailed information for individual jobs on demand
- Provide summary statistics (total jobs, completed, in progress)
- Allow customization of log file path, server port, and host through command-line flags


# Add command-line flags using Cobra

Added command-line flags to customize the Snakemake log file path, HTTP port, and host.

- Added Cobra dependency for handling command-line flags
- Implemented `--log` flag to specify the Snakemake log file path (default: snakemake.log)
- Implemented `--port` flag to set the HTTP port (default: 6060)
- Implemented `--host` flag to set the host to bind the server to (default: localhost)
- Refactored main() function to use Cobra for command execution

# Snakemake Viewer CLI

## Added debug flag for troubleshooting ParseLog function
- Implemented a new --debug flag in the CLI tool
- Updated ParseLog function to accept a debug parameter
- Added debug logging statements throughout the ParseLog function
- Modified main.go to pass the debug flag to ParseLog

This update allows users to enable debug logging, which prints detailed information about the parsing process. This feature will help in troubleshooting issues with log parsing and understanding why the output might not be as expected.

# Improve Snakemake Viewer UI with Milligram CSS

Enhanced the user interface of the Snakemake Viewer for better readability and user experience.

- Added Milligram CSS for a clean and modern look
- Restructured the layout of index.html and job_details.html
- Improved table readability and consistency
- Added "View Details" button for each job using HTMX for smoother interactions
- Reorganized job details view for better information hierarchy

# Make Job Statistics Hidable

Improved the user interface by making the Job Statistics section collapsible.

- Added a toggle button to show/hide the Job Statistics section
- Job Statistics are now hidden by default for a cleaner initial view
- Implemented JavaScript functionality to handle the toggle action
- Updated CSS to style the toggle button consistently with the Milligram theme

## Token Type Refactoring

Improved code readability and maintainability by changing token types from integer enum to string constants.

- Changed `TokenType` in `tokenizer.go` from `int` to `string`
- Updated token type constants to use descriptive string values
- Modified `parser.go` to use the new string-based token types

# Update launch configuration working directory

Updated the VS Code launch configuration to set the working directory to the project root.

- Added "cwd": "${workspaceFolder}" to the launch configuration in .vscode/launch.json
- This ensures that the Snakemake Viewer CLI runs from the correct working directory

# Support for Multiple Log Files and Additional Columns

Enhanced the Snakemake Viewer CLI to support multiple input log files and added new columns for improved analysis.

- Modified the `logfile` parameter to `logfiles`, allowing multiple input files
- Added a `filename` column to identify the source log file for each entry
- Introduced a `duration_s` column to provide job duration in seconds
- Updated both legacy and structured output to include the new information
