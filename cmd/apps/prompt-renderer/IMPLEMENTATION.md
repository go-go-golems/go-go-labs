# Prompt Renderer Implementation Guide

This document provides a comprehensive guide to understanding and developing the prompt-renderer application. It's designed for new developers (including junior developers and interns) who need to understand the codebase architecture and contribute to the project.

## Table of Contents

1. [Overview & Architecture](#overview--architecture)
2. [Project Structure](#project-structure)
3. [Core Components](#core-components)
4. [DSL (Domain-Specific Language)](#dsl-domain-specific-language)
5. [TUI (Terminal User Interface)](#tui-terminal-user-interface)
6. [Rendering Engine](#rendering-engine)
7. [State Management](#state-management)
8. [Testing Strategy](#testing-strategy)
9. [Development Workflow](#development-workflow)
10. [Adding New Features](#adding-new-features)

## Overview & Architecture

The prompt-renderer is a terminal-based application that converts YAML prompt templates into clipboard-ready prompts through an interactive configuration interface. Think of it as a form builder for AI prompts - users select a template, fill in variables, choose options, and get a fully rendered prompt.

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   DSL Parser    │    │   TUI Engine    │    │   Renderer      │
│  (parser.go)    │────│ (BubbleTea)     │────│ (renderer.go)   │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Persistence    │    │   Clipboard     │    │   Validation    │
│ (persistence.go)│    │ (clipboard.go)  │    │ (parser.go)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Data Flow

1. **Load Phase**: DSL parser reads `templates.yml` and validates structure
2. **Selection Phase**: TUI presents template list, user selects one
3. **Configuration Phase**: TUI presents form for variables, sections, and bullets
4. **Rendering Phase**: Renderer assembles final prompt from template + user inputs
5. **Output Phase**: Clipboard manager copies result, state is persisted

## Project Structure

```
prompt-renderer/
├── main.go                    # CLI entry point, Cobra commands
├── types.go                   # Core data structures
├── parser.go                  # DSL parsing and validation
├── renderer.go                # Template rendering engine
├── template_list.go           # Template selection TUI
├── template_config.go         # Template configuration TUI
├── clipboard.go               # Cross-platform clipboard operations
├── persistence.go             # State saving/loading
├── templates.yml              # Example DSL file
├── test_*.go                 # Test files
├── demo/                     # VHS demonstration scripts
└── TODO.md                   # Development roadmap
```

### File Responsibilities

- **`main.go`**: Command-line interface using Cobra, application bootstrap
- **`types.go`**: All data structures (DSL, UI state, selections)
- **`parser.go`**: YAML parsing, DSL validation, error handling
- **`renderer.go`**: Template rendering, variable substitution, bullet processing
- **`template_list.go`**: BubbleTea model for template selection screen
- **`template_config.go`**: BubbleTea model for template configuration screen
- **`clipboard.go`**: Cross-platform clipboard integration
- **`persistence.go`**: Auto-save, history, state restoration

## Core Components

### 1. DSL Parser (`parser.go`)

The parser is responsible for loading and validating YAML template definitions. It ensures the DSL structure is correct before the TUI starts.

#### Key Functions:

```go
// Main parsing entry point
func ParseDSLFile(path string) (*DSLFile, error)

// Validates entire DSL structure
func validateDSLFile(dsl *DSLFile) error

// Validates individual templates
func validateTemplate(template *TemplateDefinition, index int) error
```

#### Validation Rules:

- Templates must have unique IDs
- Sections must have at least one variant  
- Variants must have valid types: `text`, `bullets`, or `toggle`
- Required fields must be present (ID, label, content for text/toggle types)

### 2. TUI Engine (BubbleTea Models)

The application uses the BubbleTea framework for terminal UI. There are two main screens:

#### Template List (`template_list.go`)

Displays available templates in a selectable list format.

```go
type TemplateListModel struct {
    templates []TemplateDefinition
    cursor    int
    width     int
    height    int
}
```

**Key Interactions:**
- `↑↓` or `j/k`: Navigate templates
- `Enter`: Select template
- `1-9`: Direct template selection
- `q` or `Ctrl+C`: Quit

#### Template Configuration (`template_config.go`)

The main configuration interface where users set variables and choose options.

```go
type TemplateConfigModel struct {
    template     *TemplateDefinition
    selection    *SelectionState
    renderer     *PromptRenderer
    focusIndex   int
    formItems    []FormItem  // Flattened form representation
    preview      string      // Live rendered preview
    // ... other fields
}
```

**Form Item Types:**
- **Variable**: Text input fields for template variables
- **Section**: Variant selectors (if multiple variants exist)
- **Toggle**: On/off switches for toggle-type variants
- **Bullet**: Individual bullet checkboxes

#### Form Building Process

The `rebuildFormItems()` function converts the hierarchical template structure into a flat list of interactive form items:

```pseudocode
for each variable in template:
    add FormItem{Type: "variable", Key: varName, ...}

for each section in template:
    if section has multiple variants:
        add FormItem{Type: "section", ...}
    
    for current selected variant:
        if variant.Type == "toggle":
            add FormItem{Type: "toggle", ...}
        elif variant.Type == "bullets":
            for each bullet:
                add FormItem{Type: "bullet", ...}
```

### 3. Rendering Engine (`renderer.go`)

The renderer assembles final prompts from templates and user selections.

#### Key Functions:

```go
// Main rendering entry point
func (r *PromptRenderer) RenderPrompt(templateDef *TemplateDefinition, selection *SelectionState) (string, error)

// Processes individual sections
func (r *PromptRenderer) renderVariant(variant *VariantDefinition, sectionSelection SectionSelection, ...) (string, error)

// Handles bullet-type variants
func (r *PromptRenderer) renderBullets(variant *VariantDefinition, sectionSelection SectionSelection) string
```

#### Rendering Process:

1. **Section Processing**: Process each section in template order
2. **Variant Selection**: Use selected variant for each section
3. **Type-Specific Rendering**:
   - `text`: Return content directly
   - `toggle`: Return content if enabled, empty string if disabled
   - `bullets`: Process selected bullets with prefix
4. **Variable Substitution**: Replace `{{ .variable }}` with actual values
5. **Cleanup**: Normalize whitespace and line endings

#### Variable Substitution

Uses Go's `text/template` package with Sprig functions for powerful templating:

```go
tmpl, err := template.New("prompt").Funcs(sprig.TxtFuncMap()).Parse(content)
// Execute with variables map
```

## DSL (Domain-Specific Language)

The DSL defines prompt templates in YAML format. Understanding the structure is crucial for both parsing and rendering.

### Core Structure

```yaml
version: 1
globals:
  model_fallback: "claude-3-sonnet"
  bullet_prefix: "- "

templates:
  - id: example-template
    label: "Example Template"
    variables:
      var_name:
        hint: "Help text for user"
        type: "text"
    sections:
      - id: section-id
        label: "Section Label"
        variants:
          - id: variant-id
            label: "Variant Label"
            description: "What this variant does"
            type: "text|bullets|toggle"
            content: "Template content with {{ .var_name }}"
            bullets: ["Bullet 1", "Bullet 2"]  # For bullets type
```

### DSL Types in Code

```go
// Root structure
type DSLFile struct {
    Version   int                    `yaml:"version"`
    Globals   *GlobalConfig          `yaml:"globals,omitempty"`
    Templates []TemplateDefinition   `yaml:"templates"`
}

// Individual template
type TemplateDefinition struct {
    ID        string                    `yaml:"id"`
    Label     string                    `yaml:"label"`
    Variables map[string]VariableConfig `yaml:"variables,omitempty"`
    Sections  []SectionDefinition       `yaml:"sections"`
}

// Section with variants
type SectionDefinition struct {
    ID       string               `yaml:"id"`
    Label    string               `yaml:"label,omitempty"`
    Variants []VariantDefinition  `yaml:"variants"`
}

// Individual variant
type VariantDefinition struct {
    ID          string   `yaml:"id"`
    Label       string   `yaml:"label,omitempty"`
    Description string   `yaml:"description,omitempty"`
    Type        string   `yaml:"type"` // "text", "bullets", "toggle"
    Content     string   `yaml:"content,omitempty"`
    Bullets     []string `yaml:"bullets,omitempty"`
}
```

### Variant Types Explained

1. **Text Variants**: Static content with variable substitution
   ```yaml
   type: "text"
   content: "Please review this {{ .language }} code:"
   ```

2. **Toggle Variants**: Optional content that can be enabled/disabled
   ```yaml
   type: "toggle"
   content: "Please provide context before reviewing."
   ```

3. **Bullets Variants**: Selectable bullet points
   ```yaml
   type: "bullets"
   bullets:
     - "Code quality and readability"
     - "Performance considerations"
   ```

## TUI (Terminal User Interface)

The TUI uses the BubbleTea framework, which follows the Elm Architecture pattern.

### BubbleTea Pattern

Each screen is a "Model" that implements three methods:

```go
type tea.Model interface {
    Init() tea.Cmd                               // Initialize
    Update(tea.Msg) (tea.Model, tea.Cmd)       // Handle events
    View() string                               // Render UI
}
```

### Message Passing

The application uses custom messages for communication between components:

```go
// User selected a template
type SelectTemplateMsg struct {
    Template TemplateDefinition
}

// User wants to copy prompt
type CopyPromptMsg struct {
    Prompt string
}

// Copy operation completed
type CopyDoneMsg struct{}
```

### Navigation and Input Handling

The `template_config.go` handles complex navigation between form elements:

```go
func (m *TemplateConfigModel) handleNormalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "up", "k":
        if m.focusIndex > 0 {
            m.focusIndex--
        }
    case "down", "j":
        if m.focusIndex < len(m.formItems)-1 {
            m.focusIndex++
        }
    case "space":
        // Toggle bullets/toggles
        item := m.formItems[m.focusIndex]
        if item.Type == "bullet" || item.Type == "toggle" {
            m.handleToggle()
        }
    case "enter":
        // Edit variables or cycle sections
        // ...
    }
}
```

### Layout with Lipgloss

The UI uses Lipgloss for styling and layout. The configuration screen uses horizontal layout:

```go
// Form content (left side)
formContent := m.renderFormItems(leftWidth, contentHeight)

// Preview (right side) 
previewContent := m.renderPreview(rightWidth, contentHeight)

// Join horizontally
mainContent := lipgloss.JoinHorizontal(lipgloss.Top, formContent, previewContent)
```

## Rendering Engine

The rendering engine is the heart of the application, converting templates and user selections into final prompts.

### Default Selection Logic

When a template is loaded, smart defaults are applied:

```go
func CreateDefaultSelection(templateDef *TemplateDefinition) *SelectionState {
    selection := &SelectionState{
        TemplateID: templateDef.ID,
        Variables:  make(map[string]string),
        Sections:   make(map[string]SectionSelection),
    }

    // Initialize variables with empty values
    for varName := range templateDef.Variables {
        selection.Variables[varName] = ""
    }

    // Initialize sections with first variant and default bullets
    for _, section := range templateDef.Sections {
        if len(section.Variants) > 0 {
            sectionSelection := SectionSelection{
                Variant:         section.Variants[0].ID,
                SelectedBullets: make(map[string]bool),
                VariantEnabled:  false, // Toggles default to off
            }

            // Auto-select first 3 bullets for better UX
            firstVariant := section.Variants[0]
            if firstVariant.Type == "bullets" {
                maxSelections := min(3, len(firstVariant.Bullets))
                for i := 0; i < maxSelections; i++ {
                    bulletKey := fmt.Sprintf("%d", i)
                    sectionSelection.SelectedBullets[bulletKey] = true
                }
            }

            selection.Sections[section.ID] = sectionSelection
        }
    }

    return selection
}
```

### Variable Substitution Process

Variables use Go template syntax with Sprig functions for enhanced functionality:

```go
// Prepare variables map
variables := make(map[string]string)
for varName, value := range selection.Variables {
    // Handle file references
    if strings.HasPrefix(value, "@") {
        filename := strings.TrimPrefix(value, "@")
        fileContent, err := r.readFileContent(filename)
        if err != nil {
            variables[varName] = fmt.Sprintf("ERROR_READING_FILE: %s", err.Error())
        } else {
            variables[varName] = fileContent
        }
    } else {
        variables[varName] = value
    }
}

// Execute template
tmpl, err := template.New("prompt").Funcs(sprig.TxtFuncMap()).Parse(content)
err = tmpl.Execute(&buf, variables)
```

### Bullet Rendering

Bullets can be rendered in two ways:

1. **Simple Mode**: Just render selected bullets with prefix
2. **Template Mode**: Use content template with `{{.}}` substitution

```go
func (r *PromptRenderer) renderBullets(variant *VariantDefinition, sectionSelection SectionSelection) string {
    if variant.Content != "" {
        // Template mode: collect selected bullets and substitute into content
        var selectedBullets []string
        for i, bullet := range variant.Bullets {
            bulletKey := fmt.Sprintf("%d", i)
            if sectionSelection.SelectedBullets[bulletKey] {
                selectedBullets = append(selectedBullets, bulletPrefix+bullet)
            }
        }
        return strings.ReplaceAll(variant.Content, "{{.}}", strings.Join(selectedBullets, "\n"))
    } else {
        // Simple mode: just render selected bullets
        // ...
    }
}
```

## State Management

The application manages several types of state:

### Selection State

Current user choices are stored in `SelectionState`:

```go
type SelectionState struct {
    TemplateID      string                       `yaml:"template_id"`
    Timestamp       time.Time                    `yaml:"timestamp"`
    Variables       map[string]string            `yaml:"variables,omitempty"`
    Sections        map[string]SectionSelection  `yaml:"sections,omitempty"`
}

type SectionSelection struct {
    Variant         string            `yaml:"variant"`
    SelectedBullets map[string]bool   `yaml:"selected_bullets,omitempty"`
    VariantEnabled  bool              `yaml:"variant_enabled,omitempty"`
}
```

### Persistence Strategy

The persistence manager handles auto-save and history:

```
$XDG_DATA_HOME/prompt-builder/
├── templates.yml              # User's DSL file
├── last.yml                   # Current working state (auto-save)
├── history/                   # Manual saves with timestamps
│   ├── 20250621-143207_draft.yml
│   └── 20250621-151134_review.yml
└── config.yml                 # Application settings
```

### Auto-Save Implementation

State is automatically saved on significant changes with debouncing:

```go
// In template_config.go, after any selection change:
m.updatePreview()  // Regenerate preview
// Auto-save is triggered by the main app loop with debouncing
```

The persistence manager provides these key functions:

```go
func (p *PersistenceManager) SaveCurrentState(selection *SelectionState) error
func (p *PersistenceManager) LoadCurrentState() (*SelectionState, error)
func (p *PersistenceManager) SaveToHistory(selection *SelectionState) error
```

## Testing Strategy

The application uses multiple testing approaches:

### 1. Unit Tests

Core functionality tests in `test_renderer.go`:

```go
func TestRenderer() {
    // Load DSL
    dslFile, err := ParseDSLFile("templates.yml")
    
    // Create test selection
    selection := CreateDefaultSelection(template)
    selection.Variables["code_snippet"] = "test code"
    
    // Test rendering
    prompt, err := renderer.RenderPrompt(template, selection)
    
    // Verify output
}
```

### 2. Integration Tests

P0 fixes are tested in `test_p0_fixes.go`:

```go
func TestP0Fixes() {
    // Test toggle functionality
    // Test bullet selection
    // Test bounds checking
    // Test error handling
}
```

### 3. UI Tests

UI improvements are verified in `test_ui_improvements.go`:

```go
func TestUIImprovements() {
    // Test default selections
    // Test label display
    // Test styling consistency
}
```

### 4. Visual Tests (VHS)

VHS scripts provide visual verification:

- `verify_p0_fixes.tape`: Demonstrates bug fixes
- `verify_ui_improvements.tape`: Shows UI enhancements
- Demo scripts in `demo/` directory

### Running Tests

```bash
# Run core functionality test
go run . test

# Run P0 fixes verification
go run . test-p0

# Run UI improvements test  
go run . test-ui

# Run visual tests
vhs verify_p0_fixes.tape
```

## Development Workflow

### Setting Up Development Environment

1. **Clone and Build**:
   ```bash
   git clone <repository>
   cd prompt-renderer
   go build .
   ```

2. **Install VHS for visual testing**:
   ```bash
   # Install VHS for recording terminal sessions
   go install github.com/charmbracelet/vhs@latest
   ```

3. **Run tests to verify setup**:
   ```bash
   go run . test
   go run . test-p0
   go run . test-ui
   ```

### Code Style and Conventions

The project follows Go best practices:

- **Error Handling**: Use `github.com/pkg/errors` for wrapping errors with context
- **Logging**: Use `zerolog` for structured logging
- **CLI**: Use Cobra for command structure
- **TUI**: Follow BubbleTea patterns and Elm Architecture
- **Styling**: Use Lipgloss for consistent visual styling

### Adding Dependencies

Only add dependencies that are necessary. Current key dependencies:

```go
// Core TUI framework
github.com/charmbracelet/bubbletea
github.com/charmbracelet/lipgloss

// CLI framework
github.com/spf13/cobra

// Template processing
github.com/Masterminds/sprig/v3

// Utilities
github.com/pkg/errors
github.com/rs/zerolog
github.com/atotto/clipboard
gopkg.in/yaml.v3
```

### Debugging Tips

1. **Enable Debug Logging**:
   ```bash
   go run . --log-level debug
   ```

2. **Test Specific Components**:
   ```bash
   # Test only rendering
   go run . test
   
   # Test only UI
   go run . test-ui
   ```

3. **Visual Debugging with VHS**:
   Create simple VHS scripts to reproduce UI issues:
   ```vhs
   Type "go run ."
   Enter
   Sleep 2s
   # Reproduce the issue steps
   Screenshot debug.png
   ```

## Adding New Features

### 1. Adding a New Variant Type

To add a new variant type (e.g., "dropdown"):

1. **Update Types** (`types.go`):
   ```go
   // Add new fields to VariantDefinition if needed
   type VariantDefinition struct {
       // ... existing fields
       Options []string `yaml:"options,omitempty"` // For dropdown
   }
   ```

2. **Update Parser** (`parser.go`):
   ```go
   func validateVariant(variant *VariantDefinition, ...) error {
       if variant.Type != "text" && variant.Type != "bullets" && 
          variant.Type != "toggle" && variant.Type != "dropdown" {
           return fmt.Errorf("invalid type: %s", variant.Type)
       }
       
       if variant.Type == "dropdown" && len(variant.Options) == 0 {
           return fmt.Errorf("dropdown requires options")
       }
   }
   ```

3. **Update Renderer** (`renderer.go`):
   ```go
   func (r *PromptRenderer) renderVariant(...) (string, error) {
       switch variant.Type {
       case "dropdown":
           // Implement dropdown rendering logic
           return r.renderDropdown(variant, sectionSelection), nil
       }
   }
   ```

4. **Update TUI** (`template_config.go`):
   ```go
   // In rebuildFormItems()
   case "dropdown":
       // Add FormItem for dropdown selection
   
   // In input handling
   case "enter":
       if item.Type == "dropdown" {
           m.openDropdownSelection(item)
       }
   ```

### 2. Adding New DSL Features

To add template-level features:

1. **Extend DSL Structure**: Add fields to appropriate types in `types.go`
2. **Update Validation**: Add validation rules in `parser.go`
3. **Update Processing**: Modify rendering logic in `renderer.go`
4. **Update UI**: Add UI elements in TUI models
5. **Add Tests**: Create tests for the new functionality
6. **Update Documentation**: Add examples to `templates.yml`

### 3. Adding New UI Features

For UI-only features:

1. **Plan the UX**: Sketch the interaction flow
2. **Update Models**: Modify TUI models to handle new states
3. **Add Messages**: Create new message types for communication
4. **Implement Handlers**: Add input handling and state updates
5. **Style with Lipgloss**: Ensure consistent visual styling
6. **Test with VHS**: Create visual verification scripts

### 4. Example: Adding a "Help Screen"

This is a good first feature for new developers:

1. **Create Help Model** (new file `help.go`):
   ```go
   type HelpModel struct {
       width  int
       height int
   }
   
   func (m HelpModel) View() string {
       helpText := `
   Prompt Renderer Help
   
   Navigation:
     ↑↓, j/k    Navigate up/down
     Tab        Next field
     Shift+Tab  Previous field
     Enter      Edit/Select
     Space      Toggle
     c          Copy to clipboard
     s          Save state
     Esc        Back/Cancel
     q          Quit
   `
       return lipgloss.NewStyle().Padding(1).Render(helpText)
   }
   ```

2. **Add Help Message**:
   ```go
   type ShowHelpMsg struct{}
   type HideHelpMsg struct{}
   ```

3. **Update Main App Model**:
   ```go
   type AppModel struct {
       // ... existing fields
       helpModel *HelpModel
       showingHelp bool
   }
   ```

4. **Handle Help in Update**:
   ```go
   case tea.KeyMsg:
       if msg.String() == "?" {
           m.showingHelp = true
           return m, nil
       }
   ```

5. **Add to View Logic**:
   ```go
   func (m *AppModel) View() string {
       if m.showingHelp {
           return m.helpModel.View()
       }
       // ... normal view logic
   }
   ```
