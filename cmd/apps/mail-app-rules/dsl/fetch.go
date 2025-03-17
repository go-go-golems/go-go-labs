package dsl

import (
	"fmt"
	"strings"

	"github.com/emersion/go-imap/v2"
)

// MimePartMetadata contains information about a MIME part to be fetched
type MimePartMetadata struct {
	FetchSection *imap.FetchItemBodySection
	Type         string
	Subtype      string
	Params       map[string]string
	IsAttachment bool
	Filename     string
	Path         []int
}

// determineRequiredBodySections analyzes the output config and body structure to determine which parts need to be fetched
func determineRequiredBodySections(
	bodyStructure imap.BodyStructure,
	config OutputConfig,
) ([]MimePartMetadata, error) {
	if bodyStructure == nil {
		return nil, fmt.Errorf("no body structure provided")
	}

	var parts []MimePartMetadata

	// Check if we need MIME parts
	var contentField *ContentField
	needsMimeParts := false

	for _, fieldInterface := range config.Fields {
		field, ok := fieldInterface.(Field)
		if !ok {
			continue
		}

		if field.Name == "mime_parts" {
			needsMimeParts = true
			contentField = field.Content
			break
		}
	}

	// If we don't need MIME parts, return empty slice
	if !needsMimeParts {
		return parts, nil
	}

	// Helper function to determine if we should include a part based on content field settings
	shouldIncludePart := func(mediaType string) bool {
		if contentField == nil {
			return true
		}

		switch contentField.Mode {
		case "text_only":
			return strings.HasPrefix(mediaType, "text/plain")
		case "filter":
			if len(contentField.Types) == 0 {
				return true
			}
			for _, allowedType := range contentField.Types {
				if strings.HasSuffix(allowedType, "/*") {
					prefix := strings.TrimSuffix(allowedType, "/*")
					if strings.HasPrefix(mediaType, prefix+"/") {
						return true
					}
				} else if mediaType == allowedType {
					return true
				}
			}
			return false
		default: // "full" or empty
			return true
		}
	}

	// Walk through the structure and collect required sections
	bodyStructure.Walk(func(path []int, part imap.BodyStructure) bool {
		mediaType := part.MediaType()

		// For multipart containers, we don't fetch the part itself
		if strings.HasPrefix(mediaType, "multipart/") {
			return true
		}

		if shouldIncludePart(mediaType) {
			// Create section for this part
			section := &imap.FetchItemBodySection{
				Peek: true, // Don't mark as read
				Part: path,
			}

			if contentField.MaxLength > 0 {
				section.Partial = &imap.SectionPartial{
					Offset: 0,
					// fetch 1 more to be able to elide ... later on
					Size: int64(contentField.MaxLength) + 1,
				}
			}

			// Extract MIME information
			mimeType := part.MediaType()

			// Determine if it's an attachment and get filename
			isAttachment := false
			filename := ""
			if disp := part.Disposition(); disp != nil {
				isAttachment = disp.Value == "attachment"
				if len(disp.Params) > 0 {
					filename = disp.Params["filename"]
				}
			}

			metadata := MimePartMetadata{
				FetchSection: section,
				Type:         mimeType,
				Params:       map[string]string{}, // Initialize empty map since we can't access params directly
				IsAttachment: isAttachment,
				Filename:     filename,
				Path:         path,
			}

			parts = append(parts, metadata)
		}

		return true
	})

	return parts, nil
}

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
