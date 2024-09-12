package zinelayout

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"image"
	"image/color"
	"image/draw"
	"strconv"
	"strings"
)

// BorderType represents the type of border to draw
type BorderType string

const (
	BorderTypePlain  BorderType = "plain"
	BorderTypeDotted BorderType = "dotted"
	BorderTypeDashed BorderType = "dashed"
	BorderTypeCorner BorderType = "corner"
)

// ZineLayout represents the entire YAML structure
type ZineLayout struct {
	PageSetup         PageSetup    `yaml:"page_setup"`
	OutputPages       []OutputPage `yaml:"output_pages"`
	GlobalBorder      bool         `yaml:"global_border"`
	PageBorder        bool         `yaml:"page_border"`
	LayoutBorder      bool         `yaml:"layout_border"`
	InnerLayoutBorder bool         `yaml:"inner_layout_border"`
	BorderColor       CustomColor  `yaml:"border_color"`
	BorderType        BorderType   `yaml:"border_type"`
}

// PageSetup represents the page setup settings
type PageSetup struct {
	GridSize struct {
		Rows    int `yaml:"rows"`
		Columns int `yaml:"columns"`
	} `yaml:"grid_size"`
	Margin Margin `yaml:"margin"`
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
	Row    int `yaml:"row"`
	Column int `yaml:"column"`
}

// Margin represents margin settings
type Margin struct {
	Top    int `yaml:"top"`
	Bottom int `yaml:"bottom"`
	Left   int `yaml:"left"`
	Right  int `yaml:"right"`
}

// CustomColor is a wrapper around color.RGBA that implements yaml.Unmarshaler
type CustomColor struct {
	color.RGBA
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (c *CustomColor) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		// Handle hex color string or color name
		return c.unmarshalScalar(value.Value)
	} else if value.Kind == yaml.SequenceNode {
		// Handle list of numbers
		return c.unmarshalList(value)
	}
	return fmt.Errorf("invalid color format")
}

func (c *CustomColor) unmarshalScalar(s string) error {
	// Check if it's a hex color
	if strings.HasPrefix(s, "#") {
		return c.unmarshalHex(s)
	}
	// Check if it's a standard color name
	if rgba, ok := standardColors[strings.ToLower(s)]; ok {
		c.RGBA = rgba
		return nil
	}
	return fmt.Errorf("invalid color: %s", s)
}

func (c *CustomColor) unmarshalHex(s string) error {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return fmt.Errorf("invalid hex color: %s", s)
	}
	rgb, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return err
	}
	c.R = uint8(rgb >> 16)
	c.G = uint8((rgb >> 8) & 0xFF)
	c.B = uint8(rgb & 0xFF)
	c.A = 255
	return nil
}

func (c *CustomColor) unmarshalList(node *yaml.Node) error {
	var values []uint8
	if err := node.Decode(&values); err != nil {
		return err
	}
	if len(values) != 3 && len(values) != 4 {
		return fmt.Errorf("invalid color list length: %d", len(values))
	}
	c.R = values[0]
	c.G = values[1]
	c.B = values[2]
	if len(values) == 4 {
		c.A = values[3]
	} else {
		c.A = 255
	}
	return nil
}

// standardColors maps color names to their RGBA values
var standardColors = map[string]color.RGBA{
	"black":     {0, 0, 0, 255},
	"white":     {255, 255, 255, 255},
	"red":       {255, 0, 0, 255},
	"green":     {0, 255, 0, 255},
	"blue":      {0, 0, 255, 255},
	"yellow":    {255, 255, 0, 255},
	"cyan":      {0, 255, 255, 255},
	"magenta":   {255, 0, 255, 255},
	"gray":      {128, 128, 128, 255},
	"grey":      {128, 128, 128, 255},
	"lightgray": {211, 211, 211, 255},
	"lightgrey": {211, 211, 211, 255},
	"darkgray":  {169, 169, 169, 255},
	"darkgrey":  {169, 169, 169, 255},
	"orange":    {255, 165, 0, 255},
	"purple":    {128, 0, 128, 255},
	"brown":     {165, 42, 42, 255},
	"pink":      {255, 192, 203, 255},
	// Add more colors as needed
}

func (zl *ZineLayout) CreateOutputImage(outputPage OutputPage, inputImages []image.Image) (image.Image, error) {
	fmt.Println("Creating output image")
	for _, inputImage := range inputImages {
		fmt.Printf("Input image size: %v\n", inputImage.Bounds().Size())
	}
	inputSize := inputImages[0].Bounds().Size()

	fmt.Printf("Output page margins. Top: %d, Bottom: %d, Left: %d, Right: %d\n", outputPage.Margin.Top, outputPage.Margin.Bottom, outputPage.Margin.Left, outputPage.Margin.Right)

	type CellSize struct {
		Margin Margin
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
			cells[row][column] = CellSize{Margin: Margin{}}
		}
	}

	// Calculate cell sizes and update cells
	for _, layout := range outputPage.Layout {
		row, col := int(layout.Position.Row), int(layout.Position.Column)
		cells[row][col].Margin = layout.Margin
		cells[row][col].Width = inputSize.X + layout.Margin.Left + layout.Margin.Right
		cells[row][col].Height = inputSize.Y + layout.Margin.Top + layout.Margin.Bottom
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
	borderColor := zl.BorderColor.RGBA
	if borderColor == (color.RGBA{}) {
		borderColor = color.RGBA{0, 0, 0, 255} // Default to black
	}

	for _, layout := range outputPage.Layout {
		if layout.Rotation != 0 && layout.Rotation != 180 {
			return nil, fmt.Errorf("invalid rotation %d for input index %d", layout.Rotation, layout.InputIndex)
		}

		inputImage := inputImages[layout.InputIndex-1]
		destPoint := image.Point{
			X: cells[layout.Position.Row][layout.Position.Column].X + layout.Margin.Left,
			Y: cells[layout.Position.Row][layout.Position.Column].Y + layout.Margin.Top,
		}

		// Handle rotation
		rotatedImage := rotateImage(inputImage, layout.Rotation)
		rotatedSize := rotatedImage.Bounds().Size()

		// Draw the rotated input image onto the output image
		draw.Draw(outputImage, image.Rect(destPoint.X, destPoint.Y, destPoint.X+rotatedSize.X, destPoint.Y+rotatedSize.Y), rotatedImage, image.Point{}, draw.Over)
	}

	// Draw layout borders and inner layout borders
	if zl.LayoutBorder || zl.InnerLayoutBorder {
		for _, layout := range outputPage.Layout {
			cell := cells[layout.Position.Row][layout.Position.Column]
			if zl.LayoutBorder {
				drawBorder(outputImage, image.Rect(cell.X, cell.Y, cell.X+cell.Width, cell.Y+cell.Height), borderColor, zl.BorderType)
			}
			if zl.InnerLayoutBorder {
				innerRect := image.Rect(
					cell.X+layout.Margin.Left,
					cell.Y+layout.Margin.Top,
					cell.X+cell.Width-layout.Margin.Right,
					cell.Y+cell.Height-layout.Margin.Bottom,
				)
				drawBorder(outputImage, innerRect, borderColor, zl.BorderType)
			}
		}
	}
	// Add global margins to the final image
	finalWidth := width + zl.PageSetup.Margin.Left + zl.PageSetup.Margin.Right + outputPage.Margin.Left + outputPage.Margin.Right
	finalHeight := height + zl.PageSetup.Margin.Top + zl.PageSetup.Margin.Bottom + outputPage.Margin.Top + outputPage.Margin.Bottom
	finalImage := image.NewRGBA(image.Rect(0, 0, finalWidth, finalHeight))

	// Fill the final image with white color
	draw.Draw(finalImage, finalImage.Bounds(), image.White, image.Point{}, draw.Src)

	// Draw the output image onto the final image with margins
	outputRect := image.Rect(
		zl.PageSetup.Margin.Left+outputPage.Margin.Left,
		zl.PageSetup.Margin.Top+outputPage.Margin.Top,
		finalWidth-zl.PageSetup.Margin.Right-outputPage.Margin.Right,
		finalHeight-zl.PageSetup.Margin.Bottom-outputPage.Margin.Bottom,
	)
	draw.Draw(finalImage, outputRect, outputImage, image.Point{0, 0}, draw.Over)

	// Draw page border
	if zl.PageBorder {
		// Draw the output image onto the final image with margins
		borderRect := image.Rect(
			zl.PageSetup.Margin.Left,
			zl.PageSetup.Margin.Top,
			finalWidth-zl.PageSetup.Margin.Right,
			finalHeight-zl.PageSetup.Margin.Bottom,
		)
		fmt.Printf("Output page border: Top: %d, Bottom: %d, Left: %d, Right: %d\n",
			borderRect.Min.Y, borderRect.Max.Y, borderRect.Min.X, borderRect.Max.X)
		drawBorder(finalImage, borderRect, borderColor, zl.BorderType)
	}

	// Draw global border
	if zl.GlobalBorder {
		drawBorder(finalImage, finalImage.Bounds(), borderColor, zl.BorderType)
	}

	fmt.Printf("Global Margins - Top: %d, Bottom: %d, Left: %d, Right: %d\n", zl.PageSetup.Margin.Top, zl.PageSetup.Margin.Bottom, zl.PageSetup.Margin.Left, zl.PageSetup.Margin.Right)
	fmt.Printf("Output Page Margins - Top: %d, Bottom: %d, Left: %d, Right: %d\n", outputPage.Margin.Top, outputPage.Margin.Bottom, outputPage.Margin.Left, outputPage.Margin.Right)

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

// Updated function to draw a border
func drawBorder(img *image.RGBA, rect image.Rectangle, c color.Color, borderType BorderType) {
	switch borderType {
	case BorderTypePlain:
		drawPlainBorder(img, rect, c)
	case BorderTypeDotted:
		drawDottedBorder(img, rect, c)
	case BorderTypeDashed:
		drawDashedBorder(img, rect, c)
	case BorderTypeCorner:
		drawCornerBorder(img, rect, c)
	}
}

func drawPlainBorder(img *image.RGBA, rect image.Rectangle, c color.Color) {
	for x := rect.Min.X; x < rect.Max.X; x++ {
		img.Set(x, rect.Min.Y, c)
		img.Set(x, rect.Max.Y-1, c)
	}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		img.Set(rect.Min.X, y, c)
		img.Set(rect.Max.X-1, y, c)
	}
}

func drawDottedBorder(img *image.RGBA, rect image.Rectangle, c color.Color) {
	for x := rect.Min.X; x < rect.Max.X; x += 2 {
		img.Set(x, rect.Min.Y, c)
		img.Set(x, rect.Max.Y-1, c)
	}
	for y := rect.Min.Y; y < rect.Max.Y; y += 2 {
		img.Set(rect.Min.X, y, c)
		img.Set(rect.Max.X-1, y, c)
	}
}

func drawDashedBorder(img *image.RGBA, rect image.Rectangle, c color.Color) {
	dashLength := 4
	for x := rect.Min.X; x < rect.Max.X; x++ {
		if (x-rect.Min.X)%dashLength < dashLength/2 {
			img.Set(x, rect.Min.Y, c)
			img.Set(x, rect.Max.Y-1, c)
		}
	}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		if (y-rect.Min.Y)%dashLength < dashLength/2 {
			img.Set(rect.Min.X, y, c)
			img.Set(rect.Max.X-1, y, c)
		}
	}
}

func drawCornerBorder(img *image.RGBA, rect image.Rectangle, c color.Color) {
	cornerLength := 20 // Length of corner dashes in pixels

	// Top-left corner
	drawLine(img, rect.Min.X, rect.Min.Y, rect.Min.X-cornerLength, rect.Min.Y, c)
	drawLine(img, rect.Min.X, rect.Min.Y, rect.Min.X, rect.Min.Y-cornerLength, c)

	// Top-right corner
	drawLine(img, rect.Max.X-1, rect.Min.Y, rect.Max.X-1+cornerLength, rect.Min.Y, c)
	drawLine(img, rect.Max.X-1, rect.Min.Y, rect.Max.X-1, rect.Min.Y-cornerLength, c)

	// Bottom-left corner
	drawLine(img, rect.Min.X, rect.Max.Y-1, rect.Min.X-cornerLength, rect.Max.Y-1, c)
	drawLine(img, rect.Min.X, rect.Max.Y-1, rect.Min.X, rect.Max.Y-1+cornerLength, c)

	// Bottom-right corner
	drawLine(img, rect.Max.X-1, rect.Max.Y-1, rect.Max.X-1+cornerLength, rect.Max.Y-1, c)
	drawLine(img, rect.Max.X-1, rect.Max.Y-1, rect.Max.X-1, rect.Max.Y-1+cornerLength, c)
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	bounds := img.Bounds()
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 >= x2 {
		sx = -1
	}
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy

	for {
		if x1 >= bounds.Min.X && x1 < bounds.Max.X && y1 >= bounds.Min.Y && y1 < bounds.Max.Y {
			img.Set(x1, y1, c)
		}
		if x1 == x2 && y1 == y2 {
			return
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
