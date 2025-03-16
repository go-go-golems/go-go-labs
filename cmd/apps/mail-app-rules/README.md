# IMAP DSL Processor

A command-line tool for searching and processing emails using a YAML-based Domain Specific Language (DSL).

## Overview

The IMAP DSL Processor allows you to define email search and display rules using a simple YAML syntax. This tool connects to an IMAP server, searches for emails matching your criteria, and displays the results in your preferred format.

## Features

- Search emails using various criteria (date ranges, sender, etc.)
- Display email fields in different formats (text, JSON, table)
- Simple YAML-based configuration
- Secure connection to IMAP servers

## Quick Start

For a quick introduction, check out the [Quick Start Guide](examples/QUICK-START.md) in the examples directory.

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/go-go-labs.git
cd go-go-labs

# Build the application
go build -o imap-dsl ./cmd/apps/mail-app-rules
```

## Usage

```bash
./imap-dsl -rule examples/recent-emails.yaml -server imap.example.com -username user@example.com -password yourpassword
```

### Command Line Options

- `-rule`: Path to YAML rule file (required)
- `-server`: IMAP server address (required)
- `-port`: IMAP server port (default: 993)
- `-username`: IMAP username (required)
- `-password`: IMAP password (required if not set via IMAP_PASSWORD env var)
- `-mailbox`: Mailbox to search in (default: "INBOX")
- `-insecure`: Skip TLS verification (default: false)

You can also set your password via environment variable:

```bash
export IMAP_PASSWORD=yourpassword
./imap-dsl -rule examples/recent-emails.yaml -server imap.example.com -username user@example.com
```

## YAML Rule Format

The YAML rule files define what emails to search for and how to display them.

### Basic Structure

```yaml
name: "Rule Name"
description: "Description of what this rule does"
search:
  # Search criteria go here
output:
  # Output format and fields go here
```

### Search Criteria

The following search criteria are supported:

- `since`: Emails received since this date (format: YYYY-MM-DD)
- `before`: Emails received before this date (format: YYYY-MM-DD)
- `on`: Emails received on this specific date (format: YYYY-MM-DD)
- `within_days`: Emails received within the last N days
- `from`: Emails from a specific sender (partial match)

### Output Format

The output section defines how to display the results:

```yaml
output:
  format: text  # Options: text, json, table
  fields:
    - subject
    - from
    - date
    - flags
    - body:
        type: text/plain
        max_length: 500
```

Available fields:
- `uid`: Message UID
- `subject`: Email subject
- `from`: Sender
- `to`: Recipients
- `date`: Date received
- `flags`: Email flags
- `size`: Message size in bytes
- `body`: Message body (can specify type and max_length)

## Examples

The `examples/` directory contains sample YAML rule files and helper scripts:

- `recent-emails.yaml`: Display recent emails
- `from-specific-sender.yaml`: Find emails from a specific sender
- `important-emails.yaml`: Find emails with important flags
- `date-range-search.yaml`: Find emails within a specific date range
- `full-message-content.yaml`: Retrieve complete message content
- `complex-search.yaml`: Combine multiple search criteria
- `detailed-example.yaml`: A comprehensive example with comments
- `run-example.sh`: Shell script to run the examples

To run an example:

```bash
# Set your IMAP server details
export IMAP_SERVER=imap.example.com
export IMAP_USERNAME=your.email@example.com
export IMAP_PASSWORD=yourpassword

# Run the example script
cd cmd/apps/mail-app-rules/examples
./run-example.sh
```

## License

MIT 