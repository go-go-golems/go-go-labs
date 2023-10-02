package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

const authToken = "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"

func transcribeFile(client *openai.Client, mp3FilePath string, out chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Set up the audio request
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: mp3FilePath,
		Format:   openai.AudioResponseFormatJSON,
	}

	// Call the CreateTranscription method
	resp, err := client.CreateTranscription(context.Background(), req)
	if err != nil {
		log.Printf("Failed to transcribe %s: %v\n", mp3FilePath, err)
		out <- ""
		return
	}

	out <- resp.Text
}

func main() {
	// CLI arguments
	dirPath := flag.String("d", "", "Path to the directory containing MP3 files")
	workers := flag.Int("w", 4, "Number of parallel workers")
	flag.Parse()

	if *dirPath == "" {
		fmt.Println("Please specify a directory path containing MP3 files using -d flag.")
		os.Exit(1)
	}

	// Read the directory
	files, err := ioutil.ReadDir(*dirPath)
	if err != nil {
		log.Fatalf("Failed to read the directory: %v", err)
	}

	client := openai.NewClient(authToken)

	var wg sync.WaitGroup
	out := make(chan string, len(files))

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".mp3") {
			wg.Add(1)
			go transcribeFile(client, filepath.Join(*dirPath, file.Name()), out, &wg)

			// Limit concurrent workers
			for len(out) >= *workers {
				<-out
			}
		}
	}

	wg.Wait()
	close(out)

	// Collect and reassemble transcriptions
	var transcriptions []string
	for transcription := range out {
		transcriptions = append(transcriptions, transcription)
	}

	fmt.Println("Combined Transcription:", strings.Join(transcriptions, "\n"))
}
