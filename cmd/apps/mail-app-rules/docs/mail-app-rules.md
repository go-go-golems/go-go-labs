# Mail App Rules - IMAP Email Processing Tool

## Overview

Mail App Rules is a flexible command-line tool for searching, filtering, and processing emails from IMAP servers. It offers two approaches:

1. **Rule-Based Processing** (`mail-rules`) - Define complex email filtering rules in YAML files
2. **Direct Command-Line Interface** (`fetch-mail`) - Query emails with simple command-line arguments

This tool is built to be powerful yet straightforward, allowing both casual users and power users to effectively manage and process their emails programmatically.

## Key Features

- Connect to any IMAP server
- Search emails with flexible criteria:
  - Date-based (since, before, within days)
  - Header-based (from, to, subject)
  - Content-based (text in body)
  - Flag-based (read, unread, flagged)
  - Size-based (larger/smaller than)
- Customize output format (JSON, text, table)
- Control which email fields to display
- Filter and process email content and MIME parts
- Support for environment variables for credentials

## Technical Architecture

### Core Components

1. **Command Interface** (`commands/`):
   - `mail_rules.go` - Processes rules defined in YAML files
   - `fetch_mail.go` - Builds rules from command-line arguments
   - `imap_layer.go` - Handles IMAP server connection settings

2. **Rule DSL Engine** (`dsl/`):
   - `types.go` - Core data structures for rules
   - `search.go` - Email search criteria implementation
   - `fetch.go` - Message fetching from IMAP server
   - `parser.go` - YAML rule parsing
   - `processor.go` - Rule execution logic
   - `output.go` - Output formatting and handling
   - `message.go` - Email message representation

3. **Application Entry Point** (`main.go`):
   - Command registration
   - Middleware setup
   - CLI handling with cobra

### Data Flow

1. **Rule Definition**: Either loaded from YAML file (`mail-rules`) or built from command-line arguments (`fetch-mail`)
2. **IMAP Connection**: Connect to server with provided credentials
3. **Rule Processing**:
   - Apply search criteria to find matching messages
   - Fetch message data based on output configuration
4. **Output Generation**:
   - Format results according to output settings
   - Display or redirect output as requested

### Key Interfaces

The `Rule` structure is central to the application and contains:

```go
type Rule struct {
    Name        string       // Rule name
    Description string       // Rule description
    Search      SearchConfig // Search criteria
    Output      OutputConfig // Output configuration
}
```

The `SearchConfig` defines search criteria, while `OutputConfig` controls which fields to include and how to format them.

## Detailed Command Reference

### Common Parameters (Both Commands)

**IMAP Connection Settings**:
- `--server` - IMAP server address
- `--port` - IMAP server port (default: 993)
- `--username` - IMAP username
- `--password` - IMAP password
- `--mailbox` - Mailbox to search in (default: "INBOX")
- `--insecure` - Skip TLS verification (default: false)

### mail-rules Command

The `mail-rules` command processes YAML rule files that define search criteria and output formatting.

**Usage**:
```bash
smailnail mail-rules --rule path/to/rule.yaml [IMAP options]
```

**Parameters**:
- `--rule` - Path to YAML rule file (required)
- `--concatenate-mime-parts` - Join all MIME parts into a single content string (default: true)

**YAML Rule Format**:
```yaml
name: "Example Rule"
description: "Find important emails from specific sender"
search:
  from: "important@example.com"
  subject_contains: "urgent"
  within_days: 7
  flags:
    not_has: ["seen"]
output:
  format: "table"
  limit: 20
  fields:
    - "subject"
    - "from"
    - "date"
    - mime_parts:
        show_content: true
        mode: "filter"
        types: ["text/plain"]
        max_length: 500
```

### fetch-mail Command

The `fetch-mail` command builds search rules directly from command-line arguments.

**Usage**:
```bash
smailnail fetch-mail [search options] [output options] [IMAP options]
```

**Search Options**:
- `--since` - Fetch emails since date (YYYY-MM-DD)
- `--before` - Fetch emails before date (YYYY-MM-DD)
- `--within-days` - Fetch emails within the last N days
- `--from` - Fetch emails from a specific sender
- `--to` - Fetch emails sent to a specific recipient
- `--subject` - Fetch emails with an exact subject match
- `--subject-contains` - Fetch emails with subject containing a string
- `--body-contains` - Fetch emails with body containing a string
- `--has-flags` - Fetch emails with specific flags (comma-separated)
- `--not-has-flags` - Fetch emails without specific flags (comma-separated)
- `--larger-than` - Fetch emails larger than size (e.g., '1M', '500K')
- `--smaller-than` - Fetch emails smaller than size (e.g., '1M', '500K')

**Output Options**:
- `--limit` - Maximum number of emails to fetch (default: 10)
- `--format` - Output format (json, text, table) (default: "text")
- `--include-subject` - Include email subject in output (default: true)
- `--include-from` - Include sender information in output (default: true)
- `--include-to` - Include recipient information in output (default: false)
- `--include-date` - Include email date in output (default: true)
- `--include-flags` - Include email flags in output (default: false)
- `--include-size` - Include email size in output (default: false)
- `--include-content` - Include email content in output (default: true)
- `--concatenate-mime-parts` - Concatenate all MIME parts into single content (default: true)
- `--content-max-length` - Maximum length of content to display (default: 1000)
- `--content-type` - MIME type to filter content (default: "text/plain")

## Tutorials and Examples

### Setting Up Environment Variables

To avoid typing credentials repeatedly or storing them in scripts, you can set environment variables:

```bash
export SMAILNAIL_USERNAME="your.email@example.com"
export SMAILNAIL_PASSWORD="your-password"
export SMAILNAIL_SERVER="imap.example.com"
```

The application will automatically use these variables when connecting.

### Basic Usage Examples

#### 1. Fetching Recent Unread Emails

```bash
smailnail fetch-mail \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --not-has-flags "seen" \
  --within-days 3 \
  --format table
```

This command will retrieve unread emails from the last 3 days and display them in a tabular format.

#### 2. Finding Emails from a Specific Sender with Attachments

```bash
smailnail fetch-mail \
  --from "important@company.com" \
  --larger-than "100K" \
  --content-type "application/*" \
  --include-flags true \
  --format json
```

This command searches for emails from "important@company.com" that are larger than 100KB (likely containing attachments) and outputs them in JSON format.

#### 3. Searching for Specific Content in Emails

```bash
smailnail fetch-mail \
  --body-contains "quarterly report" \
  --subject-contains "Q3" \
  --within-days 90 \
  --limit 5 \
  --content-max-length 200
```

This command finds emails containing "quarterly report" in the body and "Q3" in the subject that were received in the last 90 days. It limits results to 5 emails and truncates content display to 200 characters.

### Advanced Examples Using Rule Files

#### 1. Creating a Rule File for VIP Emails

Create a file named `vip.yaml`:

```yaml
name: "VIP Emails"
description: "Track emails from VIP senders that need immediate attention"
search:
  from:
    - "ceo@company.com"
    - "boss@company.com"
  subject_contains: "urgent"
  flags:
    not_has: ["seen"]
output:
  format: "table"
  limit: 10
  fields:
    - "subject"
    - "from" 
    - "date"
    - mime_parts:
        show_content: true
        mode: "text_only"
        max_length: 300
```

Run with:

```bash
smailnail mail-rules --rule vip.yaml
```

#### 2. Weekly Report Rule

Create a file named `weekly-report.yaml`:

```yaml
name: "Weekly Report Collector"
description: "Gather all weekly reports from the last 7 days"
search:
  subject_contains: "Weekly Report"
  within_days: 7
output:
  format: "json"
  fields:
    - "subject"
    - "from"
    - "date"
    - mime_parts:
        show_content: true
        mode: "filter"
        types: ["text/plain", "text/html"]
        max_length: 1000
```

Run and save the output:

```bash
smailnail mail-rules --rule weekly-report.yaml > weekly-reports.json
```

#### 3. Complex Filtering Rule

Create a file named `project-updates.yaml`:

```yaml
name: "Project Update Tracker"
description: "Track all project updates with specific requirements"
search:
  from: "updates@projects.company.com"
  subject_contains: "Project Update"
  size:
    larger_than: "50K"
    smaller_than: "2M"
  since: "2023-01-01"
  before: "2023-12-31"
  flags:
    has: ["flagged"]
output:
  format: "table"
  limit: 50
  fields:
    - "subject"
    - "date"
    - "size"
    - "flags"
    - mime_parts:
        show_content: true
        mode: "filter"
        types: ["text/plain"]
        max_length: 500
```

Run with:

```bash
smailnail mail-rules --rule project-updates.yaml
```

### Combining with Other Tools

#### Piping to jq for Advanced JSON Processing

```bash
smailnail fetch-mail --from "reports@company.com" --format json | jq '.[] | select(.size > 500000) | {subject, date, size}'
```

This fetches emails from reports@company.com, outputs as JSON, then uses jq to filter for large emails and display only selected fields.

#### Creating Daily Email Reports

Create a shell script called `daily-report.sh`:

```bash
#!/bin/bash
today=$(date +%Y-%m-%d)
echo "Daily Email Report for $today" > report.txt
echo "===========================" >> report.txt
echo "" >> report.txt

echo "1. Unread Priority Emails:" >> report.txt
smailnail fetch-mail --not-has-flags "seen" --has-flags "flagged" --format text >> report.txt

echo "" >> report.txt
echo "2. Messages from Management:" >> report.txt
smailnail fetch-mail --from "management@company.com" --within-days 1 --format text >> report.txt

echo "" >> report.txt
echo "3. Large Attachments:" >> report.txt
smailnail fetch-mail --larger-than "5M" --within-days 1 --include-size true --format text >> report.txt

cat report.txt
```

Make executable with `chmod +x daily-report.sh` and run daily.

## Debugging and Troubleshooting

### Common Issues

#### Authentication Errors

If you encounter authentication errors, check:
- Username and password are correct
- The server supports the authentication method
- If using Gmail, ensure "Less secure app access" is enabled or use an app-specific password

#### TLS/SSL Errors

If you encounter TLS errors:
- Verify the server address and port are correct
- Try using the `--insecure` flag (for testing only)
- Check if your network blocks the required ports

#### No Results Found

If no emails are returned:
- Check the mailbox name (case-sensitive, use `INBOX` for default)
- Verify search criteria aren't too restrictive
- Ensure date formats are correct (YYYY-MM-DD)

### Verbose Logging

For debugging, you can enable verbose logging:

```bash
GLAZED_LOG_LEVEL=debug smailnail fetch-mail ...
```

## Best Practices

1. **Security**:
   - Use environment variables for credentials
   - Consider using app-specific passwords for services like Gmail
   - Avoid storing credentials in rule files or scripts

2. **Performance**:
   - Use specific search criteria to limit results
   - Add a reasonable `--limit` value to avoid fetching too many emails
   - For large mailboxes, use date constraints (`--within-days` or `--since`)

3. **Rule Organization**:
   - Create separate rule files for different purposes
   - Include descriptive names and documentation
   - Version control your rule files

4. **Output Management**:
   - Use appropriate output formats for your needs (JSON for parsing, table for human reading)
   - Set reasonable content length limits
   - Consider redirecting output to files for later processing

## Conclusion

Mail App Rules provides a powerful and flexible way to interact with email programmatically. Whether you need simple ad-hoc email searches or complex rule-based processing, this tool offers an efficient solution for email management tasks.

The combination of the `mail-rules` command for robust rule-based processing and the `fetch-mail` command for quick searches gives users the best of both worlds - powerful capabilities without sacrificing ease of use.

By integrating this tool into your workflows, you can automate email processing, create custom reports, and build more sophisticated email-based applications. 