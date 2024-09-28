package zinelayout

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func GenerateTestImages(count, width, height int) ([]image.Image, error) {
	var images []image.Image

	// Define a slice of 16 pale colors
	paleColors := []color.RGBA{
		{255, 200, 200, 255}, // Less Pale Red
		{200, 255, 200, 255}, // Less Pale Green
		{200, 200, 255, 255}, // Less Pale Blue
		{255, 255, 200, 255}, // Less Pale Yellow
		{255, 200, 255, 255}, // Less Pale Magenta
		{200, 255, 255, 255}, // Less Pale Cyan
		{255, 215, 180, 255}, // Less Pale Seashell
		{215, 255, 220, 255}, // Less Pale Mint Cream
		{200, 208, 255, 255}, // Less Pale Alice Blue
		{255, 220, 200, 255}, // Less Pale Floral White
		{255, 215, 180, 255}, // Less Pale Old Lace
		{215, 215, 215, 255}, // Less Pale White Smoke
		{223, 215, 190, 255}, // Less Pale Old Lace
		{220, 200, 190, 255}, // Less Pale Linen
		{220, 185, 165, 255}, // Less Pale Antique White
		{255, 220, 220, 255}, // Less Pale Snow
	}

	// Use default dimensions if not specified
	if width == 0 {
		width = 600
	}
	if height == 0 {
		height = 600 * 4 / 3
	}

	for i := 1; i <= count; i++ {
		img := image.NewRGBA(image.Rect(0, 0, width, height))

		// Use the (i-1) % 16 to cycle through the colors
		bgColor := paleColors[(i-1)%len(paleColors)]
		draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

		// Add page number to the image
		addLabel(img, fmt.Sprintf("Page %d", i), color.Black)

		images = append(images, img)
	}
	return images, nil
}

func addLabel(img *image.RGBA, label string, textColor color.Color) {
	point := fixed.Point26_6{
		X: fixed.Int26_6(img.Bounds().Dx()/2) << 6,
		Y: fixed.Int26_6(img.Bounds().Dy()/2) << 6,
	}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.Dot.X -= fixed.Int26_6(len(label) * 7 / 2 << 6)
	d.DrawString(label)
}

func GenerateTestImagesBW(count, width, height int) ([]image.Image, error) {
	var images []image.Image

	// Use default dimensions if not specified
	if width == 0 {
		width = 600
	}
	if height == 0 {
		height = 600 * 4 / 3
	}

	for i := 1; i <= count; i++ {
		img := image.NewGray(image.Rect(0, 0, width, height))

		// Alternate between white and light gray background
		bgColor := color.Gray{Y: uint8(255 - (i%2)*20)}
		draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

		// Add page number to the image
		addLabelBW(img, fmt.Sprintf("Page %d", i), color.Black)

		images = append(images, img)
	}
	return images, nil
}

func addLabelBW(img *image.Gray, label string, textColor color.Color) {
	point := fixed.Point26_6{
		X: fixed.Int26_6(img.Bounds().Dx()/2) << 6,
		Y: fixed.Int26_6(img.Bounds().Dy()/2) << 6,
	}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.Dot.X -= fixed.Int26_6(len(label) * 7 / 2 << 6)
	d.DrawString(label)
}
