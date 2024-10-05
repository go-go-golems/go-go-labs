## Table of Contents

1. [Introduction](#introduction)
2. [DSL Specification](#dsl-specification)
   - [Overall Structure](#overall-structure)
   - [Canvas Configuration](#canvas-configuration)
   - [Elements](#elements)
     - [Rectangle](#rectangle)
     - [Line](#line)
     - [Image](#image)
     - [Text](#text)
     - [Group](#group)
     - [Triangle](#triangle)
     - [Ellipse](#ellipse)
     - [Polygon](#polygon)
   - [Transformations](#transformations)
   - [Attributes Overview](#attributes-overview)
3. [Code Snippets](#code-snippets)
4. [Examples](#examples)
   - [Example 1: Basic Canvas with a Rectangle](#example-1-basic-canvas-with-a-rectangle)
   - [Example 2: Adding a Line and Background Color](#example-2-adding-a-line-and-background-color)
   - [Example 3: Inserting an Image and Text](#example-3-inserting-an-image-and-text)
   - [Example 4: Grouping Elements with Transformations](#example-4-grouping-elements-with-transformations)
   - [Example 5: Complex Scene with Multiple Groups and Text](#example-5-complex-scene-with-multiple-groups-and-text)
5. [Conclusion](#conclusion)

---

## Introduction

The YAML-based DSL for SVG simplifies the creation and management of SVG graphics by providing a more readable and maintainable format. By abstracting the verbose SVG XML syntax, this DSL allows developers and designers to define complex graphics with ease. The inclusion of text elements further broadens its applicability, enabling the addition of labels, annotations, and dynamic text content within SVGs.

---

## DSL Specification

### Overall Structure

The DSL is structured hierarchically, starting with the `svg` root element, which encapsulates the canvas configuration and a list of drawable `elements`. Each element can be a basic shape (like a rectangle or line), an image, text, or a group that can contain nested elements with scoped transformations.

```yaml
svg:
  width: 800
  height: 600
  background:
    color: "#ffffff"
  elements:
    - type: rectangle
      # Rectangle attributes
    - type: line
      # Line attributes
    - type: image
      # Image attributes
    - type: text
      # Text attributes
    - type: group
      # Group attributes and nested elements
```

### Canvas Configuration

Defines the size and background of the SVG canvas.

- **width**: Width of the canvas in pixels.
- **height**: Height of the canvas in pixels.
- **background**: Optional background settings.
  - **color**: Hex code for background color.
  - **image**: Path or URL to a background image.

```yaml
svg:
  width: 800
  height: 600
  background:
    color: "#f0f0f0"
    # image: "path/to/background.png"
```

### Elements

An array of drawable objects within the SVG canvas. The following are the only supported element types:

- **rectangle**
- **line**
- **image**
- **text**
- **group**
- **triangle**
- **ellipse**
- **polygon**

Each element type has specific attributes (detailed below). No other element types are supported in this DSL.

```yaml
elements:
  - type: rectangle
    # Rectangle attributes
  - type: line
    # Line attributes
  - type: image
    # Image attributes
  - type: text
    # Text attributes
  - type: group
    # Group attributes and nested elements
```

#### Rectangle

Draws a rectangle on the canvas.

- **type**: `"rectangle"`
- **id**: (Optional) Identifier for the element.
- **x**: X-coordinate of the top-left corner.
- **y**: Y-coordinate of the top-left corner.
- **width**: Width of the rectangle.
- **height**: Height of the rectangle.
- **fill**: Fill color (hex code or `none`).
- **stroke**: Border color (hex code or `none`).
- **stroke_width**: Thickness of the border.

```yaml
- type: rectangle
  id: rect1
  x: 50
  y: 50
  width: 200
  height: 100
  fill: "#ff0000"
  stroke: "#000000"
  stroke_width: 2
```

#### Line

Draws a straight line between two points.

- **type**: `"line"`
- **id**: (Optional) Identifier for the element.
- **x1**: X-coordinate of the start point.
- **y1**: Y-coordinate of the start point.
- **x2**: X-coordinate of the end point.
- **y2**: Y-coordinate of the end point.
- **stroke**: Line color.
- **stroke_width**: Thickness of the line.

```yaml
- type: line
  id: line1
  x1: 300
  y1: 300
  x2: 400
  y2: 400
  stroke: "#00ff00"
  stroke_width: 1
```

#### Image

Embeds an image (bitmap) into the SVG.

- **type**: `"image"`
- **id**: (Optional) Identifier for the element.
- **href**: Path or URL to the image file.
- **x**: X-coordinate of the top-left corner.
- **y**: Y-coordinate of the top-left corner.
- **width**: Width of the image.
- **height**: Height of the image.

```yaml
- type: image
  id: img1
  href: "path/to/bitmap.png"
  x: 500
  y: 500
  width: 100
  height: 100
```

#### Text

Adds text to the SVG canvas.

- **type**: `"text"`
- **id**: (Optional) Identifier for the element.
- **x**: X-coordinate of the text's starting point.
- **y**: Y-coordinate of the text's baseline.
- **content**: The text string to display.
- **font_size**: Size of the font (e.g., `"16px"`).
- **font_family**: Font family (e.g., `"Arial"`).
- **fill**: Text color.
- **text_anchor**: Alignment of the text (`"start"`, `"middle"`, `"end"`).
- **transform**: (Optional) Transformations specific to the text element.

```yaml
- type: text
  id: text1
  x: 100
  y: 200
  content: "Hello, SVG!"
  font_size: "24px"
  font_family: "Arial"
  fill: "#000000"
  text_anchor: "middle"
```

#### Group

Groups multiple elements together, allowing for scoped transformations.

- **type**: `"group"`
- **id**: (Optional) Identifier for the group.
- **transform**: Transformation applied to the entire group.
  - **translate**: `[x, y]` pixels to move.
  - **rotate**: Degrees to rotate.
  - **scale**: `[x, y]` scaling factors.
- **elements**: Nested array of drawable elements.

```yaml
- type: group
  id: group1
  transform:
    translate: [100, 100]
    rotate: 45
    scale: [1.5, 1.5]
  elements:
    - type: rectangle
      # Nested rectangle attributes
    - type: line
      # Nested line attributes
```

#### Ellipse

Draws an ellipse on the canvas.

- **type**: `"ellipse"`
- **id**: (Optional) Identifier for the element.
- **cx**: X-coordinate of the center.
- **cy**: Y-coordinate of the center.
- **rx**: Horizontal radius.
- **ry**: Vertical radius.
- **fill**: Fill color (hex code or `none`).
- **stroke**: Border color (hex code or `none`).
- **stroke_width**: Thickness of the border.

```yaml
- type: ellipse
  id: ellipse1
  cx: 200
  cy: 150
  rx: 100
  ry: 50
  fill: "#ffff00"
  stroke: "#000000"
  stroke_width: 2
```

#### Triangle

Draws a triangle on the canvas.

- **type**: `"triangle"`
- **id**: (Optional) Identifier for the element.
- **points**: Array of three [x, y] coordinate pairs.
- **fill**: Fill color (hex code or `none`).
- **stroke**: Border color (hex code or `none`).
- **stroke_width**: Thickness of the border.

```yaml
- type: triangle
  id: triangle1
  points: [[100, 100], [200, 100], [150, 200]]
  fill: "#00ff00"
  stroke: "#000000"
  stroke_width: 2
```

#### Polygon

Draws a polygon on the canvas.

- **type**: `"polygon"`
- **id**: (Optional) Identifier for the element.
- **points**: Array of [x, y] coordinate pairs (minimum 3 points).
- **fill**: Fill color (hex code or `none`).
- **stroke**: Border color (hex code or `none`).
- **stroke_width**: Thickness of the border.

```yaml
- type: polygon
  id: polygon1
  points: [[100, 100], [200, 100], [250, 200], [150, 250], [50, 200]]
  fill: "#ff00ff"
  stroke: "#000000"
  stroke_width: 2
```

### Transformations

Transformations can be applied to individual elements or groups to manipulate their position, rotation, and scale.

- **translate**: Moves the element by `[x, y]` pixels.
- **rotate**: Rotates the element by degrees.
- **scale**: Scales the element by `[x, y]` factors.

Transformations are specified within the `transform` attribute and can be combined as needed.

```yaml
transform:
  translate: [50, 100]
  rotate: 30
  scale: [2, 2]
```

### Attributes Overview

While each element type has specific attributes, some common attributes across multiple elements include:

- **id**: A unique identifier for referencing.
- **transform**: Scoped transformations for positioning and manipulating elements.

---

## Code Snippets

### Rectangle Example

```yaml
- type: rectangle
  id: rect1
  x: 50
  y: 50
  width: 200
  height: 100
  fill: "#ff0000"
  stroke: "#000000"
  stroke_width: 2
```

### Line Example

```yaml
- type: line
  id: line1
  x1: 300
  y1: 300
  x2: 400
  y2: 400
  stroke: "#00ff00"
  stroke_width: 1
```

### Image Example

```yaml
- type: image
  id: img1
  href: "path/to/bitmap.png"
  x: 500
  y: 500
  width: 100
  height: 100
```

### Text Example

```yaml
- type: text
  id: text1
  x: 100
  y: 200
  content: "Hello, SVG!"
  font_size: "24px"
  font_family: "Arial"
  fill: "#000000"
  text_anchor: "middle"
```

### Group Example

```yaml
- type: group
  id: group1
  transform:
    translate: [100, 100]
    rotate: 45
    scale: [1.5, 1.5]
  elements:
    - type: rectangle
      x: 0
      y: 0
      width: 50
      height: 50
      fill: "#0000ff"
    - type: line
      x1: 10
      y1: 10
      x2: 40
      y2: 40
      stroke: "#ff00ff"
      stroke_width: 2
```

### Ellipse Example

```yaml
- type: ellipse
  id: ellipse1
  cx: 200
  cy: 150
  rx: 100
  ry: 50
  fill: "#ffff00"
  stroke: "#000000"
  stroke_width: 2
```

### Triangle Example

```yaml
- type: triangle
  id: triangle1
  points: [[100, 100], [200, 100], [150, 200]]
  fill: "#00ff00"
  stroke: "#000000"
  stroke_width: 2
```

### Polygon Example

```yaml
- type: polygon
  id: polygon1
  points: [[100, 100], [200, 100], [250, 200], [150, 250], [50, 200]]
  fill: "#ff00ff"
  stroke: "#000000"
  stroke_width: 2
```

---

## Examples

Below are five examples demonstrating the YAML DSL in action, each increasing in complexity and showcasing different features.

### Example 1: Basic Canvas with a Rectangle

**YAML Input**

```yaml
svg:
  width: 400
  height: 300
  background:
    color: "#e0e0e0"

  elements:
    - type: rectangle
      x: 50
      y: 50
      width: 300
      height: 200
      fill: "#ff5733"
      stroke: "#000000"
      stroke_width: 3
```

**Description**

Creates a simple SVG canvas of size 400x300 pixels with a light gray background. A single orange rectangle with a black border is drawn within the canvas.

**Corresponding SVG Output**

```svg
<svg width="400" height="300" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%" height="100%" fill="#e0e0e0" />
  <rect x="50" y="50" width="300" height="200" fill="#ff5733" stroke="#000000" stroke-width="3" />
</svg>
```

---

### Example 2: Adding a Line and Background Color

**YAML Input**

```yaml
svg:
  width: 500
  height: 400
  background:
    color: "#ffffff"

  elements:
    - type: rectangle
      x: 100
      y: 100
      width: 300
      height: 200
      fill: "#4caf50"
      stroke: "#2e7d32"
      stroke_width: 4

    - type: line
      x1: 100
      y1: 100
      x2: 400
      y2: 300
      stroke: "#ff0000"
      stroke_width: 2
```

**Description**

Sets up a 500x400 pixel white canvas with a green rectangle and a red diagonal line crossing it from the top-left to the bottom-right.

**Corresponding SVG Output**

```svg
<svg width="500" height="400" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%" height="100%" fill="#ffffff" />
  <rect x="100" y="100" width="300" height="200" fill="#4caf50" stroke="#2e7d32" stroke-width="4" />
  <line x1="100" y1="100" x2="400" y2="300" stroke="#ff0000" stroke-width="2" />
</svg>
```

---

### Example 3: Inserting an Image and Text

**YAML Input**

```yaml
svg:
  width: 600
  height: 400
  background:
    color: "#fafafa"

  elements:
    - type: image
      href: "https://example.com/logo.png"
      x: 50
      y: 50
      width: 100
      height: 100

    - type: text
      x: 200
      y: 100
      content: "Welcome to SVG DSL"
      font_size: "24px"
      font_family: "Verdana"
      fill: "#333333"
      text_anchor: "start"

    - type: line
      x1: 200
      y1: 110
      x2: 550
      y2: 110
      stroke: "#000000"
      stroke_width: 1
```

**Description**

Creates a 600x400 pixel canvas with a light background. An image (e.g., a logo) is placed at the top-left, accompanied by a welcoming text and an underline.

**Corresponding SVG Output**

```svg
<svg width="600" height="400" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%" height="100%" fill="#fafafa" />
  <image href="https://example.com/logo.png" x="50" y="50" width="100" height="100" />
  <text x="200" y="100" font-size="24px" font-family="Verdana" fill="#333333" text-anchor="start">Welcome to SVG DSL</text>
  <line x1="200" y1="110" x2="550" y2="110" stroke="#000000" stroke-width="1" />
</svg>
```

---

### Example 4: Grouping Elements with Transformations

**YAML Input**

```yaml
svg:
  width: 500
  height: 500
  background:
    color: "#ffffff"

  elements:
    - type: group
      id: rotatedGroup
      transform:
        translate: [250, 250]
        rotate: 45
      elements:
        - type: rectangle
          x: -50
          y: -50
          width: 100
          height: 100
          fill: "#2196f3"
          stroke: "#1976d2"
          stroke_width: 2

        - type: line
          x1: -50
          y1: 0
          x2: 50
          y2: 0
          stroke: "#ffffff"
          stroke_width: 4
```

**Description**

Sets up a 500x500 pixel white canvas with a group of elements centered at (250, 250). The group is rotated by 45 degrees. Inside the group, a blue square is drawn with a white horizontal line through its center.

**Corresponding SVG Output**

```svg
<svg width="500" height="500" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%" height="100%" fill="#ffffff" />
  <g transform="translate(250,250) rotate(45)">
    <rect x="-50" y="-50" width="100" height="100" fill="#2196f3" stroke="#1976d2" stroke-width="2" />
    <line x1="-50" y1="0" x2="50" y2="0" stroke="#ffffff" stroke-width="4" />
  </g>
</svg>
```

---

### Example 5: Complex Scene with Multiple Groups and Text

**YAML Input**

```yaml
svg:
  width: 800
  height: 600
  background:
    image: "https://example.com/background.jpg"

  elements:
    - type: group
      id: mainGroup
      transform:
        translate: [400, 300]
      elements:
        - type: rectangle
          x: -150
          y: -100
          width: 300
          height: 200
          fill: "#ffeb3b"
          stroke: "#fbc02d"
          stroke_width: 5

        - type: text
          x: 0
          y: -120
          content: "Complex SVG Scene"
          font_size: "32px"
          font_family: "Helvetica"
          fill: "#000000"
          text_anchor: "middle"

        - type: group
          id: nestedGroup
          transform:
            rotate: -30
          elements:
            - type: line
              x1: -100
              y1: 0
              x2: 100
              y2: 0
              stroke: "#00bcd4"
              stroke_width: 3

            - type: text
              x: 0
              y: 20
              content: "Nested Group"
              font_size: "16px"
              font_family: "Courier New"
              fill: "#e91e63"
              text_anchor: "middle"

    - type: image
      href: "https://example.com/icon.png"
      x: 650
      y: 50
      width: 100
      height: 100
```

**Description**

Constructs an 800x600 pixel SVG with a background image. The main group is centered and contains a yellow rectangle, a title text above it, and a nested group rotated by -30 degrees. The nested group includes a horizontal line and a label. Additionally, an icon image is placed at the top-right corner.

**Corresponding SVG Output**

```svg
<svg width="800" height="600" xmlns="http://www.w3.org/2000/svg">
  <image href="https://example.com/background.jpg" width="100%" height="100%" />
  
  <g transform="translate(400,300)">
    <rect x="-150" y="-100" width="300" height="200" fill="#ffeb3b" stroke="#fbc02d" stroke-width="5" />
    <text x="0" y="-120" font-size="32px" font-family="Helvetica" fill="#000000" text-anchor="middle">Complex SVG Scene</text>
    
    <g transform="rotate(-30)">
      <line x1="-100" y1="0" x2="100" y2="0" stroke="#00bcd4" stroke-width="3" />
      <text x="0" y="20" font-size="16px" font-family="Courier New" fill="#e91e63" text-anchor="middle">Nested Group</text>
    </g>
  </g>

  <image href="https://example.com/icon.png" x="650" y="50" width="100" height="100" />
</svg>
```

---

## Conclusion

The YAML-based DSL for SVG provides a user-friendly and maintainable way to create and manage SVG graphics. By abstracting the verbose SVG XML syntax, this DSL allows developers and designers to define complex graphics with ease. The inclusion of text elements further broadens its applicability, enabling the addition of labels, annotations, and dynamic text content within SVGs. With the addition of triangle, ellipse, and polygon as drawable objects, the DSL now offers even more flexibility and versatility in creating a wide range of SVG graphics.