package filecache

import (
	"fmt"
	"strings"
	"testing"
)

func TestDiffLineType_Constants(t *testing.T) {
	// Test that constants have expected values
	if DiffLineContext != 0 {
		t.Errorf("Expected DiffLineContext to be 0, got %d", DiffLineContext)
	}
	if DiffLineAdd != 1 {
		t.Errorf("Expected DiffLineAdd to be 1, got %d", DiffLineAdd)
	}
	if DiffLineRemove != 2 {
		t.Errorf("Expected DiffLineRemove to be 2, got %d", DiffLineRemove)
	}
	if DiffLineHeader != 3 {
		t.Errorf("Expected DiffLineHeader to be 3, got %d", DiffLineHeader)
	}
	if DiffLineLocation != 4 {
		t.Errorf("Expected DiffLineLocation to be 4, got %d", DiffLineLocation)
	}
	if DiffLineElided != 5 {
		t.Errorf("Expected DiffLineElided to be 5, got %d", DiffLineElided)
	}
}

func TestParseLineWithNumber(t *testing.T) {
	tests := []struct {
		name           string
		line           string
		isAddOrContext bool
		expectedOld    int
		expectedNew    int
		expectedContent string
	}{
		{
			name:           "add line with number",
			line:           "42:hello world",
			isAddOrContext: true,
			expectedOld:    0,
			expectedNew:    42,
			expectedContent: "hello world",
		},
		{
			name:           "remove line with number",
			line:           "15:goodbye world",
			isAddOrContext: false,
			expectedOld:    15,
			expectedNew:    0,
			expectedContent: "goodbye world",
		},
		{
			name:           "context line with number",
			line:           "100:unchanged line",
			isAddOrContext: true,
			expectedOld:    0,
			expectedNew:    100,
			expectedContent: "unchanged line",
		},
		{
			name:           "line without colon",
			line:           "no colon here",
			isAddOrContext: true,
			expectedOld:    0,
			expectedNew:    0,
			expectedContent: "no colon here",
		},
		{
			name:           "line with invalid number",
			line:           "abc:invalid number",
			isAddOrContext: true,
			expectedOld:    0,
			expectedNew:    0,
			expectedContent: "abc:invalid number",
		},
		{
			name:           "line with empty content",
			line:           "5:",
			isAddOrContext: true,
			expectedOld:    0,
			expectedNew:    5,
			expectedContent: "",
		},
		{
			name:           "line with multiple colons",
			line:           "10:hello:world:test",
			isAddOrContext: true,
			expectedOld:    0,
			expectedNew:    10,
			expectedContent: "hello:world:test",
		},
		{
			name:           "zero line number",
			line:           "0:zero line",
			isAddOrContext: true,
			expectedOld:    0,
			expectedNew:    0,
			expectedContent: "zero line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldLine, newLine, content := parseLineWithNumber(tt.line, tt.isAddOrContext)
			
			if oldLine != tt.expectedOld {
				t.Errorf("Expected oldLine %d, got %d", tt.expectedOld, oldLine)
			}
			if newLine != tt.expectedNew {
				t.Errorf("Expected newLine %d, got %d", tt.expectedNew, newLine)
			}
			if content != tt.expectedContent {
				t.Errorf("Expected content %q, got %q", tt.expectedContent, content)
			}
		})
	}
}

func TestParseUnifiedDiff(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		expected []DiffLine
	}{
		{
			name: "simple diff with line numbers",
			diff: `--- file.txt (cached)
+++ file.txt (new write)
 1:unchanged line
-2:old line
+3:new line`,
			expected: []DiffLine{
				{Type: DiffLineHeader, Content: "--- file.txt (cached)", OldLine: 0, NewLine: 0},
				{Type: DiffLineHeader, Content: "+++ file.txt (new write)", OldLine: 0, NewLine: 0},
				{Type: DiffLineContext, Content: " 1:unchanged line", OldLine: 0, NewLine: 1},
				{Type: DiffLineRemove, Content: "-2:old line", OldLine: 2, NewLine: 0},
				{Type: DiffLineAdd, Content: "+3:new line", OldLine: 0, NewLine: 3},
			},
		},
		{
			name: "empty diff",
			diff: "",
			expected: []DiffLine{},
		},
		{
			name: "diff with only headers",
			diff: `--- file.txt (cached)
+++ file.txt (new write)`,
			expected: []DiffLine{
				{Type: DiffLineHeader, Content: "--- file.txt (cached)", OldLine: 0, NewLine: 0},
				{Type: DiffLineHeader, Content: "+++ file.txt (new write)", OldLine: 0, NewLine: 0},
			},
		},
		{
			name: "diff with lines without numbers",
			diff: `--- file.txt (cached)
+++ file.txt (new write)
 unchanged line
-old line
+new line`,
			expected: []DiffLine{
				{Type: DiffLineHeader, Content: "--- file.txt (cached)", OldLine: 0, NewLine: 0},
				{Type: DiffLineHeader, Content: "+++ file.txt (new write)", OldLine: 0, NewLine: 0},
				{Type: DiffLineContext, Content: " unchanged line", OldLine: 0, NewLine: 0},
				{Type: DiffLineRemove, Content: "-old line", OldLine: 0, NewLine: 0},
				{Type: DiffLineAdd, Content: "+new line", OldLine: 0, NewLine: 0},
			},
		},
		{
			name: "diff with unrecognized lines",
			diff: `--- file.txt (cached)
+++ file.txt (new write)
random line without prefix`,
			expected: []DiffLine{
				{Type: DiffLineHeader, Content: "--- file.txt (cached)", OldLine: 0, NewLine: 0},
				{Type: DiffLineHeader, Content: "+++ file.txt (new write)", OldLine: 0, NewLine: 0},
				{Type: DiffLineContext, Content: "random line without prefix", OldLine: 0, NewLine: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseUnifiedDiff(tt.diff)
			
			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d lines, got %d", len(tt.expected), len(result))
			}
			
			for i, expected := range tt.expected {
				if result[i].Type != expected.Type {
					t.Errorf("Line %d: expected type %d, got %d", i, expected.Type, result[i].Type)
				}
				if result[i].Content != expected.Content {
					t.Errorf("Line %d: expected content %q, got %q", i, expected.Content, result[i].Content)
				}
				if result[i].OldLine != expected.OldLine {
					t.Errorf("Line %d: expected oldLine %d, got %d", i, expected.OldLine, result[i].OldLine)
				}
				if result[i].NewLine != expected.NewLine {
					t.Errorf("Line %d: expected newLine %d, got %d", i, expected.NewLine, result[i].NewLine)
				}
			}
		})
	}
}

func TestGenerateBasicUnifiedDiff(t *testing.T) {
	tests := []struct {
		name     string
		oldLines []string
		newLines []string
		filename string
		expected string
	}{
		{
			name:     "simple change",
			oldLines: []string{"line1", "old line", "line3"},
			newLines: []string{"line1", "new line", "line3"},
			filename: "test.txt",
			expected: `--- test.txt (cached)
+++ test.txt (new write)
 line1
-old line
+new line
 line3
`,
		},
		{
			name:     "addition",
			oldLines: []string{"line1", "line2"},
			newLines: []string{"line1", "line2", "line3"},
			filename: "test.txt",
			expected: `--- test.txt (cached)
+++ test.txt (new write)
 line1
 line2
+line3
`,
		},
		{
			name:     "deletion",
			oldLines: []string{"line1", "line2", "line3"},
			newLines: []string{"line1", "line3"},
			filename: "test.txt",
			expected: `--- test.txt (cached)
+++ test.txt (new write)
 line1
-line2
+line3
-line3
`,
		},
		{
			name:     "empty files",
			oldLines: []string{},
			newLines: []string{},
			filename: "empty.txt",
			expected: `--- empty.txt (cached)
+++ empty.txt (new write)
`,
		},
		{
			name:     "new file",
			oldLines: []string{},
			newLines: []string{"new content"},
			filename: "new.txt",
			expected: `--- new.txt (cached)
+++ new.txt (new write)
+new content
`,
		},
		{
			name:     "deleted file",
			oldLines: []string{"old content"},
			newLines: []string{},
			filename: "deleted.txt",
			expected: `--- deleted.txt (cached)
+++ deleted.txt (new write)
-old content
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateBasicUnifiedDiff(tt.oldLines, tt.newLines, tt.filename)
			
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestElideUnifiedDiff(t *testing.T) {
	tests := []struct {
		name         string
		diff         string
		contextLines int
		expected     string
	}{
		{
			name: "elide with 1 context line",
			diff: `--- file.txt (cached)
+++ file.txt (new write)
 1:line1
 2:line2
 3:line3
-4:old line
+5:new line
 6:line6
 7:line7
 8:line8`,
			contextLines: 1,
			expected: `--- file.txt (cached)
+++ file.txt (new write)
 3:line3
-4:old line
+5:new line
 6:line6
`,
		},
		{
			name: "elide with 0 context lines",
			diff: `--- file.txt (cached)
+++ file.txt (new write)
 1:line1
 2:line2
-3:old line
+4:new line
 5:line5
 6:line6`,
			contextLines: 0,
			expected: `--- file.txt (cached)
+++ file.txt (new write)
-3:old line
+4:new line
`,
		},
		{
			name: "no elision needed - all lines kept",
			diff: `--- file.txt (cached)
+++ file.txt (new write)
 1:line1
-2:old line
+3:new line
 4:line4`,
			contextLines: 2,
			expected: `--- file.txt (cached)
+++ file.txt (new write)
 1:line1
-2:old line
+3:new line
 4:line4
`,
		},
		{
			name: "negative context lines returns original",
			diff: `--- file.txt (cached)
+++ file.txt (new write)
 1:line1
-2:old line
+3:new line`,
			contextLines: -1,
			expected: `--- file.txt (cached)
+++ file.txt (new write)
 1:line1
-2:old line
+3:new line`,
		},
		{
			name: "empty diff",
			diff: "",
			contextLines: 1,
			expected: "",
		},
		{
			name: "diff with no changes",
			diff: `--- file.txt (cached)
+++ file.txt (new write)
 1:line1
 2:line2
 3:line3`,
			contextLines: 1,
			expected: `--- file.txt (cached)
+++ file.txt (new write)
 1:line1
 2:line2
 3:line3`,
		},
		{
			name: "multiple change groups with elision",
			diff: `--- file.txt (cached)
+++ file.txt (new write)
 1:line1
-2:old1
+3:new1
 4:line4
 5:line5
 6:line6
 7:line7
-8:old2
+9:new2
 10:line10`,
			contextLines: 1,
			expected: `--- file.txt (cached)
+++ file.txt (new write)
 1:line1
-2:old1
+3:new1
 4:line4
...
 7:line7
-8:old2
+9:new2
 10:line10
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ElideUnifiedDiff(tt.diff, tt.contextLines)
			
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestGenerateElidedUnifiedDiff(t *testing.T) {
	tests := []struct {
		name         string
		oldContent   []byte
		newContent   []byte
		filename     string
		contextLines int
		contains     []string // Check if result contains these substrings
	}{
		{
			name:         "simple diff with elision",
			oldContent:   []byte("line1\nold line\nline3\nline4\nline5"),
			newContent:   []byte("line1\nnew line\nline3\nline4\nline5"),
			filename:     "test.txt",
			contextLines: 1,
			contains:     []string{"--- test.txt (cached)", "+++ test.txt (new write)", "-old line", "+new line"},
		},
		{
			name:         "no elision when contextLines is 0",
			oldContent:   []byte("line1\nold line\nline3"),
			newContent:   []byte("line1\nnew line\nline3"),
			filename:     "test.txt",
			contextLines: 0,
			contains:     []string{"--- test.txt (cached)", "+++ test.txt (new write)", "-old line", "+new line"},
		},
		{
			name:         "empty old content",
			oldContent:   []byte(""),
			newContent:   []byte("new content"),
			filename:     "new.txt",
			contextLines: 1,
			contains:     []string{"--- new.txt (cached)", "+++ new.txt (new write)", "+new content"},
		},
		{
			name:         "empty new content",
			oldContent:   []byte("old content"),
			newContent:   []byte(""),
			filename:     "deleted.txt",
			contextLines: 1,
			contains:     []string{"--- deleted.txt (cached)", "+++ deleted.txt (new write)", "-old content"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateElidedUnifiedDiff(tt.oldContent, tt.newContent, tt.filename, tt.contextLines)
			
			for _, substring := range tt.contains {
				if !strings.Contains(result, substring) {
					t.Errorf("Expected result to contain %q, but it didn't. Result:\n%s", substring, result)
				}
			}
		})
	}
}

func TestElideUnifiedDiff_EdgeCases(t *testing.T) {
	t.Run("diff with only headers", func(t *testing.T) {
		diff := `--- file.txt (cached)
+++ file.txt (new write)`
		
		result := ElideUnifiedDiff(diff, 1)
		expected := `--- file.txt (cached)
+++ file.txt (new write)`
		
		if result != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
		}
	})

	t.Run("diff with changes at start", func(t *testing.T) {
		diff := `--- file.txt (cached)
+++ file.txt (new write)
-1:old line
+2:new line
 3:line3
 4:line4
 5:line5`
		
		result := ElideUnifiedDiff(diff, 1)
		expected := `--- file.txt (cached)
+++ file.txt (new write)
-1:old line
+2:new line
 3:line3
`
		
		if result != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
		}
	})

	t.Run("diff with changes at end", func(t *testing.T) {
		diff := `--- file.txt (cached)
+++ file.txt (new write)
 1:line1
 2:line2
 3:line3
-4:old line
+5:new line`
		
		result := ElideUnifiedDiff(diff, 1)
		expected := `--- file.txt (cached)
+++ file.txt (new write)
 3:line3
-4:old line
+5:new line
`
		
		if result != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
		}
	})
}

func TestParseUnifiedDiff_Integration(t *testing.T) {
	// Test the round-trip: generate diff -> parse -> elide -> should be valid
	diff := `--- test.txt (cached)
+++ test.txt (new write)
 1:unchanged1
 2:unchanged2
 3:unchanged3
 4:unchanged4
 5:unchanged5
 6:unchanged6
 7:unchanged7
 8:unchanged8
 9:unchanged9
-10:old line
+11:new line
 12:unchanged12
 13:unchanged13
 14:unchanged14
 15:unchanged15
 16:unchanged16`

	// Parse the diff
	parsed := ParseUnifiedDiff(diff)
	
	// Verify we have the expected structure
	if len(parsed) != 18 {
		for i, line := range parsed {
			t.Logf("Line %d: Type=%d, Content=%q", i, line.Type, line.Content)
		}
		t.Fatalf("Expected 18 parsed lines, got %d", len(parsed))
	}
	
	// Check headers
	if parsed[0].Type != DiffLineHeader || !strings.Contains(parsed[0].Content, "---") {
		t.Errorf("First line should be header with ---, got %s", parsed[0].Content)
	}
	if parsed[1].Type != DiffLineHeader || !strings.Contains(parsed[1].Content, "+++") {
		t.Errorf("Second line should be header with +++, got %s", parsed[1].Content)
	}
	
	// Check content lines
	if parsed[2].Type != DiffLineContext || parsed[2].NewLine != 1 {
		t.Errorf("Third line should be context with line 1, got type %d, line %d", parsed[2].Type, parsed[2].NewLine)
	}
	if parsed[11].Type != DiffLineRemove || parsed[11].OldLine != 10 {
		t.Errorf("Twelfth line should be remove with old line 10, got type %d, line %d", parsed[11].Type, parsed[11].OldLine)
	}
	if parsed[12].Type != DiffLineAdd || parsed[12].NewLine != 11 {
		t.Errorf("Thirteenth line should be add with new line 11, got type %d, line %d", parsed[12].Type, parsed[12].NewLine)
	}
	
	// Test elision works with the parsed content
	elided := ElideUnifiedDiff(diff, 0)
	// With context=0, only changed lines are shown
	if !strings.Contains(elided, "-10:old line") {
		t.Error("Elided diff should contain the removed line")
	}
	if !strings.Contains(elided, "+11:new line") {
		t.Error("Elided diff should contain the added line")
	}
}

// Benchmark tests for performance
func BenchmarkParseUnifiedDiff(b *testing.B) {
	diff := `--- file.txt (cached)
+++ file.txt (new write)
 1:line1
 2:line2
-3:old line
+4:new line
 5:line5
 6:line6`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseUnifiedDiff(diff)
	}
}

func BenchmarkElideUnifiedDiff(b *testing.B) {
	// Create a larger diff for benchmarking
	var lines []string
	lines = append(lines, "--- file.txt (cached)")
	lines = append(lines, "+++ file.txt (new write)")
	
	for i := 1; i <= 100; i++ {
		if i == 50 {
			lines = append(lines, "-50:old line")
			lines = append(lines, "+51:new line")
		} else {
			lines = append(lines, fmt.Sprintf(" %d:line%d", i, i))
		}
	}
	
	diff := strings.Join(lines, "\n")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ElideUnifiedDiff(diff, 3)
	}
}

func BenchmarkParseLineWithNumber(b *testing.B) {
	line := "42:this is a test line with some content"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseLineWithNumber(line, true)
	}
}
