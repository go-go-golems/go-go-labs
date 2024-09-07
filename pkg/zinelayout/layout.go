package zinelayout

import (
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
	Orientation string `yaml:"orientation"`
	GridSize    struct {
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

func CreateOutputImage(pageSetup PageSetup, outputPage OutputPage, inputImages []image.Image) image.Image {
	inputSize := inputImages[0].Bounds().Size()
	width := inputSize.X * pageSetup.GridSize.Columns
	height := inputSize.Y * pageSetup.GridSize.Rows

	if pageSetup.Orientation == "landscape" {
		width, height = height, width
	}

	// Apply global margins
	width += pageSetup.Margins.Left + pageSetup.Margins.Right
	height += pageSetup.Margins.Top + pageSetup.Margins.Bottom

	// Apply output page margins
	width += outputPage.Margin.Left + outputPage.Margin.Right
	height += outputPage.Margin.Top + outputPage.Margin.Bottom

	outputImage := image.NewRGBA(image.Rect(0, 0, width, height))

	for _, layout := range outputPage.Layout {
		inputImage := inputImages[layout.InputIndex-1]
		x := int(float64(width-pageSetup.Margins.Left-pageSetup.Margins.Right-outputPage.Margin.Left-outputPage.Margin.Right) * layout.Position.Column / float64(pageSetup.GridSize.Columns))
		y := int(float64(height-pageSetup.Margins.Top-pageSetup.Margins.Bottom-outputPage.Margin.Top-outputPage.Margin.Bottom) * layout.Position.Row / float64(pageSetup.GridSize.Rows))

		// Add margins to the position
		x += pageSetup.Margins.Left + outputPage.Margin.Left
		y += pageSetup.Margins.Top + outputPage.Margin.Top

		// Draw the input image onto the output image
		draw.Draw(outputImage, image.Rect(x, y, x+inputSize.X, y+inputSize.Y), inputImage, image.Point{0, 0}, draw.Over)
	}

	return outputImage
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
