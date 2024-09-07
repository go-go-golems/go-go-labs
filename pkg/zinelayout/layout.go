package zinelayout

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
)

// ZineLayout represents the entire YAML structure
type ZineLayout struct {
	Global      Global       `yaml:"global"`
	PageSetup   PageSetup    `yaml:"page_setup"`
	OutputPages []OutputPage `yaml:"output_pages"`
}

// Global represents the global settings
type Global struct {
	Margin Margin `yaml:"margin"`
}

// PageSetup represents the page setup settings
type PageSetup struct {
	GridSize struct {
		Rows    int `yaml:"rows"`
		Columns int `yaml:"columns"`
	} `yaml:"grid_size"`
	Margins Margin `yaml:"margins"`
}

// OutputPage represents a single output page
type OutputPage struct {
	ID     string   `yaml:"id"`
	Margin Margin   `yaml:"margin"`
	Layout []Layout `yaml:"layout"`
}

// Layout represents the layout of an input page on an output page
type Layout struct {
	InputIndex int      `yaml:"input_index"`
	Position   Position `yaml:"position"`
	Rotation   int      `yaml:"rotation"`
	Margin     Margin   `yaml:"margin"`
}

// Position represents the position of an input page on the output page
type Position struct {
	Row    float64 `yaml:"row"`
	Column float64 `yaml:"column"`
}

// Margin represents margin settings
type Margin struct {
	Top    int `yaml:"top"`
	Bottom int `yaml:"bottom"`
	Left   int `yaml:"left"`
	Right  int `yaml:"right"`
}

func (zl *ZineLayout) CreateOutputImage(outputPage OutputPage, inputImages []image.Image) (image.Image, error) {
	fmt.Println("Creating output image")
	for _, inputImage := range inputImages {
		fmt.Printf("Input image size: %v\n", inputImage.Bounds().Size())
	}
	inputSize := inputImages[0].Bounds().Size()
	width := inputSize.X * zl.PageSetup.GridSize.Columns
	height := inputSize.Y * zl.PageSetup.GridSize.Rows

	// Create the output image without margins
	outputImage := image.NewRGBA(image.Rect(0, 0, width, height))

	for _, layout := range outputPage.Layout {
		if layout.Rotation != 0 && layout.Rotation != 180 {
			return nil, errors.New(fmt.Sprintf("invalid rotation %d for input index %d", layout.Rotation, layout.InputIndex))
		}

		inputImage := inputImages[layout.InputIndex-1]
		x := int(float64(width) * layout.Position.Column / float64(zl.PageSetup.GridSize.Columns))
		y := int(float64(height) * layout.Position.Row / float64(zl.PageSetup.GridSize.Rows))

		// Handle rotation
		rotatedImage := rotateImage(inputImage, layout.Rotation)
		rotatedSize := rotatedImage.Bounds().Size()

		// Draw the rotated input image onto the output image
		draw.Draw(outputImage, image.Rect(x, y, x+rotatedSize.X, y+rotatedSize.Y), rotatedImage, image.Point{0, 0}, draw.Over)
	}

	// Add margins to the final image
	finalWidth := width + zl.PageSetup.Margins.Left + zl.PageSetup.Margins.Right + outputPage.Margin.Left + outputPage.Margin.Right
	finalHeight := height + zl.PageSetup.Margins.Top + zl.PageSetup.Margins.Bottom + outputPage.Margin.Top + outputPage.Margin.Bottom
	finalImage := image.NewRGBA(image.Rect(0, 0, finalWidth, finalHeight))

	// Draw the output image onto the final image with margins
	draw.Draw(finalImage, image.Rect(
		zl.PageSetup.Margins.Left+outputPage.Margin.Left,
		zl.PageSetup.Margins.Top+outputPage.Margin.Top,
		finalWidth-zl.PageSetup.Margins.Right-outputPage.Margin.Right,
		finalHeight-zl.PageSetup.Margins.Bottom-outputPage.Margin.Bottom,
	), outputImage, image.Point{0, 0}, draw.Over)

	return finalImage, nil
}

// New function to handle image rotation
func rotateImage(img image.Image, degrees int) image.Image {
	switch degrees {
	case 0:
		return img
	case 90:
		return rotate90(img)
	case 180:
		return rotate180(img)
	case 270:
		return rotate270(img)
	default:
		return img
	}
}

func rotate90(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			newImg.Set(bounds.Max.Y-y-1, x, img.At(x, y))
		}
	}
	return newImg
}

func rotate180(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			newImg.Set(bounds.Max.X-x-1, bounds.Max.Y-y-1, img.At(x, y))
		}
	}
	return newImg
}

func rotate270(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			newImg.Set(y, bounds.Max.X-x-1, img.At(x, y))
		}
	}
	return newImg
}

func AllImagesSameSize(images []image.Image) bool {
	if len(images) == 0 {
		return true
	}
	firstSize := images[0].Bounds().Size()
	for _, img := range images[1:] {
		if img.Bounds().Size() != firstSize {
			return false
		}
	}
	return true
}
