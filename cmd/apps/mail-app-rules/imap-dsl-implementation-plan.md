
# Implementation Plan for IMAP DSL Processor (Initial Version)

## Overview
We'll implement a minimal version of the IMAP DSL processor focusing on:
- Basic search functionality (date-based and sender-based)
- Output formatting for a subset of fields
- No action functionality in this phase

## Step 1: Project Structure and Setup

Create the following files:
```
cmd/apps/mail-app-rules/
├── dsl/
│   ├── types.go         # Core data structures
│   ├── parser.go        # YAML parsing logic
│   ├── search.go        # Search criteria implementation
│   ├── output.go        # Output formatting
│   └── processor.go     # Main processing logic
├── cmd_process_rule.go  # Command implementation
```

## Step 2: Define Core Data Structures

In `dsl/types.go`, define the necessary structs:

```go
// Rule represents a complete IMAP DSL rule
type Rule struct {
    Name        string       `yaml:"name"`
    Description string       `yaml:"description"`
    Search      SearchConfig `yaml:"search"`
    Output      OutputConfig `yaml:"output"`
}

// SearchConfig defines search criteria
type SearchConfig struct {
    Since      string `yaml:"since"`
    Before     string `yaml:"before"`
    On         string `yaml:"on"`
    WithinDays int    `yaml:"within_days"`
    From       string `yaml:"from"`
}

// OutputConfig defines output formatting
type OutputConfig struct {
    Format string   `yaml:"format"` // json, text, table
    Fields []Field  `yaml:"fields"`
}

// Field represents an output field, which can be a simple string or complex field
type Field struct {
    Name string
    Body *BodyField
    // More field types will be added later
}

// BodyField represents body output configuration
type BodyField struct {
    Type      string `yaml:"type"`
    MaxLength int    `yaml:"max_length"`
}
```

## Step 3: Implement YAML Parser

In `dsl/parser.go`:

```go
// ParseRuleFile parses a YAML rule file into a Rule struct
func ParseRuleFile(filename string) (*Rule, error) {
    // Read file
    // Parse YAML into Rule struct
    // Validate basic requirements
}

// ParseRuleString parses a YAML string into a Rule struct
func ParseRuleString(yamlStr string) (*Rule, error) {
    // Similar to ParseRuleFile but works with string
}
```

## Step 4: Implement Search Criteria Processing

In `dsl/search.go`:

```go
// BuildSearchCriteria converts SearchConfig to imap.SearchCriteria
func BuildSearchCriteria(config SearchConfig) (*imap.SearchCriteria, error) {
    criteria := &imap.SearchCriteria{}
    
    // Process date criteria
    if config.Since != "" {
        // Parse and set Since
    }
    if config.Before != "" {
        // Parse and set Before
    }
    if config.On != "" {
        // Parse and set date range for specific day
    }
    if config.WithinDays > 0 {
        // Calculate and set Since based on WithinDays
    }
    
    // Process From criteria
    if config.From != "" {
        criteria.From = config.From
    }
    
    return criteria, nil
}
```

## Step 5: Implement Output Formatting

In `dsl/output.go`:

```go
// BuildFetchOptions converts OutputConfig to imap.FetchOptions
func BuildFetchOptions(config OutputConfig) (*imap.FetchOptions, error) {
    options := &imap.FetchOptions{}
    
    // Process fields
    for _, field := range config.Fields {
        switch field.Name {
        case "uid":
            options.UID = true
        case "envelope":
            options.Envelope = true
        case "flags":
            options.Flags = true
        case "body":
            if field.Body != nil {
                // Configure body section fetch
                options.BodySection = &imap.FetchItemBodySection{
                    Specifier: "TEXT",
                    Peek: true,
                }
            }
        }
    }
    
    return options, nil
}

// FormatOutput formats message data according to OutputConfig
func FormatOutput(msgData *imapclient.FetchMessageBuffer, config OutputConfig) (string, error) {
    // Format output based on config.Format (json, text, table)
    // Include requested fields only
}
```

## Step 6: Implement Main Processor

In `dsl/processor.go`:

```go
// ProcessRule executes an IMAP rule
func ProcessRule(client *imapclient.Client, rule *Rule) error {
    // 1. Build search criteria
    criteria, err := BuildSearchCriteria(rule.Search)
    if err != nil {
        return err
    }
    
    // 2. Execute search
    searchCmd := client.Search(criteria, nil)
    searchData, err := searchCmd.Wait()
    if err != nil {
        return err
    }
    
    // 3. Check if we have results
    if len(searchData.Messages) == 0 {
        return nil // No messages found
    }
    
    // 4. Create sequence set from results
    var seqSet imap.SeqSet
    seqSet.AddNum(searchData.Messages...)
    
    // 5. Build fetch options
    fetchOptions, err := BuildFetchOptions(rule.Output)
    if err != nil {
        return err
    }
    
    // 6. Fetch messages
    fetchCmd := client.Fetch(seqSet, fetchOptions)
    defer fetchCmd.Close()
    
    // 7. Process and output messages
    var results []string
    for {
        msg := fetchCmd.Next()
        if msg == nil {
            break
        }
        
        msgData, err := msg.Collect()
        if err != nil {
            return err
        }
        
        output, err := FormatOutput(msgData, rule.Output)
        if err != nil {
            return err
        }
        
        results = append(results, output)
    }
    
    // 8. Print final output
    for _, result := range results {
        fmt.Println(result)
    }
    
    return nil
}
```

## Step 7: Implement Command Line Interface

In `cmd_process_rule.go`:

```go
var processRuleCmd = &cobra.Command{
    Use:   "process-rule [rule-file]",
    Short: "Process an IMAP rule file",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Parse rule file
        rule, err := dsl.ParseRuleFile(args[0])
        if err != nil {
            return err
        }
        
        // 2. Connect to IMAP server
        options := &imapclient.Options{
            WordDecoder: &mime.WordDecoder{},
        }
        
        client, err := imapclient.DialTLS(fmt.Sprintf("%s:%d", server, port), options)
        if err != nil {
            return fmt.Errorf("failed to connect: %v", err)
        }
        defer client.Close()
        
        // 3. Login
        if err := client.Login(username, password).Wait(); err != nil {
            return fmt.Errorf("failed to login: %v", err)
        }
        
        // 4. Select INBOX (or configurable mailbox)
        if _, err := client.Select("INBOX", nil).Wait(); err != nil {
            return fmt.Errorf("failed to select inbox: %v", err)
        }
        
        // 5. Process rule
        if err := dsl.ProcessRule(client, rule); err != nil {
            return err
        }
        
        return nil
    },
}
```

## Step 8: Testing

1. Create test rule files:
   ```yaml
   # test_recent_from.yaml
   name: "Recent From Rule"
   description: "Find recent emails from a specific sender"
   
   search:
     from: "important@example.com"
     within_days: 7
   
   output:
     format: text
     fields:
       - subject
       - from
       - date
   ```