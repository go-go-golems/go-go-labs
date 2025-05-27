package formatter

import (
	"bufio"
	"strings"

	"github.com/fatih/color"
)

// ColoredDiffFormatter provides colored diff output for terminals
type ColoredDiffFormatter struct {
	// Color functions for different parts of the diff
	headerColor   *color.Color
	addedColor    *color.Color
	removedColor  *color.Color
	contextColor  *color.Color
	locationColor *color.Color
}

// NewColoredDiffFormatter creates a new colored diff formatter
func NewColoredDiffFormatter() *ColoredDiffFormatter {
	return &ColoredDiffFormatter{
		headerColor:   color.New(color.FgCyan, color.Bold),
		addedColor:    color.New(color.FgGreen),
		removedColor:  color.New(color.FgRed),
		contextColor:  color.New(color.FgWhite),
		locationColor: color.New(color.FgMagenta),
	}
}

// FormatDiff takes a unified diff string and returns a colored version
func (f *ColoredDiffFormatter) FormatDiff(diff string) string {
	if diff == "" {
		return diff
	}

	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(diff))

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++"):
			// File headers
			result.WriteString(f.headerColor.Sprint(line))
		case strings.HasPrefix(line, "@@"):
			// Location headers (hunk headers)
			result.WriteString(f.locationColor.Sprint(line))
		case strings.HasPrefix(line, "+"):
			// Added lines
			result.WriteString(f.addedColor.Sprint(line))
		case strings.HasPrefix(line, "-"):
			// Removed lines
			result.WriteString(f.removedColor.Sprint(line))
		case line == "...":
			// Elided content marker
			result.WriteString(f.locationColor.Sprint("..."))
		default:
			// Context lines (unchanged)
			result.WriteString(f.contextColor.Sprint(line))
		}
		result.WriteString("\n")
	}

	return result.String()
}

// FormatDiffForWeb takes a unified diff string and returns HTML with CSS classes
func FormatDiffForWeb(diff string) string {
	if diff == "" {
		return diff
	}

	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(diff))

	for scanner.Scan() {
		line := scanner.Text()
		escapedLine := escapeHTMLString(line)

		switch {
		case strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++"):
			// File headers
			result.WriteString(`<div class="diff-header">` + escapedLine + `</div>`)
		case strings.HasPrefix(line, "@@"):
			// Location headers (hunk headers)
			result.WriteString(`<div class="diff-location">` + escapedLine + `</div>`)
		case strings.HasPrefix(line, "+"):
			// Added lines
			result.WriteString(`<div class="diff-added">` + escapedLine + `</div>`)
		case strings.HasPrefix(line, "-"):
			// Removed lines
			result.WriteString(`<div class="diff-removed">` + escapedLine + `</div>`)
		case line == "...":
			// Elided content marker
			result.WriteString(`<div class="diff-elided">` + escapedLine + `</div>`)
		default:
			// Context lines (unchanged)
			result.WriteString(`<div class="diff-context">` + escapedLine + `</div>`)
		}
	}

	return result.String()
}

// escapeHTMLString escapes HTML special characters
func escapeHTMLString(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#x27;")
	return s
}

// IsTerminalSupported checks if the current terminal supports colors
func IsTerminalSupported() bool {
	return color.NoColor == false
}
