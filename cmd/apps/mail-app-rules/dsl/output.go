package dsl

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

// BuildFetchOptions converts OutputConfig to imap.FetchOptions
func BuildFetchOptions(config OutputConfig) (*imap.FetchOptions, error) {
	options := &imap.FetchOptions{}

	// Process fields
	for _, fieldInterface := range config.Fields {
		field, ok := fieldInterface.(Field)
		if !ok {
			// Skip fields that couldn't be properly parsed
			continue
		}

		switch field.Name {
		case "uid":
			options.UID = true
		case "envelope", "subject", "from", "to", "date":
			// All these fields require the envelope
			options.Envelope = true
		case "flags":
			options.Flags = true
		case "size":
			options.RFC822Size = true
		case "body":
			if field.Body != nil {
				// Configure body section fetch
				bodySection := &imap.FetchItemBodySection{
					Specifier: "TEXT",
					Peek:      true,
				}
				options.BodySection = append(options.BodySection, bodySection)
			}
		}
	}

	return options, nil
}

// FormatOutput formats message data according to OutputConfig
func FormatOutput(msgData *imapclient.FetchMessageBuffer, config OutputConfig) (string, error) {
	switch config.Format {
	case "json":
		return formatOutputJSON(msgData, config)
	case "table":
		return formatOutputTable(msgData, config)
	default:
		// Default to text format
		return formatOutputText(msgData, config)
	}
}

// formatOutputJSON formats message data as JSON
func formatOutputJSON(msgData *imapclient.FetchMessageBuffer, config OutputConfig) (string, error) {
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
			output["uid"] = msgData.UID
		case "subject":
			if msgData.Envelope != nil {
				output["subject"] = msgData.Envelope.Subject
			}
		case "from":
			if msgData.Envelope != nil && len(msgData.Envelope.From) > 0 {
				output["from"] = formatAddress(msgData.Envelope.From[0])
			}
		case "to":
			if msgData.Envelope != nil && len(msgData.Envelope.To) > 0 {
				addresses := make([]string, 0, len(msgData.Envelope.To))
				for _, addr := range msgData.Envelope.To {
					addresses = append(addresses, formatAddress(addr))
				}
				output["to"] = addresses
			}
		case "date":
			if msgData.Envelope != nil {
				output["date"] = msgData.Envelope.Date.Format(time.RFC3339)
			}
		case "flags":
			if msgData.Flags != nil {
				output["flags"] = msgData.Flags
			}
		case "size":
			output["size"] = msgData.RFC822Size
		case "body":
			// For simplicity, we'll just get the first body section
			if field.Body != nil && len(msgData.BodySection) > 0 {
				bodyText := string(msgData.BodySection[0].Bytes)
				if field.Body.MaxLength > 0 && len(bodyText) > field.Body.MaxLength {
					bodyText = bodyText[:field.Body.MaxLength] + "..."
				}
				output["body"] = bodyText
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
func formatOutputText(msgData *imapclient.FetchMessageBuffer, config OutputConfig) (string, error) {
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
			sb.WriteString(fmt.Sprintf("UID: %d\n", msgData.UID))
		case "subject":
			if msgData.Envelope != nil {
				sb.WriteString(fmt.Sprintf("Subject: %s\n", msgData.Envelope.Subject))
			}
		case "from":
			if msgData.Envelope != nil && len(msgData.Envelope.From) > 0 {
				sb.WriteString(fmt.Sprintf("From: %s\n", formatAddress(msgData.Envelope.From[0])))
			}
		case "to":
			if msgData.Envelope != nil && len(msgData.Envelope.To) > 0 {
				addresses := make([]string, 0, len(msgData.Envelope.To))
				for _, addr := range msgData.Envelope.To {
					addresses = append(addresses, formatAddress(addr))
				}
				sb.WriteString(fmt.Sprintf("To: %s\n", strings.Join(addresses, ", ")))
			}
		case "date":
			if msgData.Envelope != nil {
				sb.WriteString(fmt.Sprintf("Date: %s\n", msgData.Envelope.Date.Format(time.RFC3339)))
			}
		case "flags":
			if msgData.Flags != nil {
				sb.WriteString(fmt.Sprintf("Flags: %v\n", msgData.Flags))
			}
		case "size":
			sb.WriteString(fmt.Sprintf("Size: %d bytes\n", msgData.RFC822Size))
		case "body":
			// For simplicity, we'll just get the first body section
			if field.Body != nil && len(msgData.BodySection) > 0 {
				bodyText := string(msgData.BodySection[0].Bytes)
				if field.Body.MaxLength > 0 && len(bodyText) > field.Body.MaxLength {
					bodyText = bodyText[:field.Body.MaxLength] + "..."
				}
				sb.WriteString(fmt.Sprintf("\nBody:\n%s\n", bodyText))
			}
		}
	}

	return sb.String(), nil
}

// formatOutputTable formats message data as a table (simplified for now)
func formatOutputTable(msgData *imapclient.FetchMessageBuffer, config OutputConfig) (string, error) {
	// For simplicity, we'll just use the text format for now
	// In a real implementation, this would format the data as a table
	return formatOutputText(msgData, config)
}

// formatAddress formats an email address
func formatAddress(addr imap.Address) string {
	if addr.Name != "" {
		return fmt.Sprintf("%s <%s@%s>", addr.Name, addr.Mailbox, addr.Host)
	}
	return fmt.Sprintf("%s@%s", addr.Mailbox, addr.Host)
}
