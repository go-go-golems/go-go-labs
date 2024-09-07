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
	for i := 1; i <= count; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 800, 600))
		draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

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