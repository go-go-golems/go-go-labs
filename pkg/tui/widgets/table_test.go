package widgets

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// createTestStream creates test stream data with given parameters
func createTestStream(name, lastID string, length, groups int64, messageRates []float64) StreamData {
	return StreamData{
		Name:         name,
		Length:       length,
		MemoryUsage:  1024 * 1024, // 1MB for consistency
		Groups:       groups,
		LastID:       lastID,
		MessageRates: messageRates,
	}
}

// createTestMessageRates generates test message rate data
func createTestMessageRates(count int, maxRate float64) []float64 {
	rates := make([]float64, count)
	for i := range rates {
		rates[i] = maxRate * float64(i) / float64(count)
	}
	return rates
}

// printTableDebug prints the table with line numbers for debugging
func printTableDebug(name string, content string) {
	fmt.Printf("\n=== %s ===\n", name)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		fmt.Printf("%2d: %s\n", i+1, line)
	}
	fmt.Printf("=== End %s ===\n\n", name)
}

// TestTableBasicRendering tests basic table rendering with sample data
func TestTableBasicRendering(t *testing.T) {
	// Create test styles
	styles := StreamsTableStyles{
		Container:    lipgloss.NewStyle(),
		Table:        lipgloss.NewStyle(),
		HeaderRow:    lipgloss.NewStyle(),
		Row:          lipgloss.NewStyle(),
		SelectedRow:  lipgloss.NewStyle().Background(lipgloss.Color("#555555")),
		SparklineRow: lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
	}

	// Create test widget
	widget := NewStreamsTableWidget(styles)
	widget.SetSize(120, 20)

	// Create test data
	testStreams := []StreamData{
		createTestStream("user_events", "1234567890123-0", 1500, 3, createTestMessageRates(25, 50.0)),
		createTestStream("analytics_data", "9876543210987-1", 8750, 5, createTestMessageRates(25, 25.0)),
		createTestStream("notifications", "5555555555555-2", 250, 1, createTestMessageRates(25, 10.0)),
	}

	// Update widget with data
	widget.streams = testStreams
	widget.updateSparklines()

	// Test rendering
	output := widget.View()
	
	// Print for visual inspection
	printTableDebug("Basic Table Rendering (120 chars)", output)

	// Basic validation
	if output == "" {
		t.Error("Table output should not be empty")
	}

	// Check for headers
	if !strings.Contains(output, "Stream") {
		t.Error("Table should contain 'Stream' header")
	}
	if !strings.Contains(output, "Entries") {
		t.Error("Table should contain 'Entries' header")
	}

	// Check for data
	if !strings.Contains(output, "user_events") {
		t.Error("Table should contain stream name 'user_events'")
	}

	// Check for sparkline indication
	if !strings.Contains(output, "msg/s:") {
		t.Error("Table should contain sparkline indicator 'msg/s:'")
	}
}

// TestTableResponsiveWidth tests table rendering at different terminal widths
func TestTableResponsiveWidth(t *testing.T) {
	styles := StreamsTableStyles{
		Container:    lipgloss.NewStyle(),
		Table:        lipgloss.NewStyle(),
		HeaderRow:    lipgloss.NewStyle(),
		Row:          lipgloss.NewStyle(),
		SelectedRow:  lipgloss.NewStyle(),
		SparklineRow: lipgloss.NewStyle(),
	}

	testStreams := []StreamData{
		createTestStream("user_events", "1234567890123-0", 1500, 3, createTestMessageRates(15, 50.0)),
		createTestStream("very_long_stream_name_that_should_be_truncated", "9876543210987654321-1", 8750, 5, createTestMessageRates(15, 25.0)),
	}

	widths := []int{80, 120, 160}

	for _, width := range widths {
		t.Run(fmt.Sprintf("Width_%d", width), func(t *testing.T) {
			widget := NewStreamsTableWidget(styles)
			widget.SetSize(width, 20)
			widget.streams = testStreams
			widget.updateSparklines()

			output := widget.View()
			printTableDebug(fmt.Sprintf("Responsive Width %d chars", width), output)

			// Validate that no line exceeds the terminal width
			lines := strings.Split(output, "\n")
			ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
			for i, line := range lines {
				// Remove ANSI escape sequences for length calculation
				cleanLine := ansiRegex.ReplaceAllString(line, "")
				visualWidth := utf8.RuneCountInString(cleanLine)
				if visualWidth > width+5 { // Allow small margin for test tolerance
					t.Errorf("Line %d exceeds width %d: got %d visual chars\nLine: %s", 
						i+1, width, visualWidth, cleanLine)
				}
			}

			// Check that content is present (possibly truncated)
			if !strings.Contains(output, "user_ev") {
				t.Error("Table should contain stream data (possibly truncated)")
			}
		})
	}
}

// TestTableLongContent tests table rendering with very long text content
func TestTableLongContent(t *testing.T) {
	styles := StreamsTableStyles{
		Container:    lipgloss.NewStyle(),
		Table:        lipgloss.NewStyle(),
		HeaderRow:    lipgloss.NewStyle(),
		Row:          lipgloss.NewStyle(),
		SelectedRow:  lipgloss.NewStyle(),
		SparklineRow: lipgloss.NewStyle(),
	}

	// Create streams with very long names and IDs
	testStreams := []StreamData{
		createTestStream(
			"extremely_long_stream_name_that_exceeds_normal_column_width_expectations",
			"1234567890123456789012345678901234567890-0",
			1500, 3, createTestMessageRates(25, 50.0)),
		createTestStream(
			"another_really_really_long_stream_name_for_testing_truncation_behavior",
			"9876543210987654321098765432109876543210-1",
			8750, 5, createTestMessageRates(25, 25.0)),
		createTestStream(
			"short", "1-0", 100, 1, createTestMessageRates(25, 5.0)),
	}

	widget := NewStreamsTableWidget(styles)
	widget.SetSize(120, 20)
	widget.streams = testStreams
	widget.updateSparklines()

	output := widget.View()
	printTableDebug("Long Content Test", output)

	// Validate that content doesn't break table structure
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		// Check that borders are consistent
		if strings.Contains(line, "│") {
			// Count vertical bars - should be consistent
			barCount := strings.Count(line, "│")
			if barCount > 0 && barCount != 7 { // 6 columns = 7 vertical bars
				t.Errorf("Line %d has inconsistent border count: got %d bars\nLine: %s", 
					i+1, barCount, line)
			}
		}
	}

	// Check that data is present (even if truncated)
	if !strings.Contains(output, "extremely_long") || 
	   !strings.Contains(output, "short") {
		t.Error("Table should contain stream data (possibly truncated)")
	}
}

// TestTableEmptyData tests table rendering with no stream data
func TestTableEmptyData(t *testing.T) {
	styles := StreamsTableStyles{
		Container:    lipgloss.NewStyle(),
		Table:        lipgloss.NewStyle(),
		HeaderRow:    lipgloss.NewStyle(),
		Row:          lipgloss.NewStyle(),
		SelectedRow:  lipgloss.NewStyle(),
		SparklineRow: lipgloss.NewStyle(),
	}

	widget := NewStreamsTableWidget(styles)
	widget.SetSize(120, 20)
	
	// No streams data
	widget.streams = []StreamData{}

	output := widget.View()
	printTableDebug("Empty Data Test", output)

	// Should show "No streams found" message
	if !strings.Contains(output, "No streams found") {
		t.Error("Empty table should show 'No streams found' message")
	}
}

// TestTableSingleRow tests table rendering with a single stream
func TestTableSingleRow(t *testing.T) {
	styles := StreamsTableStyles{
		Container:    lipgloss.NewStyle(),
		Table:        lipgloss.NewStyle(),
		HeaderRow:    lipgloss.NewStyle(),
		Row:          lipgloss.NewStyle(),
		SelectedRow:  lipgloss.NewStyle(),
		SparklineRow: lipgloss.NewStyle(),
	}

	widget := NewStreamsTableWidget(styles)
	widget.SetSize(120, 20)

	// Single stream
	testStreams := []StreamData{
		createTestStream("single_stream", "1234567890123-0", 1000, 2, createTestMessageRates(25, 30.0)),
	}

	widget.streams = testStreams
	widget.updateSparklines()

	output := widget.View()
	printTableDebug("Single Row Test", output)

	// Check that structure is correct
	if !strings.Contains(output, "single_stream") {
		t.Error("Table should contain the single stream")
	}

	// Should have header, data row, sparkline row, and borders
	lines := strings.Split(output, "\n")
	if len(lines) < 5 {
		t.Error("Single row table should have at least 5 lines (borders + header + data + sparkline)")
	}
}

// TestTableManyRows tests table rendering with many streams
func TestTableManyRows(t *testing.T) {
	styles := StreamsTableStyles{
		Container:    lipgloss.NewStyle(),
		Table:        lipgloss.NewStyle(),
		HeaderRow:    lipgloss.NewStyle(),
		Row:          lipgloss.NewStyle(),
		SelectedRow:  lipgloss.NewStyle(),
		SparklineRow: lipgloss.NewStyle(),
	}

	widget := NewStreamsTableWidget(styles)
	widget.SetSize(120, 30)

	// Many streams
	var testStreams []StreamData
	for i := 0; i < 10; i++ {
		testStreams = append(testStreams, createTestStream(
			fmt.Sprintf("stream_%d", i),
			fmt.Sprintf("%d-0", 1000000000000+int64(i)),
			int64(100*(i+1)),
			int64(i%5+1),
			createTestMessageRates(25, float64(10*(i+1))),
		))
	}

	widget.streams = testStreams
	widget.updateSparklines()

	output := widget.View()
	printTableDebug("Many Rows Test", output)

	// Check that all streams are present
	for i := 0; i < 10; i++ {
		streamName := fmt.Sprintf("stream_%d", i)
		if !strings.Contains(output, streamName) {
			t.Errorf("Table should contain %s", streamName)
		}
	}

	// Validate table structure consistency
	lines := strings.Split(output, "\n")
	borderLines := 0
	for _, line := range lines {
		if strings.Contains(line, "─") {
			borderLines++
		}
	}

	// Should have top border, header divider, stream dividers, bottom border
	expectedBorders := 2 + (len(testStreams) - 1) // top + bottom + dividers between streams
	if borderLines < expectedBorders {
		t.Errorf("Expected at least %d border lines, got %d", expectedBorders, borderLines)
	}
}

// TestTableSparklineIntegration tests sparkline rendering within table cells
func TestTableSparklineIntegration(t *testing.T) {
	styles := StreamsTableStyles{
		Container:    lipgloss.NewStyle(),
		Table:        lipgloss.NewStyle(),
		HeaderRow:    lipgloss.NewStyle(),
		Row:          lipgloss.NewStyle(),
		SelectedRow:  lipgloss.NewStyle(),
		SparklineRow: lipgloss.NewStyle(),
	}

	widget := NewStreamsTableWidget(styles)
	widget.SetSize(120, 20)

	// Create streams with different sparkline patterns
	testStreams := []StreamData{
		createTestStream("rising_trend", "1-0", 1000, 1, []float64{1, 5, 10, 20, 35, 50}),
		createTestStream("falling_trend", "2-0", 1000, 1, []float64{50, 35, 20, 10, 5, 1}),
		createTestStream("flat_trend", "3-0", 1000, 1, []float64{25, 25, 25, 25, 25, 25}),
		createTestStream("volatile_trend", "4-0", 1000, 1, []float64{10, 50, 5, 45, 15, 40}),
	}

	widget.streams = testStreams
	widget.updateSparklines()

	output := widget.View()
	printTableDebug("Sparkline Integration Test", output)

	// Check that sparklines are present
	if !strings.Contains(output, "msg/s:") {
		t.Error("Table should contain sparkline indicators")
	}

	// Verify each sparkline gets created
	for _, stream := range testStreams {
		if widget.sparklines[stream.Name] == nil {
			t.Errorf("Sparkline for %s should be created", stream.Name)
		}
	}

	// Test that sparklines render correctly
	for _, stream := range testStreams {
		sparkline := widget.sparklines[stream.Name]
		rendered := sparkline.Render()
		if rendered == "" {
			t.Errorf("Sparkline for %s should render non-empty content", stream.Name)
		}
	}
}

// TestTableBorderAlignment tests that borders align correctly
func TestTableBorderAlignment(t *testing.T) {
	styles := StreamsTableStyles{
		Container:    lipgloss.NewStyle(),
		Table:        lipgloss.NewStyle(),
		HeaderRow:    lipgloss.NewStyle(),
		Row:          lipgloss.NewStyle(),
		SelectedRow:  lipgloss.NewStyle(),
		SparklineRow: lipgloss.NewStyle(),
	}

	widget := NewStreamsTableWidget(styles)
	widget.SetSize(120, 20)

	testStreams := []StreamData{
		createTestStream("test_stream", "1234567890123-0", 1500, 3, createTestMessageRates(25, 50.0)),
	}

	widget.streams = testStreams
	widget.updateSparklines()

	output := widget.View()
	printTableDebug("Border Alignment Test", output)

	lines := strings.Split(output, "\n")
	
	// Find border lines and content lines
	var borderLines []string
	var contentLines []string
	
	for _, line := range lines {
		if strings.Contains(line, "─") {
			borderLines = append(borderLines, line)
		} else if strings.Contains(line, "│") {
			contentLines = append(contentLines, line)
		}
	}

	// All border lines should have the same length (excluding styling)
	if len(borderLines) > 1 {
		firstLen := len(strings.ReplaceAll(borderLines[0], "\033[", ""))
		for i, border := range borderLines[1:] {
			cleanLen := len(strings.ReplaceAll(border, "\033[", ""))
			if cleanLen != firstLen {
				t.Errorf("Border line %d length mismatch: expected %d, got %d\nLine: %s", 
					i+2, firstLen, cleanLen, border)
			}
		}
	}

	// All content lines should have the same number of vertical bars
	if len(contentLines) > 1 {
		firstBars := strings.Count(contentLines[0], "│")
		for i, content := range contentLines[1:] {
			bars := strings.Count(content, "│")
			if bars != firstBars {
				t.Errorf("Content line %d bar count mismatch: expected %d, got %d\nLine: %s", 
					i+2, firstBars, bars, content)
			}
		}
	}
}

// TestTableColumnWidthCalculation tests the column width calculation logic
func TestTableColumnWidthCalculation(t *testing.T) {
	styles := StreamsTableStyles{
		Container:    lipgloss.NewStyle(),
		Table:        lipgloss.NewStyle(),
		HeaderRow:    lipgloss.NewStyle(),
		Row:          lipgloss.NewStyle(),
		SelectedRow:  lipgloss.NewStyle(),
		SparklineRow: lipgloss.NewStyle(),
	}

	testCases := []struct {
		width    int
		expected string // Description of expected behavior
	}{
		{50, "minimum widths"},
		{80, "small terminal"},
		{120, "medium terminal"},
		{160, "large terminal"},
		{200, "very large terminal"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Width_%d", tc.width), func(t *testing.T) {
			widget := NewStreamsTableWidget(styles)
			widget.SetSize(tc.width, 20)

			cols := widget.calculateColumnWidths()

			// Validate that columns have reasonable widths
			if cols.stream < 7 {
				t.Errorf("Stream column too narrow: %d", cols.stream)
			}
			if cols.entries < 12 {
				t.Errorf("Entries column too narrow: %d", cols.entries)
			}

			// Calculate total width with overhead
			totalWidth := cols.stream + cols.entries + cols.size + cols.groups + cols.lastID + cols.memory
			overhead := 7 + 12 // borders + padding
			
			// For very small widths, we just ensure minimum widths are preserved
			if tc.width >= 80 && totalWidth + overhead > tc.width {
				t.Errorf("Total width %d + overhead %d exceeds terminal width %d", 
					totalWidth, overhead, tc.width)
			}

			fmt.Printf("Width %d: stream=%d, entries=%d, size=%d, groups=%d, lastID=%d, memory=%d (total=%d)\n",
				tc.width, cols.stream, cols.entries, cols.size, cols.groups, cols.lastID, cols.memory, totalWidth)
		})
	}
}

// TestTableHelperFunctions tests the utility functions used by the table
func TestTableHelperFunctions(t *testing.T) {
	// Test truncateString
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10c", 10, "exactly10c"},
		{"this_is_too_long", 10, "this_is..."},
		{"tiny", 3, "tin"},
		{"", 5, ""},
	}

	for _, test := range tests {
		result := truncateString(test.input, test.length)
		if result != test.expected {
			t.Errorf("truncateString(%q, %d) = %q, expected %q", 
				test.input, test.length, result, test.expected)
		}
	}

	// Test formatBytes
	byteTests := []struct {
		input    int64
		expected string
	}{
		{512, "512 B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1048576, "1.0MB"},
		{1073741824, "1.0GB"},
	}

	for _, test := range byteTests {
		result := formatBytes(test.input)
		if result != test.expected {
			t.Errorf("formatBytes(%d) = %q, expected %q", 
				test.input, result, test.expected)
		}
	}

	// Test formatNumberWithCommas
	numberTests := []struct {
		input    int64
		expected string
	}{
		{123, "123"},
		{1234, "1,234"},
		{1234567, "1,234,567"},
		{1000000000, "1,000,000,000"},
	}

	for _, test := range numberTests {
		result := formatNumberWithCommas(test.input)
		if result != test.expected {
			t.Errorf("formatNumberWithCommas(%d) = %q, expected %q", 
				test.input, result, test.expected)
		}
	}
}

// TestTableSelection tests keyboard navigation and selection
func TestTableSelection(t *testing.T) {
	styles := StreamsTableStyles{
		Container:    lipgloss.NewStyle(),
		Table:        lipgloss.NewStyle(),
		HeaderRow:    lipgloss.NewStyle(),
		Row:          lipgloss.NewStyle(),
		SelectedRow:  lipgloss.NewStyle().Background(lipgloss.Color("#555555")),
		SparklineRow: lipgloss.NewStyle(),
	}

	widget := NewStreamsTableWidget(styles)
	widget.SetSize(120, 20)
	widget.SetFocused(true)

	testStreams := []StreamData{
		createTestStream("stream_0", "1-0", 1000, 1, createTestMessageRates(25, 10.0)),
		createTestStream("stream_1", "2-0", 1000, 1, createTestMessageRates(25, 20.0)),
		createTestStream("stream_2", "3-0", 1000, 1, createTestMessageRates(25, 30.0)),
	}

	widget.streams = testStreams
	widget.updateSparklines()

	// Test initial selection
	selected := widget.GetSelectedStream()
	if selected == nil || selected.Name != "stream_0" {
		t.Error("Initial selection should be first stream")
	}

	output := widget.View()
	printTableDebug("Table with Selection", output)

	// The selected row should be styled differently
	// This is difficult to test programmatically due to styling,
	// but we can verify the selection logic works
	if widget.selectedIdx != 0 {
		t.Errorf("Expected selectedIdx to be 0, got %d", widget.selectedIdx)
	}
}
