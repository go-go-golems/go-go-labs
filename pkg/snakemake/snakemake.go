package snakemake

import (
	"fmt"
	"path/filepath"
	"regexp"
)

// ParseLog parses the Snakemake log file and returns structured LogData.
func ParseLog(filename string, debug bool) (LogData, error) {
	tokenizer, err := NewTokenizer(filename, debug)
	if err != nil {
		return LogData{}, fmt.Errorf("failed to create tokenizer: %w", err)
	}
	defer tokenizer.Close()

	parser := NewParser(tokenizer, debug)
	logData, err := parser.ParseLog()
	if err != nil {
		return LogData{}, err
	}

	// Extract jobId from filename
	jobId := extractJobIdFromFilename(filename)
	if jobId != "" {
		for i := range logData.Jobs {
			logData.Jobs[i].ID = jobId
		}
	}

	return logData, nil
}

func extractJobIdFromFilename(filename string) string {
	base := filepath.Base(filename)
	re := regexp.MustCompile(`slurm-(\d+)\.out`)
	matches := re.FindStringSubmatch(base)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
