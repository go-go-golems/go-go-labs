# Quick Start Guide

This guide will help you get started with the IMAP DSL processor quickly.

## Prerequisites

- Go 1.18 or later
- Access to an IMAP email server
- Your IMAP server credentials

## Step 1: Build the Tool

```bash
# Clone the repository (if you haven't already)
git clone https://github.com/yourusername/go-go-labs.git
cd go-go-labs

# Build the application
go build -o imap-dsl ./cmd/apps/mail-app-rules
```

## Step 2: Create a Simple Rule File

Create a file named `my-rule.yaml` with the following content:

```yaml
name: "My First Rule"
description: "Find recent emails"
search:
  within_days: 7
output:
  format: text
  fields:
    - subject
    - from
    - date
```

## Step 3: Run the Tool

```bash
# Set your IMAP password as an environment variable (optional)
export IMAP_PASSWORD=yourpassword

# Run the tool
./imap-dsl \
  -rule my-rule.yaml \
  -server imap.example.com \
  -username your.email@example.com \
  -mailbox INBOX
```

If you don't set the IMAP_PASSWORD environment variable, you'll be prompted to enter your password.

## Step 4: Explore More Examples

Check out the example YAML files in the `examples/` directory:

- `recent-emails.yaml`: Find emails from the last 7 days
- `from-specific-sender.yaml`: Find emails from a specific sender
- `important-emails.yaml`: Find important emails from the last month
- `date-range-search.yaml`: Find emails within a specific date range
- `full-message-content.yaml`: Retrieve complete message content
- `detailed-example.yaml`: A comprehensive example with comments

## Step 5: Create Your Own Rules

Use the examples as templates to create your own custom rules. The YAML format is flexible and allows you to combine different search criteria and output formats.

## Troubleshooting

- **Connection issues**: Make sure your IMAP server address is correct and that your server allows IMAP access.
- **Authentication failures**: Double-check your username and password.
- **No results**: Verify that your search criteria match emails in your mailbox.
- **TLS errors**: If you're having TLS certificate issues, you can use the `-insecure` flag (not recommended for production use).

## Next Steps

- Read the full documentation in the `README.md` file
- Explore the `imap-dsl.md` specification for more details on the DSL syntax
- Contribute to the project by adding new features or fixing bugs 