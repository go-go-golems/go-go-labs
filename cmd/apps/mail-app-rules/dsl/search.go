package dsl

import (
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
)

// BuildSearchCriteria converts SearchConfig to imap.SearchCriteria
func BuildSearchCriteria(config SearchConfig) (*imap.SearchCriteria, error) {
	criteria := &imap.SearchCriteria{}

	// Process date criteria
	if config.Since != "" {
		since, err := parseDate(config.Since)
		if err != nil {
			return nil, fmt.Errorf("invalid 'since' date: %w", err)
		}
		criteria.Since = since
	}

	if config.Before != "" {
		before, err := parseDate(config.Before)
		if err != nil {
			return nil, fmt.Errorf("invalid 'before' date: %w", err)
		}
		criteria.Before = before
	}

	if config.On != "" {
		on, err := parseDate(config.On)
		if err != nil {
			return nil, fmt.Errorf("invalid 'on' date: %w", err)
		}

		// For "on" date, we need to set both since and before to cover the entire day
		// Since = start of the day, Before = start of the next day
		startOfDay := time.Date(on.Year(), on.Month(), on.Day(), 0, 0, 0, 0, on.Location())
		endOfDay := startOfDay.AddDate(0, 0, 1)

		criteria.Since = startOfDay
		criteria.Before = endOfDay
	}

	if config.WithinDays > 0 {
		// Calculate date from N days ago
		since := time.Now().AddDate(0, 0, -config.WithinDays)
		// Set to start of that day
		since = time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())
		criteria.Since = since
	}

	// Process header-based search criteria
	if config.From != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "From",
			Value: config.From,
		})
	}

	if config.To != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "To",
			Value: config.To,
		})
	}

	if config.Cc != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "Cc",
			Value: config.Cc,
		})
	}

	if config.Bcc != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "Bcc",
			Value: config.Bcc,
		})
	}

	if config.Subject != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "Subject",
			Value: config.Subject,
		})
	}

	if config.SubjectContains != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "Subject",
			Value: config.SubjectContains,
		})
	}

	if config.Header != nil && config.Header.Name != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   config.Header.Name,
			Value: config.Header.Value,
		})
	}

	// Process content-based search criteria
	if config.BodyContains != "" {
		criteria.Body = []string{config.BodyContains}
	}

	if config.Text != "" {
		criteria.Text = []string{config.Text}
	}

	// Process flag-based search criteria
	if config.Flags != nil {
		if len(config.Flags.Has) > 0 {
			for _, flag := range config.Flags.Has {
				// Convert flag name to IMAP format if needed
				imapFlag := convertToIMAPFlag(flag)
				criteria.Flag = append(criteria.Flag, imap.Flag(imapFlag))
			}
		}

		if len(config.Flags.NotHas) > 0 {
			for _, flag := range config.Flags.NotHas {
				// Convert flag name to IMAP format if needed
				imapFlag := convertToIMAPFlag(flag)
				criteria.NotFlag = append(criteria.NotFlag, imap.Flag(imapFlag))
			}
		}
	}

	// Process size-based search criteria
	if config.Size != nil {
		if config.Size.LargerThan != "" {
			size, err := parseSize(config.Size.LargerThan)
			if err != nil {
				return nil, fmt.Errorf("invalid 'larger_than' size: %w", err)
			}

			criteria.Larger = int64(size)
		}

		if config.Size.SmallerThan != "" {
			size, err := parseSize(config.Size.SmallerThan)
			if err != nil {
				return nil, fmt.Errorf("invalid 'smaller_than' size: %w", err)
			}

			criteria.Smaller = int64(size)
		}
	}

	return criteria, nil
}

// parseDate parses a date string in RFC3339 or ISO8601 format
func parseDate(dateStr string) (time.Time, error) {
	// Try RFC3339 format first
	t, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return t, nil
	}

	// Try ISO8601 date-only format
	t, err = time.Parse("2006-01-02", dateStr)
	if err == nil {
		return t, nil
	}

	// Try a few more common formats
	formats := []string{
		"2006/01/02",
		"01/02/2006",
		"02/01/2006",
		"Jan 2, 2006",
		"2 Jan 2006",
		time.RFC822,
		time.RFC1123,
	}

	for _, format := range formats {
		t, err = time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse date: %s", dateStr)
}

// convertToIMAPFlag converts a user-friendly flag name to IMAP format
func convertToIMAPFlag(flag string) string {
	// If it already starts with \ or $, return as is
	if strings.HasPrefix(flag, "\\") || strings.HasPrefix(flag, "$") {
		return flag
	}

	// Map of standard flag names to IMAP format
	standardFlags := map[string]string{
		"seen":      "\\Seen",
		"answered":  "\\Answered",
		"flagged":   "\\Flagged",
		"deleted":   "\\Deleted",
		"draft":     "\\Draft",
		"recent":    "\\Recent",
		"important": "$Important",
	}

	// Convert to lowercase for case-insensitive comparison
	flagLower := strings.ToLower(flag)

	// Check if it's a standard flag
	if imapFlag, ok := standardFlags[flagLower]; ok {
		return imapFlag
	}

	// Return as is for custom flags
	return flag
}
