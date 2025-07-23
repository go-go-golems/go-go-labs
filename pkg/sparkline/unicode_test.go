package sparkline

import (
	"fmt"
	"testing"
	"unicode/utf8"
)

func TestUnicodeLength(t *testing.T) {
	// Test the actual bar characters
	bars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	
	fmt.Printf("=== Unicode Character Analysis ===\n")
	for i, bar := range bars {
		fmt.Printf("bars[%d] = %q, len=%d, utf8.RuneCountInString=%d\n", 
			i, bar, len(bar), utf8.RuneCountInString(bar))
	}
	
	// Test the specific output from our sparkline
	output := "                        ▁▁▂▃▅█"
	fmt.Printf("\nFull output analysis:\n")
	fmt.Printf("String: %q\n", output)
	fmt.Printf("len(output): %d\n", len(output))
	fmt.Printf("utf8.RuneCountInString(output): %d\n", utf8.RuneCountInString(output))
	
	// Count each part
	spaces := "                        " // 24 spaces
	bars_part := "▁▁▂▃▅█" // 6 bar characters
	
	fmt.Printf("\nPart analysis:\n")
	fmt.Printf("Spaces: %q, len=%d, runes=%d\n", spaces, len(spaces), utf8.RuneCountInString(spaces))
	fmt.Printf("Bars: %q, len=%d, runes=%d\n", bars_part, len(bars_part), utf8.RuneCountInString(bars_part))
	
	fmt.Printf("Total: %d + %d = %d bytes\n", len(spaces), len(bars_part), len(spaces)+len(bars_part))
	fmt.Printf("Total runes: %d + %d = %d\n", utf8.RuneCountInString(spaces), utf8.RuneCountInString(bars_part), utf8.RuneCountInString(spaces)+utf8.RuneCountInString(bars_part))
}
