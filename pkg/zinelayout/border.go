package zinelayout

import (
	"image"
	"image/color"
)

// BorderType represents the type of border to draw
type BorderType string

const (
	BorderTypePlain  BorderType = "plain"
	BorderTypeDotted BorderType = "dotted"
	BorderTypeDashed BorderType = "dashed"
	BorderTypeCorner BorderType = "corner"
)

// drawBorder draws a border on the image based on the specified type
func drawBorder(img *image.RGBA, rect image.Rectangle, c color.Color, borderType BorderType) {
	// If color is 0 0 0 0, make it black 0 0 0 255
	if c == (color.RGBA{0, 0, 0, 0}) {
		c = color.RGBA{0, 0, 0, 255}
	}

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
