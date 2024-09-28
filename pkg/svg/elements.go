package svg

import (
	svg "github.com/ajstarks/svgo"
)

// Element is the interface that all SVG elements implement.
type Element interface {
	Render(canvas *svg.SVG)
}

// Rectangle represents an SVG rectangle.
type Rectangle struct {
	Type        string     `yaml:"type"`
	ID          string     `yaml:"id,omitempty"`
	X           int        `yaml:"x"`
	Y           int        `yaml:"y"`
	Width       int        `yaml:"width"`
	Height      int        `yaml:"height"`
	Fill        string     `yaml:"fill,omitempty"`
	Stroke      string     `yaml:"stroke,omitempty"`
	StrokeWidth int        `yaml:"stroke_width,omitempty"`
	Transform   *Transform `yaml:"transform,omitempty"`
}

// Line represents an SVG line.
type Line struct {
	Type        string     `yaml:"type"`
	ID          string     `yaml:"id,omitempty"`
	X1          int        `yaml:"x1"`
	Y1          int        `yaml:"y1"`
	X2          int        `yaml:"x2"`
	Y2          int        `yaml:"y2"`
	Stroke      string     `yaml:"stroke,omitempty"`
	StrokeWidth int        `yaml:"stroke_width,omitempty"`
	Transform   *Transform `yaml:"transform,omitempty"`
}

// Image represents an SVG image.
type Image struct {
	Type      string     `yaml:"type"`
	ID        string     `yaml:"id,omitempty"`
	Href      string     `yaml:"href"`
	X         int        `yaml:"x"`
	Y         int        `yaml:"y"`
	Width     int        `yaml:"width"`
	Height    int        `yaml:"height"`
	Transform *Transform `yaml:"transform,omitempty"`
}

// Text represents an SVG text element.
type Text struct {
	Type       string     `yaml:"type"`
	ID         string     `yaml:"id,omitempty"`
	X          int        `yaml:"x"`
	Y          int        `yaml:"y"`
	Content    string     `yaml:"content"`
	FontSize   string     `yaml:"font_size,omitempty"`
	FontFamily string     `yaml:"font_family,omitempty"`
	Fill       string     `yaml:"fill,omitempty"`
	TextAnchor string     `yaml:"text_anchor,omitempty"`
	Transform  *Transform `yaml:"transform,omitempty"`
}

// Group represents an SVG group, which can contain nested elements.
type Group struct {
	Type      string     `yaml:"type"`
	ID        string     `yaml:"id,omitempty"`
	Transform *Transform `yaml:"transform,omitempty"`
	Elements  []Element  `yaml:"elements"`
}

// Circle represents an SVG circle.
type Circle struct {
	Type        string     `yaml:"type"`
	ID          string     `yaml:"id,omitempty"`
	CX          int        `yaml:"cx"`
	CY          int        `yaml:"cy"`
	R           int        `yaml:"r"`
	Fill        string     `yaml:"fill,omitempty"`
	Stroke      string     `yaml:"stroke,omitempty"`
	StrokeWidth int        `yaml:"stroke_width,omitempty"`
	Transform   *Transform `yaml:"transform,omitempty"`
}

// Render methods for each element type
// ... (same as in the original file)

// Render renders the rectangle onto the SVG canvas.
func (r *Rectangle) Render(canvas *svg.SVG) {
	styles := buildStyles(r.Fill, r.Stroke, r.StrokeWidth)
	if r.Transform != nil {
		canvas.Gtransform(buildTransform(r.Transform))
	}
	canvas.Rect(r.X, r.Y, r.Width, r.Height, styles)
	if r.Transform != nil {
		canvas.Gend()
	}
}

// Render renders the line onto the SVG canvas.
func (l *Line) Render(canvas *svg.SVG) {
	styles := buildStyles("", l.Stroke, l.StrokeWidth)
	if l.Transform != nil {
		canvas.Gtransform(buildTransform(l.Transform))
	}
	canvas.Line(l.X1, l.Y1, l.X2, l.Y2, styles)
	if l.Transform != nil {
		canvas.Gend()
	}
}

// Render renders the image onto the SVG canvas.
func (img *Image) Render(canvas *svg.SVG) {
	styles := buildStyles("", "", 0) // Assuming no additional styles
	if img.Transform != nil {
		canvas.Gtransform(buildTransform(img.Transform))
	}
	canvas.Image(img.X, img.Y, img.Width, img.Height, img.Href, styles)
	if img.Transform != nil {
		canvas.Gend()
	}
}

// Render renders the text onto the SVG canvas.
func (t *Text) Render(canvas *svg.SVG) {
	styles := buildTextStyles(t.Fill, t.FontSize, t.FontFamily, t.TextAnchor)
	if t.Transform != nil {
		canvas.Gtransform(buildTransform(t.Transform))
	}
	canvas.Text(t.X, t.Y, t.Content, styles)
	if t.Transform != nil {
		canvas.Gend()
	}
}

// Render renders the group and its nested elements onto the SVG canvas.
func (g *Group) Render(canvas *svg.SVG) {
	if g.Transform != nil {
		canvas.Gtransform(buildTransform(g.Transform))
	} else {
		canvas.Group()
	}
	for _, elem := range g.Elements {
		elem.Render(canvas)
	}
	canvas.Gend()
}

// Render renders the circle onto the SVG canvas.
func (c *Circle) Render(canvas *svg.SVG) {
	styles := buildStyles(c.Fill, c.Stroke, c.StrokeWidth)
	if c.Transform != nil {
		canvas.Gtransform(buildTransform(c.Transform))
	}
	canvas.Circle(c.CX, c.CY, c.R, styles)
	if c.Transform != nil {
		canvas.Gend()
	}
}
