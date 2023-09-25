package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func isValidJSON(filepath string) bool {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return false
	}

	var js interface{}
	return json.Unmarshal(data, &js) == nil
}

func processLine(line, outputDir string, wg *sync.WaitGroup) {
	defer wg.Done()

	fields := strings.SplitN(line, ",", 2)
	if len(fields) < 2 {
		fmt.Printf("Skipping malformed line: %s\n", line)
		return
	}

	name := fields[0]
	botanicalName := fields[1]
	outputFilePath := fmt.Sprintf("%s/%s.json", outputDir, name)

	if _, err := os.Stat(outputFilePath); err == nil {
		// If file exists
		if isValidJSON(outputFilePath) {
			fmt.Printf("Valid JSON already exists for %s. Skipping...\n", name)
			return
		} else {
			fmt.Printf("Invalid JSON already exists for %s. Removing...\n", name)
			// Remove the invalid JSON file
			os.Remove(outputFilePath)
		}
	}

	fmt.Printf("Processing %s...\n", name)
	cmd := exec.Command("pinocchio", "ttc", "plants", "--name", name, "--botanical-name", botanicalName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing command for %s: %v\n", name, err)
		return
	}

	outFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Printf("Error creating file for %s: %v\n", name, err)
		return
	}
	defer func(outFile *os.File) {
		_ = outFile.Close()
	}(outFile)

	_, err = outFile.Write(output)
	if err != nil {
		fmt.Printf("Error writing to file for %s: %v\n", name, err)
	}
}

func main() {
	inputFile := flag.String("file", "", "Path to the input file")
	outputDir := flag.String("outdir", "output", "Path to the output directory")
	NTHREADs := flag.Int("threads", 4, "Number of threads to run in parallel")

	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Please specify an input file using the -file flag.")
		return
	}

	// Check if the input file exists
	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		fmt.Printf("Error: File %s not found!\n", *inputFile)
		return
	}

	// Create output directory if it doesn't exist
	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		_ = os.Mkdir(*outputDir, 0755)
	}

	file, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var wg sync.WaitGroup

	sem := make(chan struct{}, *NTHREADs)
	for scanner.Scan() {
		wg.Add(1)
		sem <- struct{}{} // Acquire one semaphore slot
		go func(line string) {
			defer func() { <-sem }() // Release one semaphore slot
			processLine(line, *outputDir, &wg)
		}(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

	wg.Wait()
}
