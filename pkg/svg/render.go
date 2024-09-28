package svg

import (
	"bytes"
	"fmt"

	svg "github.com/ajstarks/svgo"
)

// buildStyles constructs the style string for fill, stroke, and stroke-width.
func buildStyles(fill, stroke string, strokeWidth int) string {
	styles := ""
	if fill != "" {
		styles += fmt.Sprintf("fill:%s;", fill)
	}
	if stroke != "" {
		styles += fmt.Sprintf("stroke:%s;", stroke)
	}
	if strokeWidth > 0 {
		styles += fmt.Sprintf("stroke-width:%d;", strokeWidth)
	}
	return styles
}

// buildTextStyles constructs the style string for text elements.
func buildTextStyles(fill, fontSize, fontFamily, textAnchor string) string {
	styles := ""
	if fill != "" {
		styles += fmt.Sprintf("fill:%s;", fill)
	}
	if fontSize != "" {
		styles += fmt.Sprintf("font-size:%s;", fontSize)
	}
	if fontFamily != "" {
		styles += fmt.Sprintf("font-family:%s;", fontFamily)
	}
	if textAnchor != "" {
		styles += fmt.Sprintf("text-anchor:%s;", textAnchor)
	}
	return styles
}

// RenderSVG renders the SVG based on the Canvas configuration
func RenderSVG(canvas *Canvas) (string, error) {
	var buf bytes.Buffer
	s := svg.New(&buf)

	s.Start(canvas.Width, canvas.Height)

	// Set background
	if canvas.Background.Color != "" {
		s.Rect(0, 0, canvas.Width, canvas.Height, "fill:"+canvas.Background.Color)
	}
	if canvas.Background.Image != "" {
		s.Image(0, 0, canvas.Width, canvas.Height, canvas.Background.Image)
	}

	// Render elements
	for _, elem := range canvas.GetElements() {
		elem.Render(s)
	}

	s.End()

	return buf.String(), nil
}
