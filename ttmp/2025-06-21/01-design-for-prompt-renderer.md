# TUI Prompt-Builder Implementation Specification

## 1. Overview & Architecture

### 1.1 Purpose
A terminal-based application for converting YAML prompt templates into clipboard-ready prompts through an interactive configuration interface. The application uses a two-screen workflow: template selection followed by template configuration with live preview.

### 1.2 Core Components
- **DSL Parser**: Validates and loads YAML template definitions
- **Template Renderer**: Assembles final prompts from templates + user selections
- **State Manager**: Handles persistence of user selections and history
- **TUI Engine**: Manages screen rendering and user input (built with BubbleTea)
- **Clipboard Interface**: Cross-platform clipboard integration

### 1.3 Data Flow
```
YAML DSL → Parser → Template Store → TUI → User Selections → Renderer → Clipboard
                                    ↓
                               State Manager → Persistence Files
```

---

## 2. Domain-Specific Language (DSL) Specification

### 2.1 File Structure
```yaml
version: 1                    # DSL version for compatibility checking
globals:                      # Optional global defaults
  model_fallback: string     # Default model if template doesn't specify
  bullet_prefix: string      # Default bullet character (e.g., "- ")
  
templates:                    # Array of template definitions
  - id: string               # Unique identifier (slug format)
    label: string            # Human-readable display name
    model: string            # Optional model override
    variables:               # Map of variable definitions
      var_name:
        hint: string         # Help text for user
        type: "text"         # Currently only "text" supported
    sections:                # Ordered array of prompt sections
      - id: string           # Unique section identifier
        variants:            # Array of variant options
          - id: string       # Variant identifier
            type: "text" | "bullets"
            content: string  # For type="text": literal content
            groups:          # For type="bullets": bullet group definitions
              - id: string   # Group identifier
                bullets:     # Array of bullet point strings
                  - string
```

### 2.2 Variable Interpolation
- Variables use go template syntax: `{{ variable_name }}` (use the helpers in glazed/pkg/helpers/templating to get sprig templates)
- Variables can appear in any `content` field
- Unresolved variables are set to a DEFAULT_NAME_OF_VARIABLE value to avoid template rendering errors

### 2.3 Validation Rules
- All `id` fields must be unique within their scope
- Templates must have at least one section
- Sections must have at least one variant
- First variant in each section is the default selection
- Bullet groups can be empty (all bullets optional)

---

## 3. Application States & Navigation

### 3.1 Application State Machine
```
Start → Template List → Template Config → [Copy Success] → Template Config
  ↓           ↓              ↓                              ↑
Quit ←      Quit ←         Back ←─────────────────────────┘
```

### 3.2 Template List State
**Purpose**: Display available templates and allow selection

**Data Required**:
- Array of templates from parsed DSL
- Current selection index
- Template descriptions/labels

**User Actions**:
- Navigate up/down through template list
- Select template (Enter) → transition to Template Config
- Quit application (q/Esc)

### 3.3 Template Config State
**Purpose**: Configure selected template and generate prompt

**Data Required**:
- Selected template definition
- Current variable values (map: string → string)
- Current section variant selections (map: section_id → variant_id)
- Current bullet group selections (map: section_id → set of group_ids)
- Rendered prompt preview string

**User Actions**:
- Navigate between form elements
- Edit variable values
- Change section variants
- Toggle bullet group selections
- Copy prompt to clipboard (c)
- Save current selections (s)
- Return to template list (Esc/←)

---

## 4. User Interface Specification

### 4.1 Template List Screen Layout

```
┌─ PROMPT BUILDER ─────────────────────────────────────────────────────────────┐
│                                                                              │
│                              Select a Template                               │
│                                                                              │
│    1. <template.label>                                                       │
│       <template description/hint>                                           │
│                                                                              │
│ ►  2. <template.label>                                     [SELECTED]        │
│       <template description/hint>                                           │
│                                                                              │
│    N. <template.label>                                                       │
│       <template description/hint>                                           │
│                                                                              │
│                                                                              │
├─────────────────────────────────────────────────────────────────────────────┤
│ ↑↓/j/k: Navigate    Enter: Select    q: Quit                                │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Visual Elements**:
- Centered title "Select a Template"
- Numbered list of templates with labels and descriptions
- Selection indicator `►` for current item
- Clear visual separation between templates
- Status bar with key bindings

**Behavior**:
- Templates numbered 1-N for quick selection
- Up/down arrow keys or j/k for vim-style navigation
- Enter key selects current template
- Number keys 1-9 for direct selection (if ≤9 templates)
- q or Esc quits application

### 4.2 Template Configuration Screen Layout

```
┌─ <template.label> ───────────────────────────────────────────────────────────┐
│                                                                              │
│ Variables:                                                                   │
│   <variable_name>                                                            │
│   ┌────────────────────────────────────────────────────────────────────────┐ │
│   │ <current_value or placeholder_hint>                                    │ │
│   │                                                                        │ │
│   └────────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
│ Sections:                                                                    │
│                                                                              │
│   <section_id>                                                               │
│   ● <selected_variant>                                                       │
│   ○ <other_variant>                                                          │
│     Bullet Groups:                                                           │
│     ☑ <group_name>        ☐ <group_name>                                    │
│                                                                              │
│ Preview:                                                                     │
│ ┌────────────────────────────────────────────────────────────────────────┐ │
│ │ <rendered_prompt_content>                                              │ │
│ │ {{ unresolved_variables }}                                             │ │
│ │ <bullet_points_based_on_selections>                                    │ │
│ └────────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
├─────────────────────────────────────────────────────────────────────────────┤
│ ↑↓: Navigate  Enter: Edit  Space: Toggle  c: Copy  s: Save  ←/Esc: Back     │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Visual Elements**:
- Template name in title bar
- Variables section with text input boxes
- Sections with radio button variants and checkbox bullet groups
- Live preview pane showing assembled prompt
- Comprehensive status bar with all available actions

---

## 5. UI Component Specifications

### 5.1 Variable Input Component
**Appearance**: Text box with label and hint
**Behavior**:
- Single-line or multi-line based on content length
- Enter key opens full-screen editor for multi-line content
- File picker support for `@filename` syntax
- Real-time preview updates as user types

### 5.2 Section Variant Selector
**Appearance**: Radio button list
**Behavior**:
- Only one variant selectable per section
- Enter key opens dropdown/menu for variant selection
- First variant is default selection
- Selection changes trigger immediate preview update

### 5.3 Bullet Group Toggle
**Appearance**: Checkbox list under applicable variants
**Behavior**:
- Space bar toggles selection
- Multiple groups can be selected simultaneously
- Empty selection is valid (no bullets from that variant)
- Visual grouping under parent variant

### 5.4 Preview Pane
**Appearance**: Scrollable text box with rendered content
**Behavior**:
- Updates in real-time as selections change
- Shows `{{ variable_name }}` for unresolved variables
- Renders only selected bullet groups
- Scrollable for long prompts
- Read-only display

---

## 6. User Interaction Patterns

### 6.1 Navigation Flow
1. **Application Start**: Load DSL, validate, show template list
2. **Template Selection**: User browses and selects template
3. **Configuration**: User fills variables and configures sections
4. **Preview & Copy**: User reviews assembled prompt and copies to clipboard
5. **Iteration**: User can modify selections or return to template list

### 6.2 Keyboard Shortcuts
**Global**:
- `q` / `Esc`: Quit or go back
- `?`: Show help overlay

**Template List**:
- `↑↓` / `j/k`: Navigate templates
- `Enter`: Select template
- `1-9`: Direct template selection

**Template Config**:
- `↑↓` / `j/k`: Navigate form elements
- `Tab`: Next form element
- `Enter`: Edit/configure current element
- `Space`: Toggle boolean elements (checkboxes)
- `c`: Copy prompt to clipboard
- `s`: Save current selections
- `Ctrl+R`: Force preview refresh

### 6.3 Edit Modes
**Variable Editing**:
- Inline editing for short text
- Full-screen editor for long text (triggered by Enter)
- File browser for `@filename` references
- Escape cancels, Enter confirms

**Variant Selection**:
- Dropdown menu or inline selection
- Arrow keys navigate options
- Enter confirms selection
- Escape cancels

---

## 7. Data Persistence & State Management

### 7.1 File Structure
```
$XDG_DATA_HOME/prompt-builder/
├── templates.yml              # User's DSL file
├── last.yml                   # Current working state
├── history/
│   ├── 20250621-143207_draft.yml
│   ├── 20250621-151134_summarize.yml
│   └── ...
└── config.yml                 # Application settings
```

### 7.2 Selection State Format
```yaml
template_id: string
timestamp: string              # ISO 8601 format
variables:
  var_name: string
sections:
  section_id:
    variant: string
    groups: [string]           # Array of selected group IDs
```

### 7.3 Auto-save Behavior
- Save to `last.yml` on every significant change
- Create timestamped history file on manual save (s key)
- Restore from `last.yml` on application restart
- Prompt for unsaved changes on quit

---

## 8. Prompt Rendering Engine

### 8.1 Assembly Algorithm
1. **Initialize**: Start with empty prompt string
2. **Process Sections**: For each section in template order:
   - Get selected variant for section
   - If variant type is "text": append content directly
   - If variant type is "bullets": process bullet groups
3. **Process Bullets**: For bullet-type variants:
   - For each selected group, append all bullets with prefix
   - Use global bullet_prefix or section-specific override
4. **Variable Substitution**: Replace all `{{ var_name }}` with values
5. **Cleanup**: Trim excessive whitespace, normalize line endings

### 8.2 Variable Resolution
- Simple string substitution using mustache syntax
- File references (`@filename`) resolved to file contents
- Unresolved variables left as-is for user visibility
- Error handling for missing files or invalid references

### 8.3 Bullet Formatting
- Each bullet gets configured prefix (default: "- ")
- Bullets maintain original indentation relative to prefix
- Empty groups contribute no content
- Groups processed in definition order

---

## 9. Error Handling & Validation

### 9.1 DSL Validation Errors
**Parse Errors**:
- Invalid YAML syntax → show line/column, exit gracefully
- Missing required fields → show specific field path
- Duplicate IDs → show conflicting entries
- Unsupported version → show version mismatch warning

**Runtime Errors**:
- Variable substitution failures → highlight in preview
- File access errors → show file path and error
- Clipboard access failures → fall back to stdout

### 9.2 User Input Validation
**Variable Validation**:
- No specific validation (free text input)
- File references validated on access
- Large content warnings for performance

**Selection Validation**:
- Invalid variant IDs prevented by UI constraints
- Invalid bullet group IDs prevented by UI constraints
- State consistency maintained automatically

### 9.3 Recovery Mechanisms
- Auto-save prevents data loss
- Graceful degradation for clipboard failures
- Partial prompt generation with warnings for errors
- State restoration from last.yml on crash recovery

---

## 10. Platform Integration
### 10.1 Clipboard Support
**Primary Methods**:
- Use `github.com/atotto/clipboard` package for cross-platform support
- Handles Linux (xclip, wl-copy, xsel), macOS (pbcopy), Windows automatically
- Single API call: `clipboard.WriteAll(content)`

