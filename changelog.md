# Job Reports CLI

Added a new command-line interface for parsing and displaying job reports. This new tool provides both legacy text output and structured output using Glazed, allowing users to easily view and analyze job report data.

- Created `job-reports.go` with `JobReportsCommand` implementation
- Added `main.go` to set up the command-line interface
- Supports parsing multiple report files
- Allows users to choose between summary, job details, or all data
- Provides verbose output option
- Implements both legacy text output and structured Glazed output
