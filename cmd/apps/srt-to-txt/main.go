package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type SubtitleSplitter struct {
	StartTime     string
	EndTime       string
	Verbose       bool
	OutputFile    string
	PauseDuration int
}

func main() {
	splitter := &SubtitleSplitter{}

	rootCmd := &cobra.Command{
		Use:   "srt-to-txt [files...]",
		Short: "Convert SRT files to plain text",
		Long:  "A tool to convert SRT subtitle files to plain text, optionally filtering by time range.",
		Args:  cobra.MinimumNArgs(1),
		Run:   splitter.run,
	}

	rootCmd.Flags().StringVarP(&splitter.StartTime, "start", "s", "", "Start time (format: HH:MM:SS.mmm)")
	rootCmd.Flags().StringVarP(&splitter.EndTime, "end", "e", "", "End time (format: HH:MM:SS.mmm)")
	rootCmd.Flags().BoolVarP(&splitter.Verbose, "verbose", "v", false, "Show verbose output")
	rootCmd.Flags().StringVarP(&splitter.OutputFile, "output", "o", "", "Output file (default: stdout)")
	rootCmd.Flags().IntVarP(&splitter.PauseDuration, "pause", "p", 2, "Pause duration in seconds to start a new paragraph")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (s *SubtitleSplitter) run(cmd *cobra.Command, args []string) {
	start, err := parseTime(s.StartTime)
	if err != nil {
		fmt.Printf("Error parsing start time: %v\n", err)
		os.Exit(1)
	}

	end, err := parseTime(s.EndTime)
	if err != nil {
		fmt.Printf("Error parsing end time: %v\n", err)
		os.Exit(1)
	}

	for _, file := range args {
		if err := s.processFile(file, start, end); err != nil {
			fmt.Printf("Error processing file %s: %v\n", file, err)
		}
	}
}

func (s *SubtitleSplitter) processFile(filename string, start, end time.Duration) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var text strings.Builder
	var lastEndTime time.Duration

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "-->") {
			timeRange := strings.Split(line, "-->")
			if len(timeRange) != 2 {
				if s.Verbose {
					fmt.Printf("Skipping invalid time range: %s\n", line)
				}
				continue
			}

			startTime, err := parseTime(strings.TrimSpace(timeRange[0]))
			if err != nil {
				if s.Verbose {
					fmt.Printf("Error parsing time: %s\n", err)
				}
				continue
			}

			endTime, err := parseTime(strings.TrimSpace(timeRange[1]))
			if err != nil {
				if s.Verbose {
					fmt.Printf("Error parsing time: %s\n", err)
				}
				continue
			}

			// Check if we should include this subtitle based on the time range
			if (start == 0 || startTime >= start) && (end == 0 || startTime <= end) {
				if s.Verbose {
					fmt.Printf("Accepting subtitle at %s\n", timeRange[0])
					fmt.Printf("Last end time: %s\n", lastEndTime)
					fmt.Printf("Current start time: %s\n", startTime)
				}

				// Check for pause duration to start a new paragraph
				if lastEndTime != 0 && startTime-lastEndTime > time.Duration(s.PauseDuration)*time.Second {
					text.WriteString("\n\n")
					if s.Verbose {
						fmt.Printf("Inserting paragraph break due to pause longer than %d seconds\n", s.PauseDuration)
					}
				}
				lastEndTime = endTime

				for scanner.Scan() {
					subtitle := scanner.Text()
					if subtitle == "" {
						break
					}
					text.WriteString(subtitle + " ")
					if s.Verbose {
						fmt.Printf("  Subtitle: %s\n", subtitle)
					}
				}
			} else if s.Verbose {
				fmt.Printf("Skipping subtitle at %s (out of range)\n", timeRange[0])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	wrappedText := wrapText(text.String(), 80)

	if s.OutputFile == "" {
		fmt.Println(wrappedText)
	} else {
		if s.Verbose {
			fmt.Printf("Writing output to: %s\n", s.OutputFile)
		}
		return os.WriteFile(s.OutputFile, []byte(wrappedText), 0644)
	}

	return nil
}

func parseTime(timeStr string) (time.Duration, error) {
	if timeStr == "" {
		return 0, nil
	}

	// Split the time string into seconds and milliseconds
	parts := strings.Split(timeStr, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time format")
	}

	// Parse seconds
	seconds, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}

	// Parse milliseconds (if present)
	var milliseconds float64
	if len(parts[1]) > 0 {
		milliseconds, err = strconv.ParseFloat("0."+parts[1], 64)
		if err != nil {
			return 0, err
		}
	}

	// Convert to duration
	duration := time.Duration(seconds*float64(time.Second)) +
		time.Duration(milliseconds*float64(time.Second))

	return duration, nil
}

func wrapText(text string, lineWidth int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var wrappedText strings.Builder
	currentLineLength := 0

	for _, word := range words {
		if currentLineLength+len(word)+1 > lineWidth {
			wrappedText.WriteString("\n")
			currentLineLength = 0
		}
		if currentLineLength > 0 {
			wrappedText.WriteString(" ")
			currentLineLength++
		}
		wrappedText.WriteString(word)
		currentLineLength += len(word)
	}

	return wrappedText.String()
}
