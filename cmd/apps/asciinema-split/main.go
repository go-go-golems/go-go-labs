package main

// Created with chatgpt
// https://chat.openai.com/share/8d6c90cf-a058-4016-80b7-c5dca9f85c71

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type Header struct {
	Version   int               `json:"version"`
	Width     int               `json:"width"`
	Height    int               `json:"height"`
	Timestamp int               `json:"timestamp"`
	Env       map[string]string `json:"env"`
}

type Event []interface{}

var splitBytes int
var splitTime float64

func main() {
	var cmd = &cobra.Command{
		Use:   "castsplitter [inputfile]",
		Short: "Splits asciinema recordings",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}

	cmd.Flags().IntVar(&splitBytes, "split-bytes", 1048576, "Split by bytes")
	cmd.Flags().Float64Var(&splitTime, "split-time", -1, "Split by time in seconds")

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	inputfile := args[0]
	file, err := os.Open(inputfile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	headerLine := scanner.Text()

	var header Header
	err = json.Unmarshal([]byte(headerLine), &header)
	if err != nil {
		fmt.Println("Error unmarshalling header:", err)
		return
	}

	var events []Event
	for scanner.Scan() {
		var event Event
		line := strings.TrimSpace(scanner.Text())
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		events = append(events, event)
	}

	if splitTime > 0 {
		splitByTime(header, events, splitTime, inputfile)
	} else {
		splitByBytes(header, events, splitBytes, inputfile)
	}
}

func splitByTime(header Header, events []Event, splitTime float64, inputfile string) {
	startTime := events[0][0].(float64)
	newEvents := []Event{}
	partNumber := 1

	for _, event := range events {
		if event[0].(float64)-startTime > splitTime {
			writeToFile(header, newEvents, partNumber, inputfile)
			newEvents = []Event{}
			startTime = event[0].(float64)
			partNumber++
		}

		// reset event time to 0 for the new file
		event[0] = event[0].(float64) - startTime
		newEvents = append(newEvents, event)
	}

	if len(newEvents) > 0 {
		writeToFile(header, newEvents, partNumber, inputfile)
	}
}

func splitByBytes(header Header, events []Event, splitBytes int, inputfile string) {
	newEvents := []Event{}
	partNumber := 1
	currentSize := 0

	for _, event := range events {
		eventBytes, _ := json.Marshal(event)
		currentSize += len(eventBytes)

		if currentSize > splitBytes {
			writeToFile(header, newEvents, partNumber, inputfile)
			newEvents = []Event{}
			currentSize = 0
			partNumber++
		}

		newEvents = append(newEvents, event)
	}

	if len(newEvents) > 0 {
		writeToFile(header, newEvents, partNumber, inputfile)
	}
}

func writeToFile(header Header, events []Event, partNumber int, inputfile string) {
	filename := fmt.Sprintf("%s-part%s.cast", inputfile, strconv.Itoa(partNumber))
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	headerBytes, err := json.Marshal(header)
	if err != nil {
		fmt.Println("Error marshalling header:", err)
		return
	}
	file.WriteString(string(headerBytes) + "\n")

	// Calculate the timestamp offset to start at 0 for each new file.
	// If there are no events, we set it to 0 by default.
	var offset float64
	if len(events) > 0 {
		offset = events[0][0].(float64)
	} else {
		offset = 0
	}

	for _, event := range events {
		event[0] = event[0].(float64) - offset // Adjust each timestamp by the offset
		eventBytes, err := json.Marshal(event)
		if err != nil {
			continue
		}
		eventStr := string(eventBytes)
		// Replace instances of Unicode escape sequences with the desired characters
		eventStr = strings.ReplaceAll(eventStr, "\\u0008", "\\b")
		eventStr = strings.ReplaceAll(eventStr, "\\u003e", ">")
		eventStr = strings.ReplaceAll(eventStr, "\\u003c", "<")
		file.WriteString(eventStr + "\n")
	}
}
