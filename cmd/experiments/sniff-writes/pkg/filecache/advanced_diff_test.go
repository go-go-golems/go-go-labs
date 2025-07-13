package filecache

import (
	"bytes"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestTextVsBinaryDiffHandling(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	// Test text file diff
	textPathHash := uint32(11111)
	textContent := []byte("This is a text file\nwith multiple lines\nfor testing diffs")
	fc.AddRead(textPathHash, 0, textContent)

	newTextContent := []byte("This is a modified text file\nwith multiple lines\nfor testing diffs")
	diff, hasDiff := fc.GenerateDiff(1234, 5, textPathHash, 0, newTextContent)

	if !hasDiff {
		t.Error("Expected diff to be generated for text file")
	}

	if len(diff) == 0 {
		t.Error("Expected non-empty diff for text file")
	}

	// Diff should contain recognizable diff markers
	if !strings.Contains(diff, "---") || !strings.Contains(diff, "+++") {
		t.Error("Expected diff to contain unified diff markers")
	}

	// Test binary file diff
	binaryPathHash := uint32(22222)
	binaryContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D} // PNG header
	fc.AddRead(binaryPathHash, 0, binaryContent)

	newBinaryContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0E} // Modified PNG
	binaryDiff, hasBinaryDiff := fc.GenerateDiff(1234, 5, binaryPathHash, 0, newBinaryContent)

	if !hasBinaryDiff {
		t.Error("Expected diff to be generated for binary file")
	}

	if len(binaryDiff) == 0 {
		t.Error("Expected non-empty diff for binary file")
	}

	// Binary diff should handle non-printable characters gracefully
	// (implementation specific - might show hex representation or indicate binary)
	t.Logf("Binary diff: %s", binaryDiff)
}

func TestLineEndingVariations(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	pathHash := uint32(33333)

	// Test Unix line endings (\n)
	unixContent := []byte("Line 1\nLine 2\nLine 3\n")
	fc.AddRead(pathHash, 0, unixContent)

	// Compare with Windows line endings (\r\n)
	windowsContent := []byte("Line 1\r\nLine 2\r\nLine 3\r\n")
	diff, hasDiff := fc.GenerateDiff(1234, 5, pathHash, 0, windowsContent)

	if !hasDiff {
		t.Error("Expected diff between Unix and Windows line endings")
	}

	if len(diff) == 0 {
		t.Error("Expected non-empty diff for line ending differences")
	}

	// Clear cache and test Mac line endings (\r)
	fc.UpdateWithWrite(pathHash, 0, []byte("reset"))
	fc.AddRead(pathHash, 0, unixContent)

	macContent := []byte("Line 1\rLine 2\rLine 3\r")
	macDiff, hasMacDiff := fc.GenerateDiff(1234, 5, pathHash, 0, macContent)

	if !hasMacDiff {
		t.Error("Expected diff between Unix and Mac line endings")
	}

	if len(macDiff) == 0 {
		t.Error("Expected non-empty diff for Mac line ending differences")
	}

	t.Logf("Windows diff: %s", diff)
	t.Logf("Mac diff: %s", macDiff)
}

func TestUnicodeAndEncodingEdgeCases(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	pathHash := uint32(44444)

	// Test UTF-8 content with various Unicode characters
	unicodeContent := []byte("Hello 世界 🌍 café naïve résumé")
	fc.AddRead(pathHash, 0, unicodeContent)

	// Verify content is valid UTF-8
	if !utf8.Valid(unicodeContent) {
		t.Fatal("Test Unicode content is not valid UTF-8")
	}

	// Modify with different Unicode characters
	modifiedUnicode := []byte("Hello 世界 🌎 café naïve résumé")
	diff, hasDiff := fc.GenerateDiff(1234, 5, pathHash, 0, modifiedUnicode)

	if !hasDiff {
		t.Error("Expected diff for Unicode character change")
	}

	if len(diff) == 0 {
		t.Error("Expected non-empty diff for Unicode changes")
	}

	// Test with emoji changes
	fc.UpdateWithWrite(pathHash, 0, []byte("reset"))
	emojiContent := []byte("Test with emojis: 😀 😃 😄 😁")
	fc.AddRead(pathHash, 0, emojiContent)

	modifiedEmoji := []byte("Test with emojis: 😀 😃 😄 😂")
	emojiDiff, hasEmojiDiff := fc.GenerateDiff(1234, 5, pathHash, 0, modifiedEmoji)

	if !hasEmojiDiff {
		t.Error("Expected diff for emoji change")
	}

	// Test with invalid UTF-8 sequences (binary data that looks like text)
	fc.UpdateWithWrite(pathHash, 0, []byte("reset"))
	invalidUTF8 := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0xFF, 0xFE, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64}
	fc.AddRead(pathHash, 0, invalidUTF8)

	modifiedInvalid := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0xFF, 0xFD, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64}
	invalidDiff, hasInvalidDiff := fc.GenerateDiff(1234, 5, pathHash, 0, modifiedInvalid)

	if !hasInvalidDiff {
		t.Error("Expected diff for invalid UTF-8 sequence change")
	}

	t.Logf("Unicode diff: %s", diff)
	t.Logf("Emoji diff: %s", emojiDiff)
	t.Logf("Invalid UTF-8 diff: %s", invalidDiff)
}

func TestVeryLargeDiffs(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(512*1024, 1024*1024, time.Hour, mockTime)

	pathHash := uint32(55555)

	// Create large content (64KB)
	largeContent := make([]byte, 64*1024)
	for i := range largeContent {
		largeContent[i] = byte('A' + (i % 26)) // Repeating alphabet
	}
	fc.AddRead(pathHash, 0, largeContent)

	// Modify large portion of content
	modifiedLargeContent := make([]byte, 64*1024)
	copy(modifiedLargeContent, largeContent)

	// Change middle 32KB to different pattern
	start := 16 * 1024
	end := 48 * 1024
	for i := start; i < end; i++ {
		modifiedLargeContent[i] = byte('a' + ((i - start) % 26)) // Lowercase alphabet
	}

	diff, hasDiff := fc.GenerateDiff(1234, 5, pathHash, 0, modifiedLargeContent)

	if !hasDiff {
		t.Error("Expected diff for large content change")
	}

	if len(diff) == 0 {
		t.Error("Expected non-empty diff for large content")
	}

	// Diff should be manageable size (not necessarily small, but not huge)
	if len(diff) > 128*1024 {
		t.Errorf("Diff too large: %d bytes (consider implementing diff elision)", len(diff))
	}

	// Test complete content replacement
	completelyDifferent := make([]byte, 64*1024)
	for i := range completelyDifferent {
		completelyDifferent[i] = byte('Z')
	}

	fc.UpdateWithWrite(pathHash, 0, largeContent) // Reset
	fc.AddRead(pathHash, 0, largeContent)

	replacementDiff, hasReplacementDiff := fc.GenerateDiff(1234, 5, pathHash, 0, completelyDifferent)

	if !hasReplacementDiff {
		t.Error("Expected diff for complete content replacement")
	}

	t.Logf("Large diff size: %d bytes", len(diff))
	t.Logf("Replacement diff size: %d bytes", len(replacementDiff))
}

func TestWhitespaceOnlyChanges(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	pathHash := uint32(66666)

	// Original content with specific whitespace
	originalContent := []byte("function test() {\n    return true;\n}")
	fc.AddRead(pathHash, 0, originalContent)

	// Change indentation from spaces to tabs
	tabContent := []byte("function test() {\n\treturn true;\n}")
	diff, hasDiff := fc.GenerateDiff(1234, 5, pathHash, 0, tabContent)

	if !hasDiff {
		t.Error("Expected diff for whitespace change (spaces to tabs)")
	}

	if len(diff) == 0 {
		t.Error("Expected non-empty diff for whitespace changes")
	}

	// Test trailing whitespace addition
	fc.UpdateWithWrite(pathHash, 0, originalContent)
	fc.AddRead(pathHash, 0, originalContent)

	trailingContent := []byte("function test() {\n    return true;   \n}")
	trailingDiff, hasTrailingDiff := fc.GenerateDiff(1234, 5, pathHash, 0, trailingContent)

	if !hasTrailingDiff {
		t.Error("Expected diff for trailing whitespace addition")
	}

	// Test line ending whitespace
	fc.UpdateWithWrite(pathHash, 0, originalContent)
	fc.AddRead(pathHash, 0, originalContent)

	lineEndingContent := []byte("function test() {\n    return true;\n\n}")
	lineEndingDiff, hasLineEndingDiff := fc.GenerateDiff(1234, 5, pathHash, 0, lineEndingContent)

	if !hasLineEndingDiff {
		t.Error("Expected diff for line ending addition")
	}

	t.Logf("Tab diff: %s", diff)
	t.Logf("Trailing whitespace diff: %s", trailingDiff)
	t.Logf("Line ending diff: %s", lineEndingDiff)
}

func TestIdenticalContentWrites(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	pathHash := uint32(77777)

	// Original content
	content := []byte("This content will be written identically")
	fc.AddRead(pathHash, 0, content)

	// Write exactly the same content
	identicalContent := make([]byte, len(content))
	copy(identicalContent, content)

	diff, hasDiff := fc.GenerateDiff(1234, 5, pathHash, 0, identicalContent)

	if !hasDiff {
		// Some implementations might not generate diff for identical content
		t.Log("No diff generated for identical content (implementation choice)")
	} else {
		// If diff is generated, it should indicate no changes
		if strings.Contains(diff, "+") || strings.Contains(diff, "-") {
			t.Error("Diff for identical content should not show additions or deletions")
		}
		t.Logf("Identical content diff: %s", diff)
	}

	// Verify cache behavior with identical write
	fc.UpdateWithWrite(pathHash, 0, identicalContent)

	// Content should be invalidated (even if identical)
	_, exists := fc.GetOldContent(pathHash, 0, uint64(len(content)))
	if exists {
		t.Error("Expected identical write to invalidate cache")
	}

	// Read identical content back
	fc.AddRead(pathHash, 0, identicalContent)

	retrieved, exists := fc.GetOldContent(pathHash, 0, uint64(len(identicalContent)))
	if !exists {
		t.Fatal("Expected identical content to be cached after read")
	}

	if !bytes.Equal(retrieved, identicalContent) {
		t.Error("Identical content corrupted")
	}
}

func TestDiffWithDifferentOffsets(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	pathHash := uint32(88888)

	// Add content at different offsets
	header := []byte("File Header\n")
	body := []byte("File body content with multiple lines\nand various data\n")
	footer := []byte("File Footer\n")

	fc.AddRead(pathHash, 0, header)
	fc.AddRead(pathHash, uint64(len(header)), body)
	fc.AddRead(pathHash, uint64(len(header)+len(body)), footer)

	// Generate diff for modification at different offsets

	// 1. Modify header
	newHeader := []byte("Modified Header\n")
	headerDiff, hasHeaderDiff := fc.GenerateDiff(1234, 5, pathHash, 0, newHeader)

	if !hasHeaderDiff {
		t.Error("Expected diff for header modification")
	}

	// 2. Modify body at middle offset
	newBody := []byte("Modified body content with different lines\nand various data\n")
	bodyDiff, hasBodyDiff := fc.GenerateDiff(1234, 5, pathHash, uint64(len(header)), newBody)

	if !hasBodyDiff {
		t.Error("Expected diff for body modification")
	}

	// 3. Modify footer at end
	newFooter := []byte("Modified Footer\n")
	footerDiff, hasFooterDiff := fc.GenerateDiff(1234, 5, pathHash, uint64(len(header)+len(body)), newFooter)

	if !hasFooterDiff {
		t.Error("Expected diff for footer modification")
	}

	// 4. Test diff spanning multiple segments
	entireNewContent := append(append(newHeader, newBody...), newFooter...)
	entireDiff, hasEntireDiff := fc.GenerateDiff(1234, 5, pathHash, 0, entireNewContent)

	if !hasEntireDiff {
		t.Error("Expected diff for entire file modification")
	}

	t.Logf("Header diff: %s", headerDiff)
	t.Logf("Body diff: %s", bodyDiff)
	t.Logf("Footer diff: %s", footerDiff)
	t.Logf("Entire diff length: %d", len(entireDiff))
}

func TestDiffPerformanceCharacteristics(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(256*1024, 512*1024, time.Hour, mockTime)

	pathHash := uint32(99999)

	// Test diff generation performance with various content sizes
	sizes := []int{1024, 4 * 1024, 16 * 1024, 64 * 1024}

	for _, size := range sizes {
		content := make([]byte, size)
		for i := range content {
			content[i] = byte('A' + (i % 26))
		}

		fc.UpdateWithWrite(pathHash, 0, []byte("reset"))
		fc.AddRead(pathHash, 0, content)

		// Modify small portion
		modifiedContent := make([]byte, size)
		copy(modifiedContent, content)
		// Change 1% of content
		changeSize := size / 100
		if changeSize == 0 {
			changeSize = 1
		}
		for i := 0; i < changeSize; i++ {
			modifiedContent[size/2+i] = byte('a' + (i % 26))
		}

		start := time.Now()
		diff, hasDiff := fc.GenerateDiff(1234, 5, pathHash, 0, modifiedContent)
		duration := time.Since(start)

		if !hasDiff {
			t.Errorf("Expected diff for %d byte content", size)
			continue
		}

		t.Logf("Size: %d bytes, Diff time: %v, Diff size: %d bytes",
			size, duration, len(diff))

		// Diff generation should be reasonable fast (< 100ms for 64KB)
		if size <= 64*1024 && duration > 100*time.Millisecond {
			t.Errorf("Diff generation too slow for %d bytes: %v", size, duration)
		}
	}
}
