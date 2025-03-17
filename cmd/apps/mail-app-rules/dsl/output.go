package dsl

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

// OutputMessages formats and prints a list of email messages
func OutputMessages(messages []*EmailMessage, config OutputConfig) error {
	for i, msg := range messages {
		output, err := FormatOutput(msg, config)
		if err != nil {
			return fmt.Errorf("failed to format message %d: %w", i+1, err)
		}

		if i > 0 {
			fmt.Println("----------------------------------------")
		}
		fmt.Println(output)
	}

	fmt.Printf("\nFound %d message(s) matching the criteria\n", len(messages))
	return nil
}

// FormatOutput formats message data according to OutputConfig
func FormatOutput(msg *EmailMessage, config OutputConfig) (string, error) {
	switch config.Format {
	case "json":
		return formatOutputJSON(msg, config)
	case "table":
		return formatOutputTable(msg, config)
	default:
		// Default to text format
		return formatOutputText(msg, config)
	}
}

// MimePart represents a single MIME part in the message
type MimePart struct {
	Children    []MimePart
	Type        string
	Subtype     string
	Disposition string
	Encoding    string
	Size        uint32
	Content     string
	Filename    string
	Charset     string
}

func findBodySection(bodySections []imapclient.FetchBodySectionBuffer, specifier imap.PartSpecifier) []byte {
	for _, section := range bodySections {
		if section.Section.Specifier == specifier {
			return section.Bytes
		}
	}
	return nil
}

// formatOutputJSON formats message data as JSON
func formatOutputJSON(msg *EmailMessage, config OutputConfig) (string, error) {
	// Create a map to hold the output data
	output := make(map[string]interface{})

	// Process each field
	for _, fieldInterface := range config.Fields {
		field, ok := fieldInterface.(Field)
		if !ok {
			// Skip fields that couldn't be properly parsed
			continue
		}

		switch field.Name {
		case "uid":
			output["uid"] = msg.UID
		case "subject":
			if msg.Envelope != nil {
				output["subject"] = msg.Envelope.Subject
			}
		case "from":
			if msg.Envelope != nil && len(msg.Envelope.From) > 0 {
				output["from"] = msg.Envelope.From
			}
		case "to":
			if msg.Envelope != nil && len(msg.Envelope.To) > 0 {
				output["to"] = msg.Envelope.To
			}
		case "date":
			if msg.Envelope != nil {
				output["date"] = msg.Envelope.Date.Format(time.RFC3339)
			}
		case "flags":
			output["flags"] = msg.Flags
		case "size":
			output["size"] = msg.Size
		case "mime_parts":
			if len(msg.MimeParts) > 0 {
				output["mime_parts"] = msg.MimeParts
			}
		}
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonData), nil
}

// formatOutputText formats message data as plain text
func formatOutputText(msg *EmailMessage, config OutputConfig) (string, error) {
	var sb strings.Builder

	// Process each field
	for _, fieldInterface := range config.Fields {
		field, ok := fieldInterface.(Field)
		if !ok {
			// Skip fields that couldn't be properly parsed
			continue
		}

		switch field.Name {
		case "uid":
			sb.WriteString(fmt.Sprintf("UID: %d\n", msg.UID))
		case "subject":
			if msg.Envelope != nil {
				sb.WriteString(fmt.Sprintf("Subject: %s\n", msg.Envelope.Subject))
			}
		case "from":
			if msg.Envelope != nil && len(msg.Envelope.From) > 0 {
				sb.WriteString(fmt.Sprintf("From: %s\n", formatEmailAddress(msg.Envelope.From[0])))
			}
		case "to":
			if msg.Envelope != nil && len(msg.Envelope.To) > 0 {
				sb.WriteString("To: ")
				for i, addr := range msg.Envelope.To {
					if i > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(formatEmailAddress(addr))
				}
				sb.WriteString("\n")
			}
		case "date":
			if msg.Envelope != nil {
				sb.WriteString(fmt.Sprintf("Date: %s\n", msg.Envelope.Date.Format(time.RFC3339)))
			}
		case "flags":
			sb.WriteString(fmt.Sprintf("Flags: %v\n", msg.Flags))
		case "size":
			sb.WriteString(fmt.Sprintf("Size: %d bytes\n", msg.Size))
		case "mime_parts":
			if len(msg.MimeParts) > 0 {
				for _, part := range msg.MimeParts {
					if field.Content != nil && field.Content.ShowContent {
						content := part.Content
						if len(content) > field.Content.MaxLength && field.Content.MaxLength > 0 {
							content = content[:field.Content.MaxLength] + "..."
						}
						sb.WriteString(fmt.Sprintf("Content: %s\n", content))
					}
				}
			}
		}
	}

	return sb.String(), nil
}

// formatOutputTable formats message data as a table (simplified for now)
func formatOutputTable(msg *EmailMessage, config OutputConfig) (string, error) {
	// For simplicity, we'll just use the text format for now
	return formatOutputText(msg, config)
}

// formatEmailAddress formats an email address
func formatEmailAddress(addr EmailAddress) string {
	if addr.Name != "" {
		return fmt.Sprintf("%s <%s>", addr.Name, addr.Address)
	}
	return addr.Address
}
