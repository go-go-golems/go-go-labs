package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TemplateConfigModel handles template configuration screen
type TemplateConfigModel struct {
	template     *TemplateDefinition
	selection    *SelectionState
	renderer     *PromptRenderer
	focusIndex   int
	focusedField string
	focusedKey   string
	editing      bool
	editingValue string
	preview      string
	width        int
	height       int
	toastMessage string
	toastExpiry  time.Time
	formItems    []FormItem
}

// FormItem represents a configurable item in the form
type FormItem struct {
	Type        string // "variable", "section", "bullet"
	Key         string // variable name, section ID, or group ID
	SectionID   string // for bullet items
	VariantID   string // for bullet items
	Label       string
	Value       string
	Hint        string
	Options     []string // for sections
	Selected    bool     // for bullets
}

// NewTemplateConfigModel creates a new template config model
func NewTemplateConfigModel(template *TemplateDefinition, renderer *PromptRenderer) *TemplateConfigModel {
	selection := CreateDefaultSelection(template)
	selection.Timestamp = time.Now()
	
	m := &TemplateConfigModel{
		template:  template,
		selection: selection,
		renderer:  renderer,
		focusIndex: 0,
	}
	
	m.rebuildFormItems()
	m.updatePreview()
	
	return m
}

// rebuildFormItems constructs the form items list
func (m *TemplateConfigModel) rebuildFormItems() {
	m.formItems = []FormItem{}

	// Add variables
	varNames := make([]string, 0, len(m.template.Variables))
	for name := range m.template.Variables {
		varNames = append(varNames, name)
	}
	sort.Strings(varNames) // Sort for consistent ordering

	for _, name := range varNames {
		varConfig := m.template.Variables[name]
		value := ""
		if val, exists := m.selection.Variables[name]; exists {
			value = val
		}
		
		m.formItems = append(m.formItems, FormItem{
			Type:  "variable",
			Key:   name,
			Label: name,
			Value: value,
			Hint:  varConfig.Hint,
		})
	}

	// Add sections and their bullet groups
	for _, section := range m.template.Sections {
		sectionSelection := m.selection.Sections[section.ID]
		
		// Add section variant selector
		options := make([]string, len(section.Variants))
		for i, variant := range section.Variants {
			options[i] = variant.ID
		}
		
		m.formItems = append(m.formItems, FormItem{
			Type:    "section",
			Key:     section.ID,
			Label:   section.ID,
			Value:   sectionSelection.Variant,
			Options: options,
		})

		// Add bullet groups for current variant
		for _, variant := range section.Variants {
			if variant.ID == sectionSelection.Variant && variant.Type == "bullets" {
				selectedGroups := make(map[string]bool)
				for _, groupID := range sectionSelection.Groups {
					selectedGroups[groupID] = true
				}

				for _, group := range variant.Groups {
					m.formItems = append(m.formItems, FormItem{
						Type:      "bullet",
						Key:       group.ID,
						SectionID: section.ID,
						VariantID: variant.ID,
						Label:     group.ID,
						Selected:  selectedGroups[group.ID],
					})
				}
			}
		}
	}
}

// Init implements tea.Model
func (m *TemplateConfigModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *TemplateConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editing {
			return m.handleEditingInput(msg)
		}
		return m.handleNormalInput(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case CopyDoneMsg:
		m.showToast("✓ Copied to clipboard!")
		return m, tea.Tick(750*time.Millisecond, func(time.Time) tea.Msg {
			return ClearToastMsg{}
		})
	case ClearToastMsg:
		m.toastMessage = ""
	}

	return m, nil
}

// handleNormalInput handles input when not in editing mode
func (m *TemplateConfigModel) handleNormalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "left":
		return m, func() tea.Msg {
			return GoBackMsg{}
		}
	case "up", "k":
		if m.focusIndex > 0 {
			m.focusIndex--
		}
	case "down", "j":
		if m.focusIndex < len(m.formItems)-1 {
			m.focusIndex++
		}
	case "tab":
		m.focusIndex = (m.focusIndex + 1) % len(m.formItems)
	case "enter":
		if m.focusIndex < len(m.formItems) {
			item := m.formItems[m.focusIndex]
			if item.Type == "variable" {
				m.startEditing(item.Value)
			} else if item.Type == "section" {
				m.cycleSectionVariant(item.Key)
			}
		}
	case "space":
		if m.focusIndex < len(m.formItems) {
			item := m.formItems[m.focusIndex]
			if item.Type == "bullet" {
				m.toggleBulletGroup(item.SectionID, item.Key)
			}
		}
	case "c":
		return m, func() tea.Msg {
			return CopyPromptMsg{Prompt: m.preview}
		}
	case "s":
		return m, func() tea.Msg {
			return SaveSelectionMsg{Selection: m.selection}
		}
	case "ctrl+r":
		m.updatePreview()
	case "?":
		return m, func() tea.Msg {
			return ShowHelpMsg{}
		}
	}

	return m, nil
}

// handleEditingInput handles input when in editing mode
func (m *TemplateConfigModel) handleEditingInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editing = false
		m.editingValue = ""
	case "enter":
		m.commitEdit()
		m.editing = false
	case "backspace":
		if len(m.editingValue) > 0 {
			m.editingValue = m.editingValue[:len(m.editingValue)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.editingValue += msg.String()
		}
	}

	return m, nil
}

// startEditing begins editing a variable
func (m *TemplateConfigModel) startEditing(currentValue string) {
	m.editing = true
	m.editingValue = currentValue
}

// commitEdit saves the edited value
func (m *TemplateConfigModel) commitEdit() {
	if m.focusIndex < len(m.formItems) {
		item := m.formItems[m.focusIndex]
		if item.Type == "variable" {
			m.selection.Variables[item.Key] = m.editingValue
			m.formItems[m.focusIndex].Value = m.editingValue
			m.updatePreview()
		}
	}
	m.editingValue = ""
}

// cycleSectionVariant cycles through section variants
func (m *TemplateConfigModel) cycleSectionVariant(sectionID string) {
	sectionSelection := m.selection.Sections[sectionID]
	
	// Find current variant index
	var currentIndex int
	for _, section := range m.template.Sections {
		if section.ID == sectionID {
			for i, variant := range section.Variants {
				if variant.ID == sectionSelection.Variant {
					currentIndex = i
					break
				}
			}
			
			// Cycle to next variant
			nextIndex := (currentIndex + 1) % len(section.Variants)
			sectionSelection.Variant = section.Variants[nextIndex].ID
			sectionSelection.Groups = []string{} // Reset bullet selections
			m.selection.Sections[sectionID] = sectionSelection
			break
		}
	}
	
	m.rebuildFormItems()
	m.updatePreview()
}

// toggleBulletGroup toggles a bullet group selection
func (m *TemplateConfigModel) toggleBulletGroup(sectionID, groupID string) {
	sectionSelection := m.selection.Sections[sectionID]
	
	// Check if group is currently selected
	groupSelected := false
	newGroups := []string{}
	for _, g := range sectionSelection.Groups {
		if g == groupID {
			groupSelected = true
		} else {
			newGroups = append(newGroups, g)
		}
	}
	
	// Toggle selection
	if !groupSelected {
		newGroups = append(newGroups, groupID)
	}
	
	sectionSelection.Groups = newGroups
	m.selection.Sections[sectionID] = sectionSelection
	
	m.rebuildFormItems()
	m.updatePreview()
}

// updatePreview regenerates the prompt preview
func (m *TemplateConfigModel) updatePreview() {
	preview, err := m.renderer.RenderPrompt(m.template, m.selection)
	if err != nil {
		m.preview = fmt.Sprintf("Error rendering prompt: %v", err)
	} else {
		m.preview = preview
	}
}

// showToast displays a temporary message
func (m *TemplateConfigModel) showToast(message string) {
	m.toastMessage = message
	m.toastExpiry = time.Now().Add(750 * time.Millisecond)
}

// View implements tea.Model
func (m *TemplateConfigModel) View() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)

	title := titleStyle.Render(m.template.Label)
	b.WriteString(title)
	b.WriteString("\n\n")

	// Form content area
	contentHeight := m.height - 6 // Reserve space for title and status
	leftWidth := m.width / 2
	rightWidth := m.width - leftWidth

	// Form items (left side)
	formContent := m.renderFormItems(leftWidth, contentHeight)
	
	// Preview (right side)
	previewContent := m.renderPreview(rightWidth, contentHeight)

	// Use lipgloss to join horizontally
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, formContent, previewContent)
	b.WriteString(mainContent)
	b.WriteString("\n")

	// Status bar
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#333333")).
		Padding(0, 1).
		Width(m.width)

	status := "↑↓: Navigate  Enter: Edit  Space: Toggle  c: Copy  s: Save  ←/Esc: Back"
	if m.toastMessage != "" && time.Now().Before(m.toastExpiry) {
		status = m.toastMessage
	}
	
	b.WriteString(statusStyle.Render(status))

	return b.String()
}

// renderFormItems renders the form section
func (m *TemplateConfigModel) renderFormItems(width, height int) string {
	var b strings.Builder

	// Create a container with fixed width
	containerStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 1)

	b.WriteString("Variables:\n")
	
	for i, item := range m.formItems {
		if item.Type == "variable" {
			cursor := "  "
			if i == m.focusIndex {
				cursor = "► "
			}

			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
			if i == m.focusIndex {
				labelStyle = labelStyle.Bold(true).Foreground(lipgloss.Color("#7D56F4"))
			}

			b.WriteString(cursor)
			b.WriteString(labelStyle.Render(item.Label))
			b.WriteString("\n")

			// Value box
			value := item.Value
			if m.editing && i == m.focusIndex {
				value = m.editingValue + "█" // Show cursor
			}
			if value == "" {
				value = item.Hint
			}

			boxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(0, 1).
				Width(width - 8) // Account for padding and cursor

			if i == m.focusIndex {
				boxStyle = boxStyle.BorderForeground(lipgloss.Color("#7D56F4"))
			}

			b.WriteString("  ")
			b.WriteString(boxStyle.Render(value))
			b.WriteString("\n\n")
		}
	}

	b.WriteString("Sections:\n\n")
	
	for i, item := range m.formItems {
		if item.Type == "section" || item.Type == "bullet" {
			cursor := "  "
			if i == m.focusIndex {
				cursor = "► "
			}

			if item.Type == "section" {
				b.WriteString(cursor)
				b.WriteString(item.Label)
				b.WriteString("\n")
				
				for _, opt := range item.Options {
					marker := "○"
					if opt == item.Value {
						marker = "●"
					}
					b.WriteString(fmt.Sprintf("  %s %s\n", marker, opt))
				}
				b.WriteString("\n")
			} else if item.Type == "bullet" {
				marker := "☐"
				if item.Selected {
					marker = "☑"
				}
				
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
				if i == m.focusIndex {
					style = style.Bold(true).Foreground(lipgloss.Color("#7D56F4"))
				}
				
				line := fmt.Sprintf("%s%s %s", cursor, marker, item.Label)
				b.WriteString(style.Render(line))
				b.WriteString("\n")
			}
		}
	}

	return containerStyle.Render(b.String())
}

// renderPreview renders the preview section
func (m *TemplateConfigModel) renderPreview(width, height int) string {
	// Create a container with fixed width
	containerStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	previewStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(width - 6). // Account for container padding
		Height(height - 4)

	title := titleStyle.Render("Preview:")
	content := previewStyle.Render(m.preview)

	previewContent := title + "\n" + content
	return containerStyle.Render(previewContent)
}

// Message types
type CopyPromptMsg struct {
	Prompt string
}

type CopyDoneMsg struct{}

type ClearToastMsg struct{}

type SaveSelectionMsg struct {
	Selection *SelectionState
}

type GoBackMsg struct{}
