package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"

	"github.com/atotto/clipboard"
)

func main() {
	// Read data from clipboard
	data, err := clipboard.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read clipboard: %v", err)
	}

	// Print hexdump of clipboard content
	fmt.Println("Clipboard content (hexdump):")
	for i, b := range []byte(data) {
		if i%16 == 0 {
			fmt.Printf("\n%04x: ", i)
		}
		fmt.Printf("%02x ", b)
	}
	fmt.Println()

	// Convert string data to byte slice
	imageData := []byte(data)

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
