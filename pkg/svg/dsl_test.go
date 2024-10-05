package svg

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestUnmarshalRectangle(t *testing.T) {
	yamlInput := `
type: rectangle
id: rect1
x: 10
y: 20
width: 100
height: 50
fill: "#ff0000"
stroke: "#000000"
stroke_width: 2
`
	var element ElementWrapper
	err := yaml.Unmarshal([]byte(yamlInput), &element)
	if err != nil {
		t.Fatalf("Failed to unmarshal Rectangle: %v", err)
	}

	rect, ok := element.Element.(*Rectangle)
	if !ok {
		t.Fatalf("Expected Rectangle, got %T", element.Element)
	}

	expected := &Rectangle{
		Type:        "rectangle",
		ID:          "rect1",
		X:           10,
		Y:           20,
		Width:       100,
		Height:      50,
		Fill:        "#ff0000",
		Stroke:      "#000000",
		StrokeWidth: 2,
	}

	if !reflect.DeepEqual(rect, expected) {
		t.Errorf("Unmarshaled Rectangle doesn't match expected. Got %+v, want %+v", rect, expected)
	}
}

func TestUnmarshalLine(t *testing.T) {
	yamlInput := `
type: line
id: line1
x1: 0
y1: 0
x2: 100
y2: 100
stroke: "#0000ff"
stroke_width: 3
`
	var element ElementWrapper
	err := yaml.Unmarshal([]byte(yamlInput), &element)
	if err != nil {
		t.Fatalf("Failed to unmarshal Line: %v", err)
	}

	line, ok := element.Element.(*Line)
	if !ok {
		t.Fatalf("Expected Line, got %T", element.Element)
	}

	expected := &Line{
		Type:        "line",
		ID:          "line1",
		X1:          0,
		Y1:          0,
		X2:          100,
		Y2:          100,
		Stroke:      "#0000ff",
		StrokeWidth: 3,
	}

	if !reflect.DeepEqual(line, expected) {
		t.Errorf("Unmarshaled Line doesn't match expected. Got %+v, want %+v", line, expected)
	}
}

func TestUnmarshalImage(t *testing.T) {
	yamlInput := `
type: image
id: img1
href: "https://example.com/image.png"
x: 50
y: 50
width: 200
height: 150
`
	var element ElementWrapper
	err := yaml.Unmarshal([]byte(yamlInput), &element)
	if err != nil {
		t.Fatalf("Failed to unmarshal Image: %v", err)
	}

	img, ok := element.Element.(*Image)
	if !ok {
		t.Fatalf("Expected Image, got %T", element.Element)
	}

	expected := &Image{
		Type:   "image",
		ID:     "img1",
		Href:   "https://example.com/image.png",
		X:      50,
		Y:      50,
		Width:  200,
		Height: 150,
	}

	if !reflect.DeepEqual(img, expected) {
		t.Errorf("Unmarshaled Image doesn't match expected. Got %+v, want %+v", img, expected)
	}
}

func TestUnmarshalText(t *testing.T) {
	yamlInput := `
type: text
id: text1
x: 100
y: 100
content: "Hello, SVG!"
font_size: "24px"
font_family: "Arial"
fill: "#000000"
text_anchor: "middle"
`
	var element ElementWrapper
	err := yaml.Unmarshal([]byte(yamlInput), &element)
	if err != nil {
		t.Fatalf("Failed to unmarshal Text: %v", err)
	}

	text, ok := element.Element.(*Text)
	if !ok {
		t.Fatalf("Expected Text, got %T", element.Element)
	}

	expected := &Text{
		Type:       "text",
		ID:         "text1",
		X:          100,
		Y:          100,
		Content:    "Hello, SVG!",
		FontSize:   "24px",
		FontFamily: "Arial",
		Fill:       "#000000",
		TextAnchor: "middle",
	}

	if !reflect.DeepEqual(text, expected) {
		t.Errorf("Unmarshaled Text doesn't match expected. Got %+v, want %+v", text, expected)
	}
}

func TestUnmarshalCircle(t *testing.T) {
	yamlInput := `
type: circle
id: circle1
cx: 150
cy: 150
r: 50
fill: "#00ff00"
stroke: "#000000"
stroke_width: 2
`
	var element ElementWrapper
	err := yaml.Unmarshal([]byte(yamlInput), &element)
	if err != nil {
		t.Fatalf("Failed to unmarshal Circle: %v", err)
	}

	circle, ok := element.Element.(*Circle)
	if !ok {
		t.Fatalf("Expected Circle, got %T", element.Element)
	}

	expected := &Circle{
		Type:        "circle",
		ID:          "circle1",
		CX:          150,
		CY:          150,
		R:           50,
		Fill:        "#00ff00",
		Stroke:      "#000000",
		StrokeWidth: 2,
	}

	if !reflect.DeepEqual(circle, expected) {
		t.Errorf("Unmarshaled Circle doesn't match expected. Got %+v, want %+v", circle, expected)
	}
}

func TestUnmarshalGroup(t *testing.T) {
	yamlInput := `
type: group
id: group1
transform:
  translate: [10, 20]
  rotate: 45
elements:
  - type: rectangle
    x: 0
    y: 0
    width: 50
    height: 50
    fill: "#ff0000"
  - type: circle
    cx: 25
    cy: 25
    r: 25
    fill: "#0000ff"
`
	var element ElementWrapper
	err := yaml.Unmarshal([]byte(yamlInput), &element)
	if err != nil {
		t.Fatalf("Failed to unmarshal Group: %v", err)
	}

	group, ok := element.Element.(*Group)
	if !ok {
		t.Fatalf("Expected Group, got %T", element.Element)
	}

	expected := &Group{
		Type: "group",
		ID:   "group1",
		Transform: &Transform{
			Translate: []int{10, 20},
			Rotate:    45,
		},
		Elements: []Element{
			&Rectangle{
				Type:   "rectangle",
				X:      0,
				Y:      0,
				Width:  50,
				Height: 50,
				Fill:   "#ff0000",
			},
			&Circle{
				Type: "circle",
				CX:   25,
				CY:   25,
				R:    25,
				Fill: "#0000ff",
			},
		},
	}

	if !reflect.DeepEqual(group, expected) {
		t.Errorf("Unmarshaled Group doesn't match expected. Got %+v, want %+v", group, expected)
	}
}

func TestUnmarshalCanvas(t *testing.T) {
	yamlInput := `
svg:
  width: 800
  height: 600
  background:
    color: "#f0f0f0"
  elements:
    - type: rectangle
      x: 10
      y: 10
      width: 100
      height: 50
      fill: "#ff0000"
    - type: circle
      cx: 200
      cy: 200
      r: 30
      fill: "#00ff00"
    - type: triangle
      points:
        - [300, 300]
        - [350, 300]
        - [325, 350]
      fill: "#0000ff"
    - type: ellipse
      cx: 500
      cy: 300
      rx: 60
      ry: 40
      fill: "#ffff00"
    - type: polygon
      points:
        - [600, 100]
        - [650, 150]
        - [600, 200]
        - [550, 150]
      fill: "#ff00ff"
`
	var svgDSL SVGDSL
	err := yaml.Unmarshal([]byte(yamlInput), &svgDSL)
	if err != nil {
		t.Fatalf("Failed to unmarshal Canvas: %v", err)
	}

	expected := SVGDSL{
		SVG: Canvas{
			Width:  800,
			Height: 600,
			Background: Background{
				Color: "#f0f0f0",
			},
			Elements: []ElementWrapper{
				{Element: &Rectangle{
					Type:   "rectangle",
					X:      10,
					Y:      10,
					Width:  100,
					Height: 50,
					Fill:   "#ff0000",
				}},
				{Element: &Circle{
					Type: "circle",
					CX:   200,
					CY:   200,
					R:    30,
					Fill: "#00ff00",
				}},
				{Element: &Triangle{
					Type:   "triangle",
					Points: [][2]int{{300, 300}, {350, 300}, {325, 350}},
					Fill:   "#0000ff",
				}},
				{Element: &Ellipse{
					Type: "ellipse",
					CX:   500,
					CY:   300,
					RX:   60,
					RY:   40,
					Fill: "#ffff00",
				}},
				{Element: &Polygon{
					Type:   "polygon",
					Points: [][2]int{{600, 100}, {650, 150}, {600, 200}, {550, 150}},
					Fill:   "#ff00ff",
				}},
			},
		},
	}

	if !reflect.DeepEqual(svgDSL, expected) {
		t.Errorf("Unmarshaled Canvas doesn't match expected. Got %+v, want %+v", svgDSL, expected)
	}
}

func TestUnmarshalInvalidType(t *testing.T) {
	yamlInput := `
type: invalid_type
id: invalid1
`
	var element ElementWrapper
	err := yaml.Unmarshal([]byte(yamlInput), &element)
	if err == nil {
		t.Fatalf("Expected error for invalid type, got nil")
	}

	expectedError := "unsupported element type: invalid_type"
	if err.Error() != expectedError {
		t.Errorf("Unexpected error message. Got %q, want %q", err.Error(), expectedError)
	}
}

func TestUnmarshalTriangle(t *testing.T) {
	yamlInput := `
type: triangle
id: triangle1
points:
  - [100, 100]
  - [200, 100]
  - [150, 200]
fill: "#00ff00"
stroke: "#000000"
stroke_width: 2
`
	var element ElementWrapper
	err := yaml.Unmarshal([]byte(yamlInput), &element)
	if err != nil {
		t.Fatalf("Failed to unmarshal Triangle: %v", err)
	}

	triangle, ok := element.Element.(*Triangle)
	if !ok {
		t.Fatalf("Expected Triangle, got %T", element.Element)
	}

	expected := &Triangle{
		Type:        "triangle",
		ID:          "triangle1",
		Points:      [][2]int{{100, 100}, {200, 100}, {150, 200}},
		Fill:        "#00ff00",
		Stroke:      "#000000",
		StrokeWidth: 2,
	}

	if !reflect.DeepEqual(triangle, expected) {
		t.Errorf("Unmarshaled Triangle doesn't match expected. Got %+v, want %+v", triangle, expected)
	}
}

func TestUnmarshalEllipse(t *testing.T) {
	yamlInput := `
type: ellipse
id: ellipse1
cx: 200
cy: 150
rx: 100
ry: 50
fill: "#ffff00"
stroke: "#000000"
stroke_width: 2
`
	var element ElementWrapper
	err := yaml.Unmarshal([]byte(yamlInput), &element)
	if err != nil {
		t.Fatalf("Failed to unmarshal Ellipse: %v", err)
	}

	ellipse, ok := element.Element.(*Ellipse)
	if !ok {
		t.Fatalf("Expected Ellipse, got %T", element.Element)
	}

	expected := &Ellipse{
		Type:        "ellipse",
		ID:          "ellipse1",
		CX:          200,
		CY:          150,
		RX:          100,
		RY:          50,
		Fill:        "#ffff00",
		Stroke:      "#000000",
		StrokeWidth: 2,
	}

	if !reflect.DeepEqual(ellipse, expected) {
		t.Errorf("Unmarshaled Ellipse doesn't match expected. Got %+v, want %+v", ellipse, expected)
	}
}

func TestUnmarshalPolygon(t *testing.T) {
	yamlInput := `
type: polygon
id: polygon1
points:
  - [100, 100]
  - [200, 100]
  - [250, 200]
  - [150, 250]
  - [50, 200]
fill: "#ff00ff"
stroke: "#000000"
stroke_width: 2
`
	var element ElementWrapper
	err := yaml.Unmarshal([]byte(yamlInput), &element)
	if err != nil {
		t.Fatalf("Failed to unmarshal Polygon: %v", err)
	}

	polygon, ok := element.Element.(*Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", element.Element)
	}

	expected := &Polygon{
		Type:        "polygon",
		ID:          "polygon1",
		Points:      [][2]int{{100, 100}, {200, 100}, {250, 200}, {150, 250}, {50, 200}},
		Fill:        "#ff00ff",
		Stroke:      "#000000",
		StrokeWidth: 2,
	}

	if !reflect.DeepEqual(polygon, expected) {
		t.Errorf("Unmarshaled Polygon doesn't match expected. Got %+v, want %+v", polygon, expected)
	}
}
