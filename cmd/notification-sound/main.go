package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/pkg/errors"
	
	"github.com/go-go-golems/go-go-labs/pkg/audio"
)

var (
	outputFile string
	frequency  float64
	duration   time.Duration
	volume     float64
	download   bool
	soundType  string
)

var rootCmd = &cobra.Command{
	Use:   "notification-sound",
	Short: "Generate or download notification sounds",
	Long: `A tool to generate notification sounds programmatically or download them from the internet.
	
Examples:
  # Generate a default notification sound
  notification-sound generate -o notification.wav
  
  # Generate a custom sound with specific frequency and duration
  notification-sound generate -o custom.wav -f 1000 -d 300ms -v 0.5
  
  # Download a free notification sound
  notification-sound download -t gentle_bell -o bell.wav`,
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a notification sound",
	Long:  "Generate a notification sound using sine wave synthesis",
	RunE: func(cmd *cobra.Command, args []string) error {
		if outputFile == "" {
			return errors.New("output file is required (-o flag)")
		}
		
		ns := audio.NewNotificationSound()
		
		// Apply custom parameters if provided
		if frequency > 0 {
			ns.Frequency = frequency
		}
		if duration > 0 {
			ns.Duration = duration
		}
		if volume > 0 {
			ns.Volume = volume
		}
		
		fmt.Printf("Generating notification sound:\n")
		fmt.Printf("  Frequency: %.1f Hz\n", ns.Frequency)
		fmt.Printf("  Duration: %v\n", ns.Duration)
		fmt.Printf("  Volume: %.1f\n", ns.Volume)
		fmt.Printf("  Output: %s\n", outputFile)
		
		err := ns.SaveToFile(outputFile)
		if err != nil {
			return errors.Wrap(err, "failed to save notification sound")
		}
		
		fmt.Printf("✅ Notification sound saved to %s\n", outputFile)
		return nil
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a free notification sound",
	Long:  "Download a free notification sound from the internet",
	RunE: func(cmd *cobra.Command, args []string) error {
		if outputFile == "" {
			return errors.New("output file is required (-o flag)")
		}
		
		sounds := audio.GetFreeNotificationSounds()
		
		if soundType == "" {
			fmt.Println("Available sound types:")
			for name := range sounds {
				fmt.Printf("  - %s\n", name)
			}
			return errors.New("sound type is required (-t flag)")
		}
		
		url, exists := sounds[soundType]
		if !exists {
			fmt.Println("Available sound types:")
			for name := range sounds {
				fmt.Printf("  - %s\n", name)
			}
			return fmt.Errorf("unknown sound type: %s", soundType)
		}
		
		fmt.Printf("Downloading %s from %s...\n", soundType, url)
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		err := audio.DownloadNotificationSound(ctx, url, outputFile)
		if err != nil {
			return errors.Wrap(err, "failed to download notification sound")
		}
		
		fmt.Printf("✅ Notification sound downloaded to %s\n", outputFile)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available notification sounds for download",
	Long:  "List all available notification sounds that can be downloaded",
	Run: func(cmd *cobra.Command, args []string) {
		sounds := audio.GetFreeNotificationSounds()
		
		fmt.Println("Available notification sounds for download:")
		for name, url := range sounds {
			fmt.Printf("  %-15s %s\n", name, url)
		}
	},
}

func init() {
	// Generate command flags
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (required)")
	generateCmd.Flags().Float64VarP(&frequency, "frequency", "f", 0, "Sound frequency in Hz (default: 800)")
	generateCmd.Flags().DurationVarP(&duration, "duration", "d", 0, "Sound duration (default: 500ms)")
	generateCmd.Flags().Float64VarP(&volume, "volume", "v", 0, "Sound volume 0.0-1.0 (default: 0.3)")
	
	// Download command flags
	downloadCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (required)")
	downloadCmd.Flags().StringVarP(&soundType, "type", "t", "", "Sound type to download (required)")
	
	// Add subcommands
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(listCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
} 