package dsl

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
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
		case "mime_parts":
			// We need the body structure for MIME parts
			options.BodyStructure = &imap.FetchItemBodyStructure{
				Extended: true,
			}
		}
	}

	return options, nil
}

// FormatOutput formats message data according to OutputConfig
func FormatOutput(msgData *imapclient.FetchMessageBuffer, config OutputConfig, bodySectionData imapclient.FetchItemDataBodySection) (string, error) {
	switch config.Format {
	case "json":
		return formatOutputJSON(msgData, config, bodySectionData)
	case "table":
		return formatOutputTable(msgData, config, bodySectionData)
	default:
		// Default to text format
		return formatOutputText(msgData, config, bodySectionData)
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

// fetchMimeParts extracts MIME part content types and content from the message
func fetchMimeParts(
	bodySectionData imapclient.FetchItemDataBodySection,
) ([]MimePart, error) {
	result := []MimePart{}

	mr, err := mail.CreateReader(bodySectionData.Literal)
	if err != nil {
		return nil, err
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch header := part.Header.(type) {
		case *mail.InlineHeader:
			contentType, params, _ := header.ContentType()
			content, err := io.ReadAll(part.Body)
			if err != nil {
				return nil, err
			}
			result = append(result, MimePart{
				Type:    contentType,
				Subtype: params["subtype"],
				Content: string(content),
				Size:    uint32(len(content)),
				Charset: params["charset"],
			})
		case *mail.AttachmentHeader:
			filename, _ := header.Filename()
			contentType, _, _ := header.ContentType()
			content, err := io.ReadAll(part.Body)
			if err != nil {
				return nil, err
			}
			result = append(result, MimePart{
				Filename: filename,
				Type:     contentType,
				Size:     uint32(len(content)),
			})
		}
	}

	return result, nil
}

// formatOutputJSON formats message data as JSON
func formatOutputJSON(msgData *imapclient.FetchMessageBuffer, config OutputConfig, bodySectionData imapclient.FetchItemDataBodySection) (string, error) {
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
		case "mime_parts":
			if msgData.BodyStructure != nil {
				mimeParts, err := fetchMimeParts(bodySectionData)
				if err != nil {
					return "", err
				}
				output["mime_parts"] = mimeParts
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
func formatOutputText(msgData *imapclient.FetchMessageBuffer, config OutputConfig, bodySectionData imapclient.FetchItemDataBodySection) (string, error) {
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
				sb.WriteString("To: ")
				for i, addr := range msgData.Envelope.To {
					if i > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(formatAddress(addr))
				}
				sb.WriteString("\n")
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
		case "mime_parts":
			content, err := io.ReadAll(bodySectionData.Literal)
			_ = content
			_ = err
			if msgData.BodyStructure != nil {
				for _, section := range msgData.BodySection {

					/// XXX properly print out the collected sections
					if len(section.Bytes) > 0 && field.Content != nil && field.Content.ShowContent {
						var data []byte
						if len(section.Bytes) > field.Content.MaxLength && field.Content.MaxLength > 0 {
							data = section.Bytes[:field.Content.MaxLength]
							data = append(data, []byte("...")...)
						} else {
							data = section.Bytes
						}
						sb.WriteString(fmt.Sprintf("Content: %s\n", string(data)))
					}
				}
			}
		}
	}

	return sb.String(), nil
}

// formatOutputTable formats message data as a table (simplified for now)
func formatOutputTable(msgData *imapclient.FetchMessageBuffer, config OutputConfig, bodySectionData imapclient.FetchItemDataBodySection) (string, error) {
	// For simplicity, we'll just use the text format for now
	// In a real implementation, this would format the data as a table with proper column alignment
	return formatOutputText(msgData, config, bodySectionData)
}

// formatAddress formats an email address
func formatAddress(addr imap.Address) string {
	if addr.Name != "" {
		return fmt.Sprintf("%s <%s@%s>", addr.Name, addr.Mailbox, addr.Host)
	}
	return fmt.Sprintf("%s@%s", addr.Mailbox, addr.Host)
}
