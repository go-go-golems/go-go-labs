package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"

	"golang.design/x/clipboard"
)

func main() {
	// Initialize the clipboard package
	err := clipboard.Init()
	if err != nil {
		log.Fatalf("Failed to initialize clipboard: %v", err)
	}

	// Read image data from clipboard
	imageData := clipboard.Read(clipboard.FmtImage)
	if imageData == nil {
		log.Fatalf("No image data found in clipboard")
	}

	// Print hexdump of clipboard content
	fmt.Println("Clipboard content (hexdump):")
	for i, b := range imageData {
		if i%16 == 0 {
			fmt.Printf("\n%04x: ", i)
		}
		fmt.Printf("%02x ", b)
	}
	fmt.Println()

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		log.Fatalf("Failed to decode image: %v", err)
	}

	// Get image dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Print dimensions
	fmt.Printf("Image dimensions: %dx%d\n", width, height)
}
