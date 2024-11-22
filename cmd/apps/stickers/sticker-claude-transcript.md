# Film Roll Sticker Customization DSL

Project: Unnamed Project

Created: 2024-11-17T17:42:44.160632+00:00
Updated: 2024-11-17T17:48:38.186934+00:00

## Messages

### Message 1

Sender: human
Created: 2024-11-17T17:43:04.510119+00:00

create a yaml DSL to represent little stickers I want to put on my developed rolls of film tracking things like start date, end date, film type, push pull, maybe notes abut conditions when shooting, titel, camera, lens.

The DSL can be used to create different personalized stickers based on the information that I might want to add on that day.

### Message 2

Sender: assistant
Created: 2024-11-17T17:43:04.510119+00:00



<antArtifact identifier="film-sticker-dsl" type="application/vnd.ant.code" language="yaml" title="Film Roll Sticker DSL">
# Film Sticker DSL Schema
---
# Sticker template definition
template:
  size: small  # small, medium, large
  style: minimal  # minimal, detailed, artistic
  orientation: landscape  # landscape, portrait

# Required film information
film_info:
title: "Summer Memories"  # A name for this roll
camera: "Nikon FM2"
lens: "50mm f/1.8"
film:
brand: "Kodak"
type: "Portra 400"
format: "35mm"  # 35mm, 120, 4x5, etc.
iso: 400

# Processing information
processing:
push_pull: 0  # Number of stops (can be negative)
development:
process: "C-41"  # C-41, E-6, BW
lab: "Downtown Photo"
special_instructions: "Standard process"

# Timeline
dates:
started: 2024-03-15
completed: 2024-04-01
developed: 2024-04-03

# Optional shooting conditions
conditions:
weather:
- sunny
- overcast
temperature: 22C
location: "City Park"
notes: "Shot during golden hour"

# Optional technical notes
technical_notes:
filters_used: ["Polarizer"]
meter_used: "Internal"
exposure_compensation: "-0.3"

---
# Example Usage 1: Minimal Street Photography Sticker
template:
size: small
style: minimal
orientation: landscape

film_info:
title: "Street Walk"
camera: "Leica M6"
lens: "35mm f/2"
film:
brand: "Ilford"
type: "HP5+"
format: "35mm"
iso: 400

processing:
push_pull: 1
development:
process: "BW"

dates:
started: 2024-03-20
completed: 2024-03-21
developed: 2024-03-22

---
# Example Usage 2: Detailed Landscape Photography Sticker
template:
size: large
style: detailed
orientation: landscape

film_info:
title: "Mountain Sunrise"
camera: "Mamiya RB67"
lens: "90mm f/3.8"
film:
brand: "Fujifilm"
type: "Velvia 50"
format: "120"
iso: 50

processing:
push_pull: 0
development:
process: "E-6"
lab: "Pro Lab"
special_instructions: "Handle with care"

dates:
started: 2024-04-01
completed: 2024-04-01
developed: 2024-04-03

conditions:
weather:
- clear
- cold
temperature: -5C
location: "Mount Baker"
notes: "Shot at sunrise, clear conditions"

technical_notes:
filters_used:
- "Graduated ND"
- "Polarizer"
meter_used: "Spot meter"
exposure_compensation: 0
</antArtifact>

I've created a YAML-based DSL for your film roll stickers. The DSL is structured to be both flexible and comprehensive, allowing you to create stickers with varying levels of detail. Here are the key features:

1. Template section: Define the physical properties of the sticker
2. Required film information: Camera, lens, and film details
3. Processing information: Development specifications
4. Timeline: Track when the roll was started and completed
5. Optional sections for:
    - Shooting conditions
    - Technical notes
    - Weather information
    - Location details

The DSL includes two example usages:
1. A minimal sticker for street photography
2. A detailed sticker for landscape photography

You can use as much or as little of the schema as needed for each sticker. Would you like me to modify any part of the schema or add additional fields?

### Message 3

Sender: human
Created: 2024-11-17T17:43:33.030373+00:00

no, the DSL should describe the empty sticker template that i can then fill in by hand using apen while on the road shooting.

### Message 4

Sender: assistant
Created: 2024-11-17T17:43:33.030373+00:00

Ah, I understand now - you want a YAML DSL that describes the layout of a physical sticker template that you can print and then fill in by hand while shooting. Let me revise that.



<antArtifact identifier="film-sticker-template-dsl" type="application/vnd.ant.code" language="yaml" title="Handwritten Film Sticker Template DSL">
# Film Sticker Template DSL
---
template:
  name: "Basic Film Roll"
  size: 
    width: "3in"
    height: "2in"
  orientation: landscape
  paper_type: "weatherproof"  # weatherproof, standard

fields:
- type: text_field
  label: "Title"
  size: large
  position: top

- type: line
  label: "Camera"
  length: medium

- type: line
  label: "Lens"
  length: medium

- type: box_group
  label: "Film"
  fields:
    - type: line
      label: "Type"
      length: short
    - type: checkbox
      label: "Format"
      options: ["35mm", "120", "4x5"]

- type: box_group
  label: "Development"
  fields:
    - type: line
      label: "Push/Pull"
      length: short
    - type: checkbox
      label: "Process"
      options: ["C41", "BW", "E6"]

- type: date_fields
  labels: ["Started", "Finished"]
  format: "MM/DD/YY"

- type: notes
  label: "Notes"
  lines: 2
  size: small

---
# Example alternative template: Minimal
template:
name: "Minimal"
size:
width: "2in"
height: "1in"
orientation: landscape
paper_type: "standard"

fields:
- type: line
  label: "Title"
  length: medium

- type: box_group
  label: "Film"
  fields:
    - type: line
      label: "Type"
      length: short
    - type: checkbox
      label: "Format"
      options: ["35mm", "120"]

- type: date_fields
  labels: ["S:", "F:"]
  format: "MM/DD"
  </antArtifact>

I've created a DSL that describes physical sticker templates you can print and fill in by hand. Key features:

1. Template properties:
    - Physical size and orientation
    - Paper type specification
    - Overall layout structure

2. Field types:
    - Text fields: Basic lines to write on
    - Checkboxes: For quick format/process selection
    - Date fields: Formatted spaces for dates
    - Notes: Multi-line areas for additional information

3. Field attributes:
    - Labels: What goes next to each field
    - Sizes: How much space to allocate
    - Positioning: Where on the sticker it goes

The DSL includes two templates:
1. A comprehensive template with all common fields
2. A minimal template for quick notation

This way you can print these templates and fill them in with a pen while shooting. Would you like me to adjust any of the field types or layouts?

### Message 5

Sender: human
Created: 2024-11-17T17:44:46.030034+00:00

make a list of the widgets i might want to print on a sticker.

### Message 6

Sender: assistant
Created: 2024-11-17T17:44:46.030034+00:00



<antArtifact identifier="film-sticker-widgets" type="application/vnd.ant.code" language="yaml" title="Film Sticker Widget Catalog">
# Film Sticker Widget Catalog
---
basic_widgets:
  text_line:
    description: "Single line for writing text"
    variations:
      - short  # For numbers, codes
      - medium # For camera names, film types
      - long   # For titles, locations

checkbox:
description: "Single checkbox with label"
example_uses:
- "[ ] Push processed"
- "[ ] Contains flash photos"

checkbox_group:
description: "Group of related checkboxes"
example_uses:
- format: ["[ ] 35mm", "[ ] 120", "[ ] 4x5"]
- process: ["[ ] C41", "[ ] BW", "[ ] E6"]

circles:
description: "Small circles to fill/check"
example_uses:
- push_pull: "○ -2 ○ -1 ○ 0 ○ +1 ○ +2"
- weather: "○ ☀️ ○ ☁️ ○ 🌧️"

date_field:
description: "Formatted boxes for dates"
variations:
- "DD/MM"
- "MM/YY"
- "DD/MM/YY"

notes_area:
description: "Multi-line area for notes"
variations:
- small  # 2-3 lines
- medium # 4-5 lines

decorative_widgets:
dividers:
description: "Visual separators"
types:
- solid_line
- dotted_line
- double_line

borders:
description: "Edge decorations"
types:
- simple_frame
- rounded_corners
- tick_marks

icons:
description: "Simple pictograms"
examples:
- "📷 Camera"
- "🎞️ Film"
- "⚡ Flash"
- "🌡️ Temperature"
- "⚙️ Settings"

layout_widgets:
sections:
description: "Grouped information areas"
types:
- boxed
- underlined
- bracketed

columns:
description: "Vertical divisions"
variations:
- two_column
- three_column

tabs:
description: "Side or top tabs"
uses:
- roll_number
- date_ranges
- quick_reference

measurement_widgets:
temperature_scale:
description: "Temperature marking area"
format: "[-10°C ──┼──┼──┼── 40°C]"

exposure_scale:
description: "Exposure compensation marks"
format: "[-2 ──┼──┼──┼── +2]"

light_meter:
description: "Simple exposure value scale"
format: "[EV: 1-2-3-4-5-6-7-8-9]"

utility_widgets:
qr_placeholder:
description: "Space for QR code"
uses:
- link_to_digital_scan
- roll_identifier

color_chips:
description: "Color reference squares"
uses:
- white_balance
- color_matching

grid:
description: "Dot or line grid"
uses:
- exposure_log
- frame_tracker
</antArtifact>

I've organized the widgets into categories:

1. Basic Widgets:
    - Essential text fields and checkboxes
    - Date fields and notes areas
    - Simple input mechanisms

2. Decorative Widgets:
    - Visual elements to organize information
    - Icons and borders
    - Aesthetic elements that also serve functional purposes

3. Layout Widgets:
    - Ways to organize information spatially
    - Section dividers and grouping elements
    - Structural components

4. Measurement Widgets:
    - Specialized scales and marks
    - Technical reference elements
    - Visual measurement tools

5. Utility Widgets:
    - QR codes and tracking elements
    - Color references
    - Grid systems for logging

Each widget includes variations and example uses to show how it might be applied in practice. Would you like me to:
1. Add more specific widgets for particular use cases?
2. Provide more detailed specifications for any of these widgets?
3. Show how these widgets might be combined in example layouts?

### Message 7

Sender: human
Created: 2024-11-17T17:46:26.265079+00:00

simplify, i like text line, multiline, date fields, checkbox group (with an option to do circles), and standard text (paragrpah, or title).

For layout, allow goruping vertical or horizontally

### Message 8

Sender: assistant
Created: 2024-11-17T17:46:26.265079+00:00

Let me create a simplified widget system focusing on those core elements.

<antArtifact identifier="simplified-sticker-widgets" type="application/vnd.ant.code" language="yaml" title="Simplified Film Sticker Widgets">
# Simple Film Sticker Widget System
---
widgets:
  # Basic text input
  text_line:
    description: "Single line for writing"
    sizes:
      - short  # 15-20 chars (ISO, serial numbers)
      - medium # 30-40 chars (camera names, lenses)
      - long   # 50-60 chars (titles, locations)
    attributes:
      label: string  # Label that appears before the line
      underline: boolean  # Whether to use solid line or dots

text_block:
description: "Multiple lines for longer text"
sizes:
- small  # 2-3 lines
- medium # 4-5 lines
attributes:
label: string
lines: number
line_style: "solid | dotted"

date_field:
description: "Formatted date input"
formats:
- "MM/DD"    # When year isn't needed
- "MM/DD/YY" # Full date
attributes:
label: string
format: string

choice_group:
description: "Group of options to mark"
styles:
- checkbox: "[ ]"  # Square boxes
- circle: "○"      # Round circles
attributes:
label: string
options: string[]
style: "checkbox | circle"

heading:
description: "Section title or label"
sizes:
- large   # Main titles
- medium  # Section headers
- small   # Field labels
attributes:
text: string
weight: "bold | normal"

# Layout containers
layouts:
horizontal_group:
description: "Arrange widgets in a row"
attributes:
spacing: "tight | normal | loose"
alignment: "top | center | bottom"
widgets: widget[]

vertical_group:
description: "Stack widgets in a column"
attributes:
spacing: "tight | normal | loose"
alignment: "left | center | right"
widgets: widget[]

# Example usage
---
example_sticker:
- type: vertical_group
  alignment: left
  spacing: normal
  widgets:
    - type: heading
      text: "Film Roll Details"
      size: large

    - type: horizontal_group
      spacing: normal
      widgets:
        - type: text_line
          label: "Camera"
          size: medium
        - type: text_line
          label: "Lens"
          size: medium

    - type: choice_group
      label: "Format"
      style: circle
      options: ["35mm", "120", "4x5"]

    - type: horizontal_group
      spacing: tight
      widgets:
        - type: date_field
          label: "Started"
          format: "MM/DD"
        - type: date_field
          label: "Finished"
          format: "MM/DD"

    - type: text_block
      label: "Notes"
      size: small
      lines: 3
      </antArtifact>

I've simplified the system down to five core widgets:
1. `text_line`: For single-line inputs
2. `text_block`: For multi-line text areas
3. `date_field`: For formatted dates
4. `choice_group`: For checkboxes/circles
5. `heading`: For labels and titles

And two layout containers:
1. `horizontal_group`: Arranges things in a row
2. `vertical_group`: Stacks things in a column

Each widget has minimal but useful attributes, and the example shows how they can be composed together. Would you like me to:
1. Adjust any of the widget sizes or attributes?
2. Show more example compositions?
3. Add or remove any features?

### Message 9

Sender: human
Created: 2024-11-17T17:47:17.747116+00:00

Implement an app to render such a sticker.

### Message 10

Sender: assistant
Created: 2024-11-17T17:47:17.747116+00:00

I'll create a React app that renders a film roll sticker based on the YAML configuration.



<antArtifact identifier="film-sticker-renderer" type="application/vnd.ant.react" title="Film Roll Sticker Renderer">
import React from 'react';
import { Card } from '@/components/ui/card';

// Widget Components
const TextLine = ({ label, size = "medium" }) => {
const widthClass = {
short: "w-20",
medium: "w-40",
long: "w-64"
}[size];

return (
<div className="flex items-center gap-2">
{label && <span className="text-sm font-medium">{label}</span>}
<div className={`${widthClass} h-6 border-b border-gray-300`} />
</div>
);
};

const TextBlock = ({ label, size = "small" }) => {
const lines = size === "small" ? 3 : 5;
return (
<div className="flex flex-col gap-1">
{label && <span className="text-sm font-medium">{label}</span>}
<div className="flex flex-col gap-4">
{Array(lines).fill(0).map((_, i) => (
<div key={i} className="w-full h-0.5 border-b border-gray-300" />
))}
</div>
</div>
);
};

const DateField = ({ label, format = "MM/DD" }) => {
const segments = format.split('/');

return (
<div className="flex items-center gap-2">
{label && <span className="text-sm font-medium">{label}</span>}
<div className="flex items-center gap-1">
{segments.map((seg, i) => (
<React.Fragment key={i}>
<div className="w-8 h-6 border-b border-gray-300" />
{i < segments.length - 1 && <span>/</span>}
</React.Fragment>
))}
</div>
</div>
);
};

const ChoiceGroup = ({ label, options = [], style = "checkbox" }) => {
const marker = style === "checkbox" ? "□" : "○";

return (
<div className="flex flex-col gap-1">
{label && <span className="text-sm font-medium">{label}</span>}
<div className="flex gap-4">
{options.map((option, i) => (
<div key={i} className="flex items-center gap-1">
<span className="text-lg">{marker}</span>
<span className="text-sm">{option}</span>
</div>
))}
</div>
</div>
);
};

const Heading = ({ text, size = "medium" }) => {
const sizeClass = {
large: "text-lg font-bold",
medium: "text-base font-semibold",
small: "text-sm font-medium"
}[size];

return <h3 className={sizeClass}>{text}</h3>;
};

// Layout Components
const HorizontalGroup = ({ children, spacing = "normal", alignment = "center" }) => {
const spacingClass = {
tight: "gap-2",
normal: "gap-4",
loose: "gap-6"
}[spacing];

const alignClass = {
top: "items-start",
center: "items-center",
bottom: "items-end"
}[alignment];

return (
<div className={`flex ${spacingClass} ${alignClass}`}>
{children}
</div>
);
};

const VerticalGroup = ({ children, spacing = "normal", alignment = "left" }) => {
const spacingClass = {
tight: "gap-2",
normal: "gap-4",
loose: "gap-6"
}[spacing];

const alignClass = {
left: "items-start",
center: "items-center",
right: "items-end"
}[alignment];

return (
<div className={`flex flex-col ${spacingClass} ${alignClass}`}>
{children}
</div>
);
};

// Example Sticker Component
const FilmSticker = () => {
return (
<Card className="p-6 w-96 bg-white">
<VerticalGroup spacing="normal">
<Heading text="Film Roll Details" size="large" />

        <HorizontalGroup spacing="normal">
          <TextLine label="Camera" size="medium" />
          <TextLine label="Lens" size="medium" />
        </HorizontalGroup>
        
        <ChoiceGroup 
          label="Format" 
          style="circle"
          options={["35mm", "120", "4x5"]} 
        />
        
        <HorizontalGroup spacing="tight">
          <DateField label="Started" format="MM/DD" />
          <DateField label="Finished" format="MM/DD" />
        </HorizontalGroup>
        
        <TextBlock label="Notes" size="small" />
      </VerticalGroup>
    </Card>
);
};

export default FilmSticker;
</antArtifact>

I've created a React application that renders film roll stickers. The implementation includes:

1. Core Widget Components:
    - `TextLine`: For single-line inputs with different widths
    - `TextBlock`: For multi-line notes areas
    - `DateField`: For formatted date inputs
    - `ChoiceGroup`: For checkbox/circle option groups
    - `Heading`: For section titles and labels

2. Layout Components:
    - `HorizontalGroup`: Flex row container with spacing options
    - `VerticalGroup`: Flex column container with spacing options

3. A complete example sticker implementation showing how to compose the widgets

The sticker is rendered in a card with a white background and appropriate spacing. All components use Tailwind CSS for styling.

Would you like me to:
1. Add more widget variations?
2. Create additional example layouts?
3. Add printing functionality?
4. Make the sticker configurable through a YAML input?

### Message 11

Sender: human
Created: 2024-11-17T17:48:38.186934+00:00

Make it so that it can take an arbitrary JSON / JS data structure representing the sticker, based on our dsl definition. Then, create 4 different stickers. One very detailed. One just camera and lens, one push pull with dates, etc...

### Message 12

Sender: assistant
Created: 2024-11-17T17:48:38.186934+00:00

I'll modify the component to be data-driven and create several different sticker templates.

<antArtifact identifier="configurable-film-sticker" type="application/vnd.ant.react" title="Configurable Film Roll Sticker">
import React from 'react';
import { Card } from '@/components/ui/card';

// Widget Components (same as before, just showing usage)
const TextLine = ({ label, size = "medium" }) => {
const widthClass = {
short: "w-20",
medium: "w-40",
long: "w-64"
}[size];

return (
<div className="flex items-center gap-2">
{label && <span className="text-sm font-medium">{label}</span>}
<div className={`${widthClass} h-6 border-b border-gray-300`} />
</div>
);
};

const TextBlock = ({ label, size = "small" }) => {
const lines = size === "small" ? 3 : 5;
return (
<div className="flex flex-col gap-1">
{label && <span className="text-sm font-medium">{label}</span>}
<div className="flex flex-col gap-4">
{Array(lines).fill(0).map((_, i) => (
<div key={i} className="w-full h-0.5 border-b border-gray-300" />
))}
</div>
</div>
);
};

const DateField = ({ label, format = "MM/DD" }) => {
const segments = format.split('/');

return (
<div className="flex items-center gap-2">
{label && <span className="text-sm font-medium">{label}</span>}
<div className="flex items-center gap-1">
{segments.map((seg, i) => (
<React.Fragment key={i}>
<div className="w-8 h-6 border-b border-gray-300" />
{i < segments.length - 1 && <span>/</span>}
</React.Fragment>
))}
</div>
</div>
);
};

const ChoiceGroup = ({ label, options = [], style = "checkbox" }) => {
const marker = style === "checkbox" ? "□" : "○";

return (
<div className="flex flex-col gap-1">
{label && <span className="text-sm font-medium">{label}</span>}
<div className="flex gap-4">
{options.map((option, i) => (
<div key={i} className="flex items-center gap-1">
<span className="text-lg">{marker}</span>
<span className="text-sm">{option}</span>
</div>
))}
</div>
</div>
);
};

const Heading = ({ text, size = "medium" }) => {
const sizeClass = {
large: "text-lg font-bold",
medium: "text-base font-semibold",
small: "text-sm font-medium"
}[size];

return <h3 className={sizeClass}>{text}</h3>;
};

// Layout Components
const Group = ({ type = "vertical", children, spacing = "normal", alignment = "left" }) => {
const spacingClass = {
tight: "gap-2",
normal: "gap-4",
loose: "gap-6"
}[spacing];

const alignClass = type === "vertical" ? {
left: "items-start",
center: "items-center",
right: "items-end"
}[alignment] : {
top: "items-start",
center: "items-center",
bottom: "items-end"
}[alignment];

const directionClass = type === "vertical" ? "flex-col" : "flex-row";

return (
<div className={`flex ${directionClass} ${spacingClass} ${alignClass}`}>
{children}
</div>
);
};

// Widget renderer
const renderWidget = (widget) => {
switch (widget.type) {
case "text_line":
return <TextLine key={widget.label} {...widget} />;
case "text_block":
return <TextBlock key={widget.label} {...widget} />;
case "date_field":
return <DateField key={widget.label} {...widget} />;
case "choice_group":
return <ChoiceGroup key={widget.label} {...widget} />;
case "heading":
return <Heading key={widget.text} {...widget} />;
case "horizontal_group":
return (
<Group key={widget.label} type="horizontal" spacing={widget.spacing} alignment={widget.alignment}>
{widget.widgets.map(renderWidget)}
</Group>
);
case "vertical_group":
return (
<Group key={widget.label} type="vertical" spacing={widget.spacing} alignment={widget.alignment}>
{widget.widgets.map(renderWidget)}
</Group>
);
default:
return null;
}
};

// Example sticker configurations
const stickerConfigs = {
detailed: {
type: "vertical_group",
spacing: "normal",
widgets: [
{
type: "heading",
text: "Professional Film Log",
size: "large"
},
{
type: "horizontal_group",
spacing: "normal",
widgets: [
{ type: "text_line", label: "Camera", size: "medium" },
{ type: "text_line", label: "Lens", size: "medium" }
]
},
{
type: "horizontal_group",
spacing: "normal",
widgets: [
{
type: "choice_group",
label: "Format",
style: "circle",
options: ["35mm", "120", "4x5"]
},
{
type: "choice_group",
label: "Process",
style: "circle",
options: ["C41", "B&W", "E6"]
}
]
},
{
type: "horizontal_group",
spacing: "normal",
widgets: [
{ type: "text_line", label: "Film", size: "medium" },
{ type: "text_line", label: "ISO", size: "short" }
]
},
{
type: "horizontal_group",
spacing: "tight",
widgets: [
{ type: "date_field", label: "Started", format: "MM/DD/YY" },
{ type: "date_field", label: "Finished", format: "MM/DD/YY" }
]
},
{
type: "choice_group",
label: "Push/Pull",
style: "circle",
options: ["-2", "-1", "0", "+1", "+2"]
},
{ type: "text_block", label: "Notes", size: "medium" }
]
},

minimal: {
type: "vertical_group",
spacing: "tight",
widgets: [
{
type: "horizontal_group",
spacing: "normal",
widgets: [
{ type: "text_line", label: "Camera", size: "medium" },
{ type: "text_line", label: "Lens", size: "short" }
]
}
]
},

pushPull: {
type: "vertical_group",
spacing: "normal",
widgets: [
{
type: "horizontal_group",
spacing: "tight",
widgets: [
{ type: "date_field", label: "Start", format: "MM/DD" },
{ type: "date_field", label: "End", format: "MM/DD" }
]
},
{
type: "choice_group",
label: "Push/Pull",
style: "circle",
options: ["-2", "-1", "0", "+1", "+2"]
},
{ type: "text_line", label: "Process Notes", size: "long" }
]
},

compact: {
type: "vertical_group",
spacing: "tight",
widgets: [
{
type: "horizontal_group",
spacing: "tight",
widgets: [
{ type: "text_line", label: "Film", size: "medium" },
{ type: "text_line", label: "ISO", size: "short" }
]
},
{
type: "choice_group",
label: "Process",
style: "circle",
options: ["C41", "B&W", "E6"]
},
{ type: "date_field", label: "Date", format: "MM/DD" }
]
}
};

// Configurable Sticker Component
const FilmSticker = ({ config }) => {
return (
<Card className="p-6 bg-white">
{renderWidget(config)}
</Card>
);
};

// Demo component showing all sticker variants
const StickerDemo = () => {
return (
<div className="flex flex-wrap gap-6 p-6">
<div className="w-96">
<h2 className="mb-4 text-lg font-bold">Detailed Sticker</h2>
<FilmSticker config={stickerConfigs.detailed} />
</div>

      <div className="w-64">
        <h2 className="mb-4 text-lg font-bold">Minimal Sticker</h2>
        <FilmSticker config={stickerConfigs.minimal} />
      </div>
      
      <div className="w-72">
        <h2 className="mb-4 text-lg font-bold">Push/Pull Sticker</h2>
        <FilmSticker config={stickerConfigs.pushPull} />
      </div>
      
      <div className="w-72">
        <h2 className="mb-4 text-lg font-bold">Compact Sticker</h2>
        <FilmSticker config={stickerConfigs.compact} />
      </div>
    </div>
);
};

export default StickerDemo;
</antArtifact>

I've created a fully configurable sticker system with four example layouts:

1. Detailed Sticker:
    - Complete professional log with all fields
    - Camera, lens, format, process type
    - Dates, push/pull settings, and notes
    - Suitable for professional workflows

2. Minimal Sticker:
    - Just camera and lens fields
    - Perfect for quick identification

3. Push/Pull Sticker:
    - Focused on development details
    - Dates and push/pull settings
    - Process notes
    - Ideal for lab instructions

4. Compact Sticker:
    - Film type and ISO
    - Process type selection
    - Single date
    - Good for basic film tracking

The system:
- Takes a JSON configuration that matches our DSL
- Renders any combination of widgets
- Supports nested layouts (horizontal/vertical groups)
- Is fully customizable through the config structure

You can create new sticker layouts by defining new configurations following the same pattern. Would you like me to:
1. Add more example configurations?
2. Add additional widget types?
3. Add printing functionality?
4. Add a configuration builder interface?

