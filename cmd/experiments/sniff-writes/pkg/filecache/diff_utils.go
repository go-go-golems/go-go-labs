package filecache

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// DiffLine represents a single line in a diff with its type
type DiffLine struct {
	Type    DiffLineType // Type of line (add, remove, context)
	Content string       // Content of the line
	OldLine int          // Line number in old file (0 if not applicable)
	NewLine int          // Line number in new file (0 if not applicable)
}

// DiffLineType represents the type of a diff line
type DiffLineType int

const (
	DiffLineContext  DiffLineType = iota // Unchanged line
	DiffLineAdd                          // Added line (+)
	DiffLineRemove                       // Removed line (-)
	DiffLineHeader                       // File header (---, +++)
	DiffLineLocation                     // Hunk header (@@)
	DiffLineElided                       // Elided content marker (...)
)

// ParseUnifiedDiff parses a unified diff into structured diff lines with line numbers
func ParseUnifiedDiff(diff string) []DiffLine {
	var lines []DiffLine
	scanner := bufio.NewScanner(strings.NewReader(diff))

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++"):
			lines = append(lines, DiffLine{
				Type:    DiffLineHeader,
				Content: line,
			})

		case strings.HasPrefix(line, "+"):
			// Parse line number from format "+123:content"
			oldLine, newLine, _ := parseLineWithNumber(line[1:], true)
			lines = append(lines, DiffLine{
				Type:    DiffLineAdd,
				Content: line, // Keep original line format with prefix
				OldLine: oldLine,
				NewLine: newLine,
			})
		case strings.HasPrefix(line, "-"):
			// Parse line number from format "-123:content"
			oldLine, newLine, _ := parseLineWithNumber(line[1:], false)
			lines = append(lines, DiffLine{
				Type:    DiffLineRemove,
				Content: line, // Keep original line format with prefix
				OldLine: oldLine,
				NewLine: newLine,
			})
		case strings.HasPrefix(line, " "):
			// Parse line number from format " 123:content"
			oldLine, newLine, _ := parseLineWithNumber(line[1:], true)
			lines = append(lines, DiffLine{
				Type:    DiffLineContext,
				Content: line, // Keep original line format with prefix
				OldLine: oldLine,
				NewLine: newLine,
			})
		default:
			lines = append(lines, DiffLine{
				Type:    DiffLineContext,
				Content: line,
			})
		}
	}

	return lines
}

// ElideUnifiedDiff takes a unified diff and limits context lines around changes
func ElideUnifiedDiff(diff string, contextLines int) string {
	if contextLines < 0 {
		return diff
	}

	lines := ParseUnifiedDiff(diff)
	if len(lines) == 0 {
		return diff
	}

	// Find header lines first
	var headerLines []DiffLine
	var contentLines []DiffLine

	for _, line := range lines {
		if line.Type == DiffLineHeader {
			headerLines = append(headerLines, line)
		} else {
			contentLines = append(contentLines, line)
		}
	}

	if len(contentLines) == 0 {
		return diff
	}

	// Find positions of changed lines
	var changePositions []int
	for i, line := range contentLines {
		if line.Type == DiffLineAdd || line.Type == DiffLineRemove {
			changePositions = append(changePositions, i)
		}
	}

	if len(changePositions) == 0 {
		// No changes, return original
		return diff
	}

	// Calculate which lines to keep
	linesToKeep := make(map[int]bool)

	for _, pos := range changePositions {
		// Keep the change line itself
		linesToKeep[pos] = true

		// Keep context lines before
		for i := pos - contextLines; i < pos; i++ {
			if i >= 0 {
				linesToKeep[i] = true
			}
		}

		// Keep context lines after
		for i := pos + 1; i <= pos+contextLines; i++ {
			if i < len(contentLines) {
				linesToKeep[i] = true
			}
		}
	}

	// Build the elided diff
	var result strings.Builder

	// Add headers
	for _, header := range headerLines {
		result.WriteString(header.Content + "\n")
	}

	// Add content with elision
	lastKeptLine := -1 // Start before any valid index

	for i, line := range contentLines {
		if linesToKeep[i] {
			// Check if we need to add an elision marker
			// Only add "..." if there's a gap between kept lines
			if lastKeptLine >= 0 && i > lastKeptLine+1 {
				result.WriteString("...\n")
			}
			result.WriteString(line.Content + "\n")
			lastKeptLine = i
		}
	}

	return result.String()
}

// GenerateElidedUnifiedDiff creates a unified diff with context limiting
func GenerateElidedUnifiedDiff(oldContent, newContent []byte, filename string, contextLines int) string {
	oldLines := strings.Split(string(oldContent), "\n")
	newLines := strings.Split(string(newContent), "\n")

	// Generate a basic unified diff first
	diff := generateBasicUnifiedDiff(oldLines, newLines, filename)

	// Apply elision if needed
	if contextLines > 0 {
		return ElideUnifiedDiff(diff, contextLines)
	}

	return diff
}

// generateBasicUnifiedDiff creates a simple unified diff
func generateBasicUnifiedDiff(oldLines, newLines []string, filename string) string {
	var result strings.Builder

	// Add headers
	result.WriteString(fmt.Sprintf("--- %s (cached)\n", filename))
	result.WriteString(fmt.Sprintf("+++ %s (new write)\n", filename))

	// Simple diff algorithm - for real implementation you'd want Myers or similar
	maxLines := len(oldLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}

	for i := 0; i < maxLines; i++ {
		oldLine := ""
		newLine := ""

		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine == newLine {
			// Context line
			if i < len(oldLines) {
				result.WriteString(" " + oldLine + "\n")
			}
		} else {
			// Lines differ
			if i < len(oldLines) && oldLine != "" {
				result.WriteString("-" + oldLine + "\n")
			}
			if i < len(newLines) && newLine != "" {
				result.WriteString("+" + newLine + "\n")
			}
		}
	}

	return result.String()
}

// parseLineWithNumber parses a line in format "123:content" and extracts line number and content
func parseLineWithNumber(line string, isAddOrContext bool) (oldLine int, newLine int, content string) {
	// Find the first colon
	colonIndex := strings.Index(line, ":")
	if colonIndex == -1 {
		// No line number format, return original line
		return 0, 0, line
	}

	// Extract line number
	lineNumStr := line[:colonIndex]
	lineNum, err := strconv.Atoi(lineNumStr)
	if err != nil {
		// Invalid line number, return original line
		return 0, 0, line
	}

	// Extract content (everything after the colon)
	content = line[colonIndex+1:]

	// Set appropriate line numbers based on line type
	if isAddOrContext {
		// For add and context lines, this is the new file line number
		newLine = lineNum
	} else {
		// For remove lines, this is the old file line number
		oldLine = lineNum
	}

	return oldLine, newLine, content
}
