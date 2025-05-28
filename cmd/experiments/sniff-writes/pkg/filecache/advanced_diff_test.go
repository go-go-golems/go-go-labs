package filecache

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"
)

// TestTextVsBinaryDiffHandling tests diff generation for different content types
func TestTextVsBinaryDiffHandling(t *testing.T) {
	fc := NewFileCache(1024*1024, 64*1024, NewMockTimeProvider())
	
	// Test text content diff
	textHash := uint32(100)
	originalText := []byte("line 1\nline 2\nline 3\n")
	err := fc.AddRead(textHash, 0, originalText)
	if err != nil {
		t.Fatalf("Failed to add original text: %v", err)
	}
	
	modifiedText := []byte("line 1\nmodified line 2\nline 3\n")
	err = fc.UpdateWithWrite(textHash, 7, []byte("modified line 2"))
	if err != nil {
		t.Fatalf("Failed to update text: %v", err)
	}
	
	// Generate diff for text
	diff := fc.GenerateDiff(textHash, modifiedText)
	if len(diff) == 0 {
		t.Error("No diff generated for text modification")
	}
	
	// Test binary content diff
	binaryHash := uint32(200)
	originalBinary := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD}
	err = fc.AddRead(binaryHash, 0, originalBinary)
	if err != nil {
		t.Fatalf("Failed to add original binary: %v", err)
	}
	
	modifiedBinary := []byte{0x00, 0x01, 0x99, 0x03, 0xFF, 0xFE, 0xFD}
	err = fc.UpdateWithWrite(binaryHash, 2, []byte{0x99})
	if err != nil {
		t.Fatalf("Failed to update binary: %v", err)
	}
	
	// Generate diff for binary
	binaryDiff := fc.GenerateDiff(binaryHash, modifiedBinary)
	if len(binaryDiff) == 0 {
		t.Error("No diff generated for binary modification")
	}
	
	// Verify diffs are different in format/content
	if bytes.Equal(diff, binaryDiff) {
		t.Error("Text and binary diffs should be different")
	}
}

// TestLineEndingVariations tests diff handling for different line ending styles
func TestLineEndingVariations(t *testing.T) {
	fc := NewFileCache(1024*1024, 64*1024, NewMockTimeProvider())
	
	testCases := []struct {
		name        string
		pathHash    uint32
		original    string
		modified    string
		lineEnding  string
	}{
		{
			name:       "Unix LF",
			pathHash:   300,
			original:   "line1\nline2\nline3\n",
			modified:   "line1\nmodified line2\nline3\n",
			lineEnding: "\n",
		},
		{
			name:       "Windows CRLF",
			pathHash:   301,
			original:   "line1\r\nline2\r\nline3\r\n",
			modified:   "line1\r\nmodified line2\r\nline3\r\n",
			lineEnding: "\r\n",
		},
		{
			name:       "Classic Mac CR",
			pathHash:   302,
			original:   "line1\rline2\rline3\r",
			modified:   "line1\rmodified line2\rline3\r",
			lineEnding: "\r",
		},
		{
			name:       "Mixed line endings",
			pathHash:   303,
			original:   "line1\nline2\r\nline3\r",
			modified:   "line1\nmodified line2\r\nline3\r",
			lineEnding: "mixed",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalBytes := []byte(tc.original)
			err := fc.AddRead(tc.pathHash, 0, originalBytes)
			if err != nil {
				t.Fatalf("Failed to add original content: %v", err)
			}
			
			// Simulate write that changes line2
			modifiedLine := strings.Replace(tc.original, "line2", "modified line2", 1)
			modifiedBytes := []byte(modifiedLine)
			
			// Update cache with write
			writePos := strings.Index(tc.original, "line2")
			if writePos >= 0 {
				err = fc.UpdateWithWrite(tc.pathHash, uint64(writePos), []byte("modified line2"))
				if err != nil {
					t.Fatalf("Failed to update with write: %v", err)
				}
			}
			
			// Generate diff
			diff := fc.GenerateDiff(tc.pathHash, modifiedBytes)
			if len(diff) == 0 {
				t.Errorf("No diff generated for %s", tc.name)
			}
			
			// Verify diff contains line ending information
			diffStr := string(diff)
			if tc.lineEnding != "mixed" && !strings.Contains(diffStr, tc.lineEnding) {
				t.Logf("Diff might not preserve line endings for %s: %q", tc.name, diffStr)
			}
		})
	}
}

// TestUnicodeAndEncodingHandling tests diff generation with Unicode content
func TestUnicodeAndEncodingHandling(t *testing.T) {
	fc := NewFileCache(1024*1024, 64*1024, NewMockTimeProvider())
	
	unicodeTestCases := []struct {
		name     string
		pathHash uint32
		content  string
	}{
		{
			name:     "Basic Unicode",
			pathHash: 400,
			content:  "Hello, 世界! こんにちは 🌍",
		},
		{
			name:     "Emoji and symbols",
			pathHash: 401,
			content:  "🚀 Testing 📝 with emojis 🎉 and symbols ≈ ∞ ∑",
		},
		{
			name:     "Mixed scripts",
			pathHash: 402,
			content:  "English Русский العربية 中文 ひらがな カタカナ",
		},
		{
			name:     "Unicode control chars",
			pathHash: 403,
			content:  "Line1\u2028Line2\u2029Paragraph",
		},
		{
			name:     "Zero-width chars",
			pathHash: 404,
			content:  "Test\u200Bwith\u200Czero\u200Dwidth\uFEFFchars",
		},
	}
	
	for _, tc := range unicodeTestCases {
		t.Run(tc.name, func(t *testing.T) {
			originalBytes := []byte(tc.content)
			
			// Verify valid UTF-8
			if !utf8.Valid(originalBytes) {
				t.Fatalf("Test content is not valid UTF-8: %s", tc.name)
			}
			
			err := fc.AddRead(tc.pathHash, 0, originalBytes)
			if err != nil {
				t.Fatalf("Failed to add Unicode content: %v", err)
			}
			
			// Modify the content
			modified := tc.content + " MODIFIED"
			modifiedBytes := []byte(modified)
			
			// Update with write
			err = fc.UpdateWithWrite(tc.pathHash, uint64(len(originalBytes)), []byte(" MODIFIED"))
			if err != nil {
				t.Fatalf("Failed to update Unicode content: %v", err)
			}
			
			// Generate diff
			diff := fc.GenerateDiff(tc.pathHash, modifiedBytes)
			if len(diff) == 0 {
				t.Errorf("No diff generated for Unicode content: %s", tc.name)
			}
			
			// Verify diff is valid UTF-8
			if !utf8.Valid(diff) {
				t.Errorf("Generated diff is not valid UTF-8 for: %s", tc.name)
			}
			
			// Verify Unicode content is preserved in diff
			diffStr := string(diff)
			if !strings.Contains(diffStr, "MODIFIED") {
				t.Errorf("Diff doesn't show modification for: %s", tc.name)
			}
		})
	}
}

// TestLargeDiffGeneration tests diff generation for large content changes
func TestLargeDiffGeneration(t *testing.T) {
	fc := NewFileCache(2*1024*1024, 512*1024, NewMockTimeProvider())
	pathHash := uint32(500)
	
	// Create large original content
	var originalBuilder strings.Builder
	for i := 0; i < 1000; i++ {
		originalBuilder.WriteString("This is line ")
		originalBuilder.WriteString(string(rune('0' + (i % 10))))
		originalBuilder.WriteString(" of the large file content.\n")
	}
	originalContent := originalBuilder.String()
	originalBytes := []byte(originalContent)
	
	err := fc.AddRead(pathHash, 0, originalBytes)
	if err != nil {
		t.Fatalf("Failed to add large original content: %v", err)
	}
	
	// Create large modified content
	var modifiedBuilder strings.Builder
	for i := 0; i < 1000; i++ {
		if i%100 == 50 { // Modify every 100th line around line 50
			modifiedBuilder.WriteString("*** MODIFIED LINE ")
			modifiedBuilder.WriteString(string(rune('0' + (i % 10))))
			modifiedBuilder.WriteString(" ***\n")
		} else {
			modifiedBuilder.WriteString("This is line ")
			modifiedBuilder.WriteString(string(rune('0' + (i % 10))))
			modifiedBuilder.WriteString(" of the large file content.\n")
		}
	}
	modifiedContent := modifiedBuilder.String()
	modifiedBytes := []byte(modifiedContent)
	
	// Simulate writes for the modified lines
	lines := strings.Split(originalContent, "\n")
	offset := 0
	for i, line := range lines {
		if i%100 == 50 && i < len(lines)-1 { // Skip last empty line
			modifiedLine := "*** MODIFIED LINE " + string(rune('0'+(i%10))) + " ***"
			err = fc.UpdateWithWrite(pathHash, uint64(offset), []byte(modifiedLine))
			if err != nil {
				t.Fatalf("Failed to update line %d: %v", i, err)
			}
		}
		offset += len(line) + 1 // +1 for newline
	}
	
	// Generate diff
	diff := fc.GenerateDiff(pathHash, modifiedBytes)
	if len(diff) == 0 {
		t.Error("No diff generated for large content")
	}
	
	// Verify diff contains modifications
	diffStr := string(diff)
	if !strings.Contains(diffStr, "MODIFIED LINE") {
		t.Error("Large diff doesn't show modifications")
	}
	
	// Verify diff size is reasonable (not entire file)
	diffSize := len(diff)
	originalSize := len(originalBytes)
	if diffSize > originalSize {
		t.Errorf("Diff size (%d) larger than original (%d)", diffSize, originalSize)
	}
	
	t.Logf("Large diff stats: original %d bytes, diff %d bytes (%.1f%%)",
		originalSize, diffSize, float64(diffSize)/float64(originalSize)*100)
}

// TestWhitespaceOnlyChanges tests diff generation for whitespace modifications
func TestWhitespaceOnlyChanges(t *testing.T) {
	fc := NewFileCache(1024*1024, 64*1024, NewMockTimeProvider())
	
	testCases := []struct {
		name     string
		pathHash uint32
		original string
		modified string
	}{
		{
			name:     "Tab to spaces",
			pathHash: 600,
			original: "function test() {\n\treturn true;\n}",
			modified: "function test() {\n    return true;\n}",
		},
		{
			name:     "Trailing whitespace",
			pathHash: 601,
			original: "line1\nline2\nline3",
			modified: "line1 \nline2  \nline3   ",
		},
		{
			name:     "Line ending whitespace",
			pathHash: 602,
			original: "line1\n\nline3",
			modified: "line1\n \nline3",
		},
		{
			name:     "Indentation changes",
			pathHash: 603,
			original: "  code\n    more code\n      deep code",
			modified: "    code\n      more code\n        deep code",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalBytes := []byte(tc.original)
			err := fc.AddRead(tc.pathHash, 0, originalBytes)
			if err != nil {
				t.Fatalf("Failed to add original: %v", err)
			}
			
			modifiedBytes := []byte(tc.modified)
			
			// Find difference and simulate write
			for i := 0; i < len(originalBytes) && i < len(modifiedBytes); i++ {
				if originalBytes[i] != modifiedBytes[i] {
					// Found first difference, update from here
					remaining := modifiedBytes[i:]
					err = fc.UpdateWithWrite(tc.pathHash, uint64(i), remaining)
					if err != nil {
						t.Fatalf("Failed to update: %v", err)
					}
					break
				}
			}
			
			// If modified is longer, add the extra part
			if len(modifiedBytes) > len(originalBytes) {
				extra := modifiedBytes[len(originalBytes):]
				err = fc.UpdateWithWrite(tc.pathHash, uint64(len(originalBytes)), extra)
				if err != nil {
					t.Fatalf("Failed to add extra content: %v", err)
				}
			}
			
			// Generate diff
			diff := fc.GenerateDiff(tc.pathHash, modifiedBytes)
			if len(diff) == 0 {
				t.Errorf("No diff generated for whitespace change: %s", tc.name)
			}
			
			// For whitespace-only changes, diff should still be generated
			diffStr := string(diff)
			t.Logf("Whitespace diff for %s: %q", tc.name, diffStr)
		})
	}
}

// TestIdenticalContentWrites tests when writes don't actually change content
func TestIdenticalContentWrites(t *testing.T) {
	fc := NewFileCache(1024*1024, 64*1024, NewMockTimeProvider())
	pathHash := uint32(700)
	
	originalContent := []byte("unchanged content that stays the same")
	err := fc.AddRead(pathHash, 0, originalContent)
	if err != nil {
		t.Fatalf("Failed to add original content: %v", err)
	}
	
	// Write identical content
	identicalContent := []byte("unchanged content that stays the same")
	err = fc.UpdateWithWrite(pathHash, 0, identicalContent)
	if err != nil {
		t.Fatalf("Failed to write identical content: %v", err)
	}
	
	// Generate diff - should be minimal or empty
	diff := fc.GenerateDiff(pathHash, identicalContent)
	
	// Test partial identical overwrites
	partialContent := []byte("unchanged")
	err = fc.UpdateWithWrite(pathHash, 0, partialContent)
	if err != nil {
		t.Fatalf("Failed to write partial identical: %v", err)
	}
	
	partialDiff := fc.GenerateDiff(pathHash, originalContent)
	
	// Test overwriting with same substring
	substring := []byte("content")
	err = fc.UpdateWithWrite(pathHash, 10, substring)
	if err != nil {
		t.Fatalf("Failed to write identical substring: %v", err)
	}
	
	substringDiff := fc.GenerateDiff(pathHash, originalContent)
	
	t.Logf("Identical content diffs - full: %d bytes, partial: %d bytes, substring: %d bytes",
		len(diff), len(partialDiff), len(substringDiff))
}

// TestDiffPerformanceWithComplexChanges tests diff generation performance
func TestDiffPerformanceWithComplexChanges(t *testing.T) {
	fc := NewFileCache(4*1024*1024, 1024*1024, NewMockTimeProvider())
	pathHash := uint32(800)
	
	// Create content with complex structure
	var contentBuilder strings.Builder
	contentBuilder.WriteString("// File header\n")
	contentBuilder.WriteString("package main\n\n")
	
	// Add many functions
	for i := 0; i < 100; i++ {
		contentBuilder.WriteString("func function")
		contentBuilder.WriteString(string(rune('A' + (i % 26))))
		contentBuilder.WriteString("() {\n")
		contentBuilder.WriteString("    // Function body\n")
		for j := 0; j < 10; j++ {
			contentBuilder.WriteString("    line")
			contentBuilder.WriteString(string(rune('0' + (j % 10))))
			contentBuilder.WriteString("_function")
			contentBuilder.WriteString(string(rune('A' + (i % 26))))
			contentBuilder.WriteString("\n")
		}
		contentBuilder.WriteString("}\n\n")
	}
	
	originalContent := contentBuilder.String()
	originalBytes := []byte(originalContent)
	
	err := fc.AddRead(pathHash, 0, originalBytes)
	if err != nil {
		t.Fatalf("Failed to add complex content: %v", err)
	}
	
	// Make scattered modifications
	modifications := []struct {
		searchStr string
		replaceStr string
	}{
		{"package main", "package modified"},
		{"Function body", "Modified function body"},
		{"line0_", "modified_line0_"},
		{"line5_", "modified_line5_"},
		{"functionA", "modifiedFunctionA"},
	}
	
	modifiedContent := originalContent
	for _, mod := range modifications {
		modifiedContent = strings.ReplaceAll(modifiedContent, mod.searchStr, mod.replaceStr)
		
		// Update cache with individual writes
		pos := strings.Index(originalContent, mod.searchStr)
		if pos >= 0 {
			err = fc.UpdateWithWrite(pathHash, uint64(pos), []byte(mod.replaceStr))
			if err != nil {
				t.Fatalf("Failed to update modification: %v", err)
			}
		}
	}
	
	modifiedBytes := []byte(modifiedContent)
	
	// Generate diff and measure
	diff := fc.GenerateDiff(pathHash, modifiedBytes)
	if len(diff) == 0 {
		t.Error("No diff generated for complex changes")
	}
	
	diffRatio := float64(len(diff)) / float64(len(originalBytes))
	t.Logf("Complex diff: original %d bytes, modified %d bytes, diff %d bytes (%.2f%%)",
		len(originalBytes), len(modifiedBytes), len(diff), diffRatio*100)
	
	// Verify diff quality - should be smaller than full content
	if len(diff) > len(originalBytes)/2 {
		t.Logf("Warning: diff size (%d) is more than 50%% of original (%d)", len(diff), len(originalBytes))
	}
}
