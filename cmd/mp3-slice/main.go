package main

import (
	"flag"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/mp3-slice/mp3lib"
	"os"
	"path/filepath"
)

func ensureDirExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

func main() {
	// Define command line flags
	mp3FilePath := flag.String("file", "", "Path to the mp3 file to slice")
	duration := flag.Int("duration", 0, "Duration of each slice in seconds")
	outputDir := flag.String("output", ".", "Output directory for sliced mp3 segments")

	// Parse the flags
	flag.Parse()

	// Basic validation
	if *mp3FilePath == "" {
		fmt.Println("Please provide a valid mp3 file path using the -file flag.")
		return
	}
	if *duration <= 0 {
		fmt.Println("Please provide a valid slice duration in seconds using the -duration flag.")
		return
	}

	// Ensure the output directory exists
	if err := ensureDirExists(*outputDir); err != nil {
		fmt.Printf("Error ensuring output directory exists: %v\n", err)
		return
	}

	// Get the length of the MP3 file
	length, err := mp3lib.GetLengthSeconds(*mp3FilePath)
	if err != nil {
		fmt.Printf("Error getting mp3 file length: %v\n", err)
		return
	}

	// Calculate the number of slices
	numSlices := length / *duration
	if length%*duration != 0 {
		numSlices++
	}

	// Start slicing the mp3 file
	for i := 0; i < numSlices; i++ {
		startSec := i * *duration
		endSec := startSec + *duration
		if endSec > length {
			endSec = length
		}

		outputFilePath := filepath.Join(*outputDir, fmt.Sprintf("slice_%d.mp3", i+1))
		err := mp3lib.ExtractSectionToFile(*mp3FilePath, outputFilePath, startSec, endSec)
		if err != nil {
			fmt.Printf("Error extracting segment from %d to %d seconds: %v\n", startSec, endSec, err)
			return
		}

		fmt.Printf("Segment %d saved to %s\n", i+1, outputFilePath)
	}

	fmt.Println("MP3 slicing complete.")
}
