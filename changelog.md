# Job Reports CLI

Added a new command-line interface for parsing and displaying job reports. This new tool provides both legacy text output and structured output using Glazed, allowing users to easily view and analyze job report data.

- Created `job-reports.go` with `JobReportsCommand` implementation
- Added `main.go` to set up the command-line interface
- Supports parsing multiple report files
- Allows users to choose between summary, job details, or all data
- Provides verbose output option
- Implements both legacy text output and structured Glazed output

# Lua Server Test Files

Added a set of Lua test files to demonstrate various server functionalities.

- Created `lua/` subdirectory with example Lua files
- Added `hello.lua` for basic string responses and query parameter handling
- Added `echo.lua` for echoing request details
- Added `calculator.lua` for simple arithmetic operations
- Added `counter.lua` to demonstrate maintaining state between requests

# Cross-Platform Extension Support

Added Firefox support to the Claude Intercept Extension while maintaining Chrome compatibility.

- Added WebExtension browser API polyfill for cross-browser compatibility
- Updated manifest.json with Firefox-specific settings
- Refactored background.js and popup.js to use browser API
- Updated build system to generate both Chrome and Firefox versions
- Improved popup UI and button handling

# Google Maps Places API Commands

Added new commands for interacting with the Google Places API:
- `maps places search`: Search for places using text queries and filters
- `maps places nearby`: Find places near a specific location
- `maps places details`: Get detailed information about a specific place

These commands provide a CLI interface to the Google Places API functionality, allowing users to search for and get information about places of interest.
