# IMAP DSL Specification

This document specifies a YAML-based Domain Specific Language (DSL) for interacting with IMAP servers. The DSL allows you to define search criteria, specify output formats, and perform actions on email messages.

## Basic Structure

Every IMAP DSL rule consists of three main sections:
```yaml
search:  # Define search criteria
  # search parameters

output:  # Specify what to display/extract
  # output parameters

actions: # Define actions to take on matched messages
  # action parameters
```

## Search Criteria

The `search` section supports various criteria for finding messages:

### Date-based Search
```yaml
search:
  since: "2024-01-01"     # Messages since date
  before: "2024-03-01"    # Messages before date
  on: "2024-02-14"        # Messages on specific date
  within_days: 7          # Messages within last N days
```

### Header-based Search
```yaml
search:
  from: "sender@example.com"
  to: "recipient@example.com"
  cc: "cc@example.com"
  bcc: "bcc@example.com"
  subject: "Important Meeting"
  subject_contains: "Report"
  header:
    name: "Message-ID"
    value: "<123@example.com>"
```

### Content-based Search
```yaml
search:
  body_contains: "specific text"
  text: "search anywhere"  # Searches in headers and body
```

### Flag-based Search
```yaml
search:
  flags:
    has:
      - seen
      - flagged
    not_has:
      - deleted
      - draft
```

### Size-based Search
```yaml
search:
  size:
    larger_than: 1M    # Supports B, K, M, G units
    smaller_than: 5M
```

### Complex Conditions
```yaml
search:
  operator: and  # or, not
  conditions:
    - from: "team@company.com"
    - subject_contains: "Report"
    - operator: or
      conditions:
        - flags:
            has: ["urgent"]
        - subject_contains: "URGENT"
```

## Output Format

The `output` section defines what information to extract and how to format it:

### Basic Fields
```yaml
output:
  format: json  # json, text, table
  limit: 10     # Maximum number of messages to return (optional)
  fields:
    - uid
    - subject
    - from
    - to
    - date
    - size
    - flags
```

### Message Body Options
```yaml
output:
  fields:
    - body:
        type: "text/plain"  # or text/html
        max_length: 1000
        strip_quotes: true
```

### MIME Parts Listing
```yaml
output:
  fields:
    - mime_parts:
        list_only: true  # Just list content types without content
        mode: "full"     # Options: "full", "text_only", "filter"
        types:           # Required when mode is "filter"
          - "image/*"    # Supports wildcards
          - "application/pdf"
```

### Header Selection
```yaml
output:
  fields:
    - headers:
        include:
          - "Message-ID"
          - "In-Reply-To"
          - "References"
```

### Attachment Handling
```yaml
output:
  fields:
    - attachments:
        list: true           # Just list attachments
        download: false      # Whether to download
        save_path: "./attachments"
        types:              # Filter by mime types
          - "application/pdf"
          - "image/*"
```

### Message Body and MIME Parts Options
```yaml
output:
  fields:
    # Body field
    - body:
        type: "text/plain"  # or text/html
        max_length: 1000    # Maximum length of content to return
        min_length: 10      # Minimum length of content to return
    
    # MIME parts field
    - mime_parts:
        mode: "full"        # Options: "full", "text_only", "filter"
        types:              # Required when mode is "filter"
          - "image/*"       # Supports wildcards
          - "application/pdf"
        show_types: true    # Whether to show MIME types (default true)
        show_content: true  # Whether to show content (default false)
        max_length: 1000    # Maximum length of content to return
        min_length: 10      # Minimum length of content to return
```

## Actions

The `actions` section defines operations to perform on matched messages:

### Flag Operations
```yaml
actions:
  flags:
    add:
      - seen
      - flagged
    remove:
      - draft
```

### Move/Copy Operations
```yaml
actions:
  move_to: "Archive/2024"
  copy_to: "Backup"
```

### Delete Operation
```yaml
actions:
  delete: true  # Permanently delete
  # or
  delete:
    trash: true  # Move to trash instead
```

### Export Operations
```yaml
actions:
  export:
    format: "eml"          # eml, mbox
    directory: "./exports"
    filename_template: "{{date}}-{{subject}}"
```

## Complete Examples

### Example 1: Process Newsletter
```yaml
name: "Process Newsletters"
description: "Move newsletters to appropriate folder and mark as read"

search:
  from: "newsletter@company.com"
  flags:
    not_has: ["seen"]
  within_days: 7

output:
  format: text
  fields:
    - subject
    - date
    - from

actions:
  flags:
    add: ["seen"]
  move_to: "Newsletters"
```

### Example 2: Archive Large Old Emails
```yaml
name: "Archive Large Emails"
description: "Move large old emails to archive"

search:
  before: "2023-12-31"
  size:
    larger_than: 10M
  flags:
    has: ["seen"]
    not_has: ["flagged"]

output:
  format: json
  fields:
    - uid
    - subject
    - size
    - date

actions:
  move_to: "Archive/Large"
  export:
    format: "eml"
    directory: "./large-emails"
```

### Example 3: Urgent Customer Emails
```yaml
name: "Urgent Customer Support"
description: "Flag urgent customer support emails"

search:
  operator: and
  conditions:
    - from: "*@customers.com"
    - operator: or
      conditions:
        - subject_contains: "urgent"
        - subject_contains: "ASAP"
        - subject_contains: "emergency"
    - flags:
        not_has: ["seen"]

output:
  format: table
  fields:
    - date
    - from
    - subject
    - body:
        type: "text/plain"
        max_length: 200

actions:
  flags:
    add: ["flagged", "urgent"]
  copy_to: "Urgent-Support"
```

## Error Handling

The DSL processor will validate the YAML structure and provide clear error messages for:
- Invalid search criteria
- Unsupported output formats or fields
- Invalid action combinations
- Missing required fields
- Permission issues

## Implementation Notes

1. Date formats should follow RFC3339 or ISO8601
2. Size units: B (bytes), K (kilobytes), M (megabytes), G (gigabytes)
3. All actions are executed in the order specified
4. Search criteria are combined using AND logic unless specified otherwise
5. Flag names are case-insensitive 

## Library Mapping

This section explains how the DSL constructs map to the go-imap/v2 library functions and types.

### Search Criteria Mapping

The `search` section maps to `imap.SearchCriteria` and is used with the following client methods:

```go
// Basic search
client.Search(criteria *imap.SearchCriteria, options *imap.SearchOptions) *SearchCommand

// UID-based search
client.UIDSearch(criteria *imap.SearchCriteria, options *imap.SearchOptions) *SearchCommand
```

DSL to Go type mappings for search criteria:

```yaml
# DSL:
search:
  since: "2024-01-01"
  from: "sender@example.com"
  flags:
    has: ["seen"]
  size:
    larger_than: 1M
```

```go
// Go equivalent:
criteria := &imap.SearchCriteria{
    Since: time.Parse(time.RFC3339, "2024-01-01T00:00:00Z"),
    From: "sender@example.com",
    WithFlags: []imap.Flag{"\\Seen"},
    SizeLimit: &imap.SearchNumber{Min: 1024*1024}, // 1M
}
```

### Output Format Mapping

The `output` section maps to `imap.FetchOptions` and is used with:

```go
client.Fetch(numSet imap.NumSet, options *imap.FetchOptions) *FetchCommand
```

DSL to Go type mappings for fetch options:

```yaml
# DSL:
output:
  fields:
    - uid
    - envelope
    - body:
        type: "text/plain"
    - flags
```

```go
// Go equivalent:
fetchOptions := &imap.FetchOptions{
    UID: true,
    Envelope: true,
    BodySection: &imap.FetchItemBodySection{
        Specifier: "TEXT",
        Peek: true,
    },
    Flags: true,
}

// Results are accessed through:
msgData, err := msg.Collect()
if err != nil {
    return err
}
if msgData.Envelope != nil {
    // Access envelope data
}
if msgData.Flags != nil {
    // Access flags
}
```

### Actions Mapping

The `actions` section maps to various client methods:

#### Flag Operations
```yaml
# DSL:
actions:
  flags:
    add: ["seen", "flagged"]
```

```go
// Go equivalent:
store := &imap.StoreFlags{
    Op: imap.StoreFlagsAdd,
    Flags: []imap.Flag{"\\Seen", "\\Flagged"},
}
client.Store(numSet, store, nil)
```

#### Move/Copy Operations
```yaml
# DSL:
actions:
  move_to: "Archive"
  copy_to: "Backup"
```

```go
// Go equivalent:
// Move
moveCmd := client.Move(numSet, "Archive")
if err := moveCmd.Wait(); err != nil {
    return err
}

// Copy
copyCmd := client.Copy(numSet, "Backup")
if err := copyCmd.Wait(); err != nil {
    return err
}
```

### Message Selection Mapping

Message selection uses `imap.SeqSet` or `imap.UIDSet`:

```go
// For sequence numbers
var seqSet imap.SeqSet
seqSet.AddNum(1, 2, 3)
// or range
seqSet.AddRange(1, 10)

// For UIDs
var uidSet imap.UIDSet
uidSet.AddNum(100, 101, 102)
// or range
uidSet.AddRange(100, 200)
```

### Complex Operations Example

Here's how a complete DSL rule maps to Go code:

```yaml
# DSL:
search:
  from: "newsletter@company.com"
  flags:
    not_has: ["seen"]
  within_days: 7

output:
  fields:
    - subject
    - from
    - flags

actions:
  flags:
    add: ["seen"]
  move_to: "Newsletters"
```

```go
// Go equivalent:
func processRule(client *imapclient.Client) error {
    // 1. Build search criteria
    criteria := &imap.SearchCriteria{
        From: "newsletter@company.com",
        WithoutFlags: []imap.Flag{"\\Seen"},
        Since: time.Now().AddDate(0, 0, -7),
    }
    
    // 2. Execute search
    searchCmd := client.Search(criteria, nil)
    searchData, err := searchCmd.Wait()
    if err != nil {
        return err
    }
    
    // 3. Create sequence set from results
    var seqSet imap.SeqSet
    seqSet.AddNum(searchData.Messages...)
    
    // 4. Fetch required fields
    fetchOptions := &imap.FetchOptions{
        Envelope: true,
        Flags: true,
    }
    
    fetchCmd := client.Fetch(seqSet, fetchOptions)
    defer fetchCmd.Close()
    
    // 5. Process messages and apply actions
    for {
        msg := fetchCmd.Next()
        if msg == nil {
            break
        }
        
        msgData, err := msg.Collect()
        if err != nil {
            return err
        }
        
        // Apply flags
        store := &imap.StoreFlags{
            Op: imap.StoreFlagsAdd,
            Flags: []imap.Flag{"\\Seen"},
        }
        if err := client.Store(seqSet, store, nil).Wait(); err != nil {
            return err
        }
        
        // Move message
        if err := client.Move(seqSet, "Newsletters").Wait(); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Error Handling and Validation

The DSL processor should implement validation for:

1. Search Criteria:
   - Valid date formats
   - Valid flag names (against IMAP standard flags)
   - Valid size units and values

2. Output Format:
   - Valid field names
   - Valid MIME types for body sections
   - Valid format options (json/text/table)

3. Actions:
   - Valid flag combinations
   - Mailbox existence for move/copy operations
   - Write permissions for target mailboxes

Each operation should be wrapped with appropriate error handling:

```go
func executeAction(cmd *imapclient.Command) error {
    if err := cmd.Wait(); err != nil {
        switch err := err.(type) {
        case *imap.Error:
            // Handle IMAP protocol errors
            return fmt.Errorf("IMAP error: %v", err)
        default:
            // Handle other errors
            return fmt.Errorf("operation failed: %v", err)
        }
    }
    return nil
}
``` 