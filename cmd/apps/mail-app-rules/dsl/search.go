package dsl

import (
	"fmt"
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

	// Process From criteria
	if config.From != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "From",
			Value: config.From,
		})
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
		"Jan 2, 2006",
		"2 Jan 2006",
	}

	for _, format := range formats {
		t, err = time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse date: %s", dateStr)
}
