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

func GenerateTestImages(count int) ([]image.Image, error) {
	var images []image.Image

	// Define a slice of 16 pale colors
	paleColors := []color.RGBA{
		{255, 240, 240, 255}, // Pale Red
		{240, 255, 240, 255}, // Pale Green
		{240, 240, 255, 255}, // Pale Blue
		{255, 255, 240, 255}, // Pale Yellow
		{255, 240, 255, 255}, // Pale Magenta
		{240, 255, 255, 255}, // Pale Cyan
		{255, 245, 238, 255}, // Seashell
		{245, 255, 250, 255}, // Mint Cream
		{240, 248, 255, 255}, // Alice Blue
		{255, 250, 240, 255}, // Floral White
		{255, 245, 238, 255}, // Old Lace
		{245, 245, 245, 255}, // White Smoke
		{253, 245, 230, 255}, // Old Lace
		{250, 240, 230, 255}, // Linen
		{250, 235, 215, 255}, // Antique White
		{255, 250, 250, 255}, // Snow
	}

	for i := 1; i <= count; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 600, 600*4/3))

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
