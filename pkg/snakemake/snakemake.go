package snakemake

// This file serves as the main entry point for the snakemake package.
// It imports and uses the other components defined in separate files.

import (
	"fmt"
)

// ParseLog parses the Snakemake log file and returns structured LogData.
func ParseLog(filename string, debug bool) (LogData, error) {
	tokenizer, err := NewTokenizer(filename, debug)
	if err != nil {
		return LogData{}, fmt.Errorf("failed to create tokenizer: %w", err)
	}
	defer tokenizer.Close()

	parser := NewParser(tokenizer, debug)
	return parser.ParseLog()
}
