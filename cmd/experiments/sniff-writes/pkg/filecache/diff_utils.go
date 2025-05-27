package filecache

import (
	"bufio"
	"fmt"
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
	DiffLineContext DiffLineType = iota // Unchanged line
	DiffLineAdd                         // Added line (+)
	DiffLineRemove                      // Removed line (-)
	DiffLineHeader                      // File header (---, +++)
	DiffLineLocation                    // Hunk header (@@)
	DiffLineElided                      // Elided content marker (...)
)

// ParseUnifiedDiff parses a unified diff into structured diff lines
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
		case strings.HasPrefix(line, "@@"):
			lines = append(lines, DiffLine{
				Type:    DiffLineLocation,
				Content: line,
			})
		case strings.HasPrefix(line, "+"):
			lines = append(lines, DiffLine{
				Type:    DiffLineAdd,
				Content: line,
			})
		case strings.HasPrefix(line, "-"):
			lines = append(lines, DiffLine{
				Type:    DiffLineRemove,
				Content: line,
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
	var locationLine DiffLine
	
	for _, line := range lines {
		if line.Type == DiffLineHeader {
			headerLines = append(headerLines, line)
		} else if line.Type == DiffLineLocation {
			locationLine = line
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
	
	// Add location header (update it if necessary)
	if locationLine.Content != "" {
		result.WriteString(locationLine.Content + "\n")
	}
	
	// Add content with elision
	lastKeptLine := -2 // Start before any valid index
	
	for i, line := range contentLines {
		if linesToKeep[i] {
			// Check if we need to add an elision marker
			if i > lastKeptLine+1 {
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
	result.WriteString(fmt.Sprintf("@@ -1,%d +1,%d @@\n", len(oldLines), len(newLines)))
	
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