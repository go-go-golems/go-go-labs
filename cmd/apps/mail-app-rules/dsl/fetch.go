package dsl

import (
	"fmt"
	"strings"

	"github.com/emersion/go-imap/v2"
)

// determineRequiredBodySections analyzes the output config and body structure to determine which parts need to be fetched
func determineRequiredBodySections(
	bodyStructure imap.BodyStructure,
	config OutputConfig,
) ([]*imap.FetchItemBodySection, error) {
	if bodyStructure == nil {
		return nil, fmt.Errorf("no body structure provided")
	}

	var sections []*imap.FetchItemBodySection

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
		return sections, nil
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
				// Part: path,
				Peek: true, // Don't mark as read
				Part: path,
				// Specifier: "TEXT", // Get the actual content
			}
			sections = append(sections, section)

		}

		return true
	})

	return sections, nil
}
