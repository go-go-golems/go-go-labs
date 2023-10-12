package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

type Transcription struct {
	File     string                `json:"file"`
	Response *openai.AudioResponse `json:"response"`
	err      error
}

type TranscriptionClient struct {
	client      *openai.Client
	model       string
	prompt      string
	language    string
	temperature float32
}

func NewTranscriptionClient(apiKey, model, prompt, language string, temperature float32) *TranscriptionClient {
	return &TranscriptionClient{
		client:      openai.NewClient(apiKey),
		model:       model,
		prompt:      prompt,
		language:    language,
		temperature: temperature,
	}
}

func (tc *TranscriptionClient) transcribeFile(mp3FilePath string, out chan<- Transcription, wg *sync.WaitGroup) {
	defer wg.Done()

	// Set up the audio request
	req := openai.AudioRequest{
		Model:       tc.model,
		FilePath:    mp3FilePath,
		Prompt:      tc.prompt,
		Temperature: tc.temperature,
		Language:    tc.language,
		Format:      openai.AudioResponseFormatJSON,
	}

	log.Info().Str("file", mp3FilePath).Msg("Transcribing...")
	// Call the CreateTranscription method
	resp, err := tc.client.CreateTranscription(context.Background(), req)
	if err != nil {
		log.Printf("Failed to transcribe %s: %v\n", mp3FilePath, err)
		out <- Transcription{File: mp3FilePath, err: err}
		return
	}

	out <- Transcription{File: mp3FilePath, Response: &resp}
}

func main() {
	// CLI arguments
	dirPath := flag.String("d", "", "Path to the directory containing MP3 files")
	workers := flag.Int("w", 4, "Number of parallel workers")
	model := flag.String("model", openai.Whisper1, "Model used for transcription")
	prompt := flag.String("prompt", "", "Prompt for the transcription model")
	language := flag.String("language", "", "Language for the transcription model")
	temperature := flag.Float64("temperature", 0, "Temperature for the transcription model")
	flag.Parse()

	// Create the TranscriptionClient
	tc := NewTranscriptionClient(os.Getenv("OPENAI_API_KEY"), *model, *prompt, *language, float32(*temperature))

	if *dirPath == "" {
		fmt.Println("Please specify a directory path containing MP3 files using -d flag.")
		os.Exit(1)
	}

	// Read the directory
	files, err := os.ReadDir(*dirPath)
	if err != nil {
		log.Fatal().Err(err).Str("dir", *dirPath).Msg("Failed to read the directory")
	}

	var wg sync.WaitGroup
	out := make(chan Transcription, len(files))

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".mp3") {
			wg.Add(1)
			go tc.transcribeFile(filepath.Join(*dirPath, file.Name()), out, &wg)

			// Limit concurrent workers
			for len(out) >= *workers {
				<-out
			}
		}
	}

	wg.Wait()
	close(out)

	// Collect and reassemble transcriptions
	transcriptions := map[string]Transcription{}
	for transcription := range out {
		transcriptions[transcription.File] = transcription
	}

	for _, file := range files {
		transcription, ok := transcriptions[filepath.Join(*dirPath, file.Name())]
		if !ok {
			log.Warn().Str("file", file.Name()).Msg("No transcription found")
			continue
		}
		fmt.Println("Transcription for", file.Name(), ":", transcription.Response.Text)
	}

}
