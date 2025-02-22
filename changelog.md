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

# Debug Logging for Places API Commands

Added debug logging for Places API commands:
- Added detailed logging of query parameters for all commands
- Added result logging for search, nearby, and details operations
- Using zerolog for structured logging output

The debug logs help track the execution flow and parameters of Places API operations.

# Places API Implementation

Implemented the actual Google Places API functionality:
- Added text search with support for location, radius, and type filters
- Added place details retrieval with comprehensive information display
- Added nearby search with location-based filtering
- Improved error handling and input validation
- Added proper type conversion for place types
- Fixed context handling between root and subcommands

# Structured Output for Places API Commands

Converted Places API commands to use Glazed structured output:
- Implemented GlazeCommand interface for all commands
- Added structured data output using types.Row
- Organized place data into consistent fields
- Improved data formatting for opening hours
- Removed direct printing in favor of structured output
- Added support for Glazed output formatting options

# Glazed Settings and Parameters

Enhanced Places API commands with Glazed settings and parameters:
- Added settings structs with glazed.parameter tags for all commands
- Added proper parameter definitions with types, help text, and defaults
- Implemented parameter initialization from parsed layers
- Improved command help text and documentation
- Added validation for required parameters
- Unified parameter handling across all commands

# Added Glazed Parameter Layer

Added Glazed parameter layer to all Places API commands:
- Added support for standard Glazed output formatting options
- Enabled table, JSON, YAML, and other output formats
- Added column selection and filtering capabilities
- Added output formatting options (width, headers, etc.)
- Unified output handling across all commands
- Improved command-line user experience

# Added Google Maps Directions Command

Added new command for getting directions between locations:
- Added `maps directions` command with comprehensive routing options
- Support for different travel modes (driving, walking, bicycling, transit)
- Support for waypoints and route preferences
- Avoid options for tolls, highways, and ferries
- Metric and imperial unit support
- Structured output with route summaries and step-by-step instructions
- Detailed distance, duration, and location information
- Full Glazed parameter layer and output formatting support
