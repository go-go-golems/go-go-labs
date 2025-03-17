# IMAP and Email Message Examples

This directory contains a collection of examples demonstrating how to use the [go-imap](https://github.com/emersion/go-imap) and [go-message](https://github.com/emersion/go-message) libraries to fetch and process email messages over IMAP.

## Building the Examples

```bash
go build -o imap-examples
```

## Available Commands

The examples are organized as subcommands:

### Basic IMAP Operations

- **connect**: Connect to an IMAP server, authenticate, and select a mailbox
  ```bash
  ./imap-examples connect --server imap.example.com --username user@example.com --password secret
  ```

- **fetch-metadata**: Fetch message metadata (envelope, flags, size)
  ```bash
  ./imap-examples fetch-metadata --server imap.example.com --username user@example.com --password secret --num 5
  ```

### Message Structure and Content

- **fetch-structure**: Fetch and display the MIME structure of a message
  ```bash
  ./imap-examples fetch-structure --server imap.example.com --username user@example.com --password secret --uid 123
  ```

- **fetch-content**: Fetch the full content of a message and parse it
  ```bash
  ./imap-examples fetch-content --server imap.example.com --username user@example.com --password secret --uid 123
  ```

- **fetch-parts**: Fetch specific parts of a message based on the structure
  ```bash
  ./imap-examples fetch-parts --server imap.example.com --username user@example.com --password secret --uid 123 --part 1.2 --header
  ```

### Advanced Message Processing

- **stream-large**: Stream large messages efficiently, saving attachments directly to disk
  ```bash
  ./imap-examples stream-large --server imap.example.com --username user@example.com --password secret --uid 123 --output attachments
  ```

- **handle-alternatives**: Handle multipart/alternative messages with HTML and plain text versions
  ```bash
  ./imap-examples handle-alternatives --server imap.example.com --username user@example.com --password secret --uid 123 --prefer-html --save
  ```

- **handle-embedded-images**: Extract embedded images from HTML emails and process the HTML
  ```bash
  ./imap-examples handle-embedded-images --server imap.example.com --username user@example.com --password secret --uid 123
  ```

## Common Flags

All commands support the following flags:

- `--server` (required): IMAP server address
- `--port`: IMAP server port (default: 993)
- `--username` (required): IMAP username
- `--password` (required): IMAP password
- `--mailbox`: Mailbox to select (default: INBOX)
- `--ssl`: Use SSL/TLS connection (default: true)
- `--uid`: Message UID to fetch (when applicable)

## Examples

### Fetching and Displaying Message Structure

```bash
# First, connect and list the last 5 messages to find a UID
./imap-examples fetch-metadata --server imap.example.com --username user@example.com --password secret --num 5

# Then, fetch the structure of a specific message
./imap-examples fetch-structure --server imap.example.com --username user@example.com --password secret --uid 123
```

### Downloading Attachments from a Message

```bash
./imap-examples stream-large --server imap.example.com --username user@example.com --password secret --uid 123 --output my_attachments --save-attachments
```

### Viewing HTML Email with Embedded Images

```bash
./imap-examples handle-embedded-images --server imap.example.com --username user@example.com --password secret --uid 123 --process-html
```

## Security Note

These examples require your IMAP password to be provided on the command line, which is not secure for production use. In a real application, you should use a more secure authentication method or a password prompt. 