package zinelayout

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/rs/zerolog/log"
)

type ZineLayout struct {
	PageSetup   *PageSetup    `yaml:"page_setup"`
	OutputPages []*OutputPage `yaml:"output_pages"`
	Global      *Global       `yaml:"global"`
}

type Global struct {
	Border *Border `yaml:"border"`
	PPI    float64 `yaml:"ppi"`
}

type PageSetup struct {
	GridSize struct {
		Rows    int `yaml:"rows"`
		Columns int `yaml:"columns"`
	} `yaml:"grid_size"`
	Margin     *Margin `yaml:"margin"`
	PageBorder *Border `yaml:"border"`
}

type OutputPage struct {
	ID           string    `yaml:"id"`
	Margin       *Margin   `yaml:"margin"`
	Layout       []*Layout `yaml:"layout"`
	LayoutBorder *Border   `yaml:"border"`
}

type Layout struct {
	InputIndex        int      `yaml:"input_index"`
	Position          Position `yaml:"position"`
	Rotation          int      `yaml:"rotation"`
	Margin            *Margin  `yaml:"margin"`
	InnerLayoutBorder *Border  `yaml:"border"`
}

type Border struct {
	Enabled bool        `yaml:"enabled"`
	Color   CustomColor `yaml:"color"`
	Type    BorderType  `yaml:"type"`
}

// Position represents the position of an input page on the output page
type Position struct {
	Row    int `yaml:"row"`
	Column int `yaml:"column"`
}

func (zl *ZineLayout) CreateOutputImage(outputPage *OutputPage, inputImages []image.Image) (image.Image, error) {
	if zl.Global.PPI == 0 {
		return nil, fmt.Errorf("ppi is not set")
	}
	err := zl.ComputeAllMargins()
	if err != nil {
		return nil, fmt.Errorf("error computing all margins: %w", err)
	}

	fmt.Println("Creating output image")
	for _, inputImage := range inputImages {
		fmt.Printf("Input image size: %v\n", inputImage.Bounds().Size())
	}
	inputSize := inputImages[0].Bounds().Size()

	type CellSize struct {
		Margin *Margin
		Width  int
		Height int
		X      int
		Y      int
	}

	// Create a 2D array to store CellSize for each cell
	cells := make([][]CellSize, zl.PageSetup.GridSize.Rows)
	for row := range cells {
		cells[row] = make([]CellSize, zl.PageSetup.GridSize.Columns)
		for column := range cells[row] {
			cells[row][column] = CellSize{Margin: &Margin{}}
		}
	}

	// Calculate cell sizes and update cells
	for _, layout := range outputPage.Layout {
		row, col := int(layout.Position.Row), int(layout.Position.Column)
		cells[row][col].Margin = layout.Margin
		cells[row][col].Width = inputSize.X + layout.Margin.Left.Pixels + layout.Margin.Right.Pixels
		cells[row][col].Height = inputSize.Y + layout.Margin.Top.Pixels + layout.Margin.Bottom.Pixels
	}

	totalHeight := 0
	totalWidth := 0
	// Calculate output image size and cell positions
	width, height := 0, 0
	for row := range cells {
		maxCellHeight := 0
		for column := range cells[row] {
			cells[row][column].X = width
			cells[row][column].Y = height
			width += cells[row][column].Width
			maxCellHeight = max(maxCellHeight, cells[row][column].Height)
		}
		height += maxCellHeight
		totalWidth = max(totalWidth, width)
		totalHeight += maxCellHeight
		width = 0 // Reset width for the next row
	}

	// Final output image size
	width = totalWidth
	height = totalHeight

	fmt.Printf("Total width: %d, Total height: %d\n", width, height)

	// Create the output image without global margins
	outputImage := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill the output image with white color
	draw.Draw(outputImage, outputImage.Bounds(), image.White, image.Point{}, draw.Src)

	// Use the specified border color or default to black if not set
	globalBorderColor := color.RGBA{0, 0, 0, 255}
	if zl.Global.Border != nil && zl.Global.Border.Color.RGBA != (color.RGBA{}) {
		globalBorderColor = zl.Global.Border.Color.RGBA
	}

	for _, layout := range outputPage.Layout {
		if layout.Rotation != 0 && layout.Rotation != 180 {
			return nil, fmt.Errorf("invalid rotation %d for input index %d", layout.Rotation, layout.InputIndex)
		}

		inputImage := inputImages[layout.InputIndex-1]
		destPoint := image.Point{
			X: cells[layout.Position.Row][layout.Position.Column].X + layout.Margin.Left.Pixels,
			Y: cells[layout.Position.Row][layout.Position.Column].Y + layout.Margin.Top.Pixels,
		}

		// Handle rotation
		rotatedImage := rotateImage(inputImage, layout.Rotation)
		rotatedSize := rotatedImage.Bounds().Size()

		// Draw the rotated input image onto the output image
		draw.Draw(outputImage, image.Rect(destPoint.X, destPoint.Y, destPoint.X+rotatedSize.X, destPoint.Y+rotatedSize.Y), rotatedImage, image.Point{}, draw.Over)
	}

	// Draw layout borders and inner layout borders
	for _, layout := range outputPage.Layout {
		cell := cells[layout.Position.Row][layout.Position.Column]
		if outputPage.LayoutBorder != nil && outputPage.LayoutBorder.Enabled {
			drawBorder(outputImage, image.Rect(cell.X, cell.Y, cell.X+cell.Width, cell.Y+cell.Height), outputPage.LayoutBorder.Color.RGBA, outputPage.LayoutBorder.Type)
		}
		if layout.InnerLayoutBorder != nil && layout.InnerLayoutBorder.Enabled {
			innerRect := image.Rect(
				cell.X+layout.Margin.Left.Pixels,
				cell.Y+layout.Margin.Top.Pixels,
				cell.X+cell.Width-layout.Margin.Right.Pixels,
				cell.Y+cell.Height-layout.Margin.Bottom.Pixels,
			)
			drawBorder(outputImage, innerRect, layout.InnerLayoutBorder.Color.RGBA, layout.InnerLayoutBorder.Type)
		}
	}

	// Add global margins to the final image
	finalWidth := width + zl.PageSetup.Margin.Left.Pixels + zl.PageSetup.Margin.Right.Pixels + outputPage.Margin.Left.Pixels + outputPage.Margin.Right.Pixels
	finalHeight := height + zl.PageSetup.Margin.Top.Pixels + zl.PageSetup.Margin.Bottom.Pixels + outputPage.Margin.Top.Pixels + outputPage.Margin.Bottom.Pixels
	finalImage := image.NewRGBA(image.Rect(0, 0, finalWidth, finalHeight))

	// Fill the final image with white color
	draw.Draw(finalImage, finalImage.Bounds(), image.White, image.Point{}, draw.Src)

	// Draw the output image onto the final image with margins
	outputRect := image.Rect(
		zl.PageSetup.Margin.Left.Pixels+outputPage.Margin.Left.Pixels,
		zl.PageSetup.Margin.Top.Pixels+outputPage.Margin.Top.Pixels,
		finalWidth-zl.PageSetup.Margin.Right.Pixels-outputPage.Margin.Right.Pixels,
		finalHeight-zl.PageSetup.Margin.Bottom.Pixels-outputPage.Margin.Bottom.Pixels,
	)
	draw.Draw(finalImage, outputRect, outputImage, image.Point{0, 0}, draw.Over)

	// Draw page border
	if zl.PageSetup.PageBorder != nil && zl.PageSetup.PageBorder.Enabled {
		borderRect := image.Rect(
			zl.PageSetup.Margin.Left.Pixels,
			zl.PageSetup.Margin.Top.Pixels,
			finalWidth-zl.PageSetup.Margin.Right.Pixels,
			finalHeight-zl.PageSetup.Margin.Bottom.Pixels,
		)
		fmt.Printf("Output page border: Top: %d, Bottom: %d, Left: %d, Right: %d, Color: %v, Type: %v\n",
			borderRect.Min.Y, borderRect.Max.Y, borderRect.Min.X, borderRect.Max.X, zl.PageSetup.PageBorder.Color.RGBA, zl.PageSetup.PageBorder.Type)
		drawBorder(finalImage, borderRect, zl.PageSetup.PageBorder.Color.RGBA, zl.PageSetup.PageBorder.Type)
	}

	// Draw global border
	if zl.Global.Border != nil && zl.Global.Border.Enabled {
		drawBorder(finalImage, finalImage.Bounds(), globalBorderColor, zl.Global.Border.Type)
	}

	fmt.Printf("Global Margins - Top: %s, Bottom: %s, Left: %s, Right: %s\n",
		zl.PageSetup.Margin.Top.String(),
		zl.PageSetup.Margin.Bottom.String(),
		zl.PageSetup.Margin.Left.String(),
		zl.PageSetup.Margin.Right.String(),
	)
	fmt.Printf("Output Page Margins - Top: %s, Bottom: %s, Left: %s, Right: %s\n",
		outputPage.Margin.Top.String(),
		outputPage.Margin.Bottom.String(),
		outputPage.Margin.Left.String(),
		outputPage.Margin.Right.String(),
	)

	return finalImage, nil
}

// New function to handle image rotation
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

// Helper function to find the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ParseBorderType converts a string to a BorderType
func ParseBorderType(borderTypeString string) (BorderType, error) {
	switch strings.ToLower(borderTypeString) {
	case "plain":
		return BorderTypePlain, nil
	case "dotted":
		return BorderTypeDotted, nil
	case "dashed":
		return BorderTypeDashed, nil
	case "corner":
		return BorderTypeCorner, nil
	default:
		return "", fmt.Errorf("invalid border type: %s", borderTypeString)
	}
}

func (zl *ZineLayout) ComputeAllMargins() error {
	if zl.PageSetup.Margin == nil {
		zl.PageSetup.Margin = &Margin{}
	}
	margins := []*Margin{
		zl.PageSetup.Margin,
	}

	for i := range zl.OutputPages {
		if zl.OutputPages[i].Margin == nil {
			zl.OutputPages[i].Margin = &Margin{}
		}
		margins = append(margins, zl.OutputPages[i].Margin)
		for j := range zl.OutputPages[i].Layout {
			if zl.OutputPages[i].Layout[j].Margin == nil {
				zl.OutputPages[i].Layout[j].Margin = &Margin{}
			}
			margins = append(margins, zl.OutputPages[i].Layout[j].Margin)
			fmt.Printf("Margin: %+v\n", zl.OutputPages[i].Layout[j].Margin)
		}
	}

	for _, margin := range margins {
		log.Trace().
			Interface("margin", margin).
			Float64("ppi", zl.Global.PPI).
			Msg("Margin before")
		if err := margin.ComputePixelValues(zl.Global.PPI); err != nil {
			return fmt.Errorf("error computing margin values: %w", err)
		}

		log.Trace().
			Interface("margin", margin).
			Msg("Margin after")
	}

	return nil
}
