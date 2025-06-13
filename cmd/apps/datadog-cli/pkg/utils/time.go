package utils

import (
	"time"

	"github.com/pkg/errors"
)

// ParseTimeParameter parses various time formats including relative times
func ParseTimeParameter(timeStr string) (time.Time, error) {
	// Handle "now"
	if timeStr == "now" {
		return time.Now(), nil
	}

	// Handle relative times like "-1h", "-30m", "-1d"
	if len(timeStr) > 0 && (timeStr[0] == '-' || timeStr[0] == '+') {
		duration, err := time.ParseDuration(timeStr)
		if err != nil {
			// Try parsing as days (not supported by time.ParseDuration)
			if len(timeStr) > 1 && timeStr[len(timeStr)-1] == 'd' {
				days := timeStr[:len(timeStr)-1]
				if d, err := time.ParseDuration(days + "h"); err == nil {
					duration = d * 24
				} else {
					return time.Time{}, err
				}
			} else {
				return time.Time{}, err
			}
		}
		return time.Now().Add(duration), nil
	}

	// Try various time formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.Errorf("unable to parse time: %s", timeStr)
}
