package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// UIRenderer handles the visual rendering of the template config interface
type UIRenderer struct {
	Width  int
	Height int
}

// NewUIRenderer creates a new UI renderer
func NewUIRenderer() *UIRenderer {
	return &UIRenderer{}
}

// SetSize updates the renderer dimensions
func (u *UIRenderer) SetSize(width, height int) {
	u.Width = width
	u.Height = height
}

// RenderView renders the complete template config view
func (u *UIRenderer) RenderView(template *TemplateDefinition, formHandler *FormHandler, inputHandler *InputHandler, preview, toastMessage string, toastExpiry time.Time) string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(u.Width)

	title := titleStyle.Render(template.Label)
	b.WriteString(title)
	b.WriteString("\n\n")

	// Form content area
	contentHeight := u.Height - 6 // Reserve space for title and status
	leftWidth := u.Width / 2
	rightWidth := u.Width - leftWidth

	// Form items (left side)
	formContent := u.renderFormItems(template, formHandler, inputHandler, leftWidth, contentHeight)

	// Preview (right side)
	previewContent := u.renderPreview(preview, rightWidth, contentHeight)

	// Use lipgloss to join horizontally
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, formContent, previewContent)
	b.WriteString(mainContent)
	b.WriteString("\n")

	// Status bar
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#333333")).
		Padding(0, 1).
		Width(u.Width)

	status := "↑↓/j/k: Navigate  Tab: Next  Space/Enter: Toggle  c: Copy  s: Save  ←/Esc: Back"
	if toastMessage != "" && time.Now().Before(toastExpiry) {
		status = toastMessage
	}

	b.WriteString(statusStyle.Render(status))

	return b.String()
}

// renderFormItems renders the form section
func (u *UIRenderer) renderFormItems(template *TemplateDefinition, formHandler *FormHandler, inputHandler *InputHandler, width, height int) string {
	var b strings.Builder

	// Create a container with fixed width
	containerStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 1)

	b.WriteString("Variables:\n")

	for i, item := range formHandler.Items {
		if item.Type == "variable" {
			cursor := "  "
			if i == formHandler.FocusIndex {
				cursor = "► "
			}

			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
			if i == formHandler.FocusIndex {
				labelStyle = labelStyle.Bold(true).Foreground(lipgloss.Color("#7D56F4"))
			}

			b.WriteString(cursor)
			b.WriteString(labelStyle.Render(item.Label))
			b.WriteString("\n")

			// Value box
			value := item.Value
			if i == formHandler.FocusIndex {
				value = inputHandler.GetDisplayValue(item.Value)
			}
			if value == "" {
				value = item.Hint
			}

			boxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(0, 1).
				Width(width - 8) // Account for padding and cursor

			if i == formHandler.FocusIndex {
				boxStyle = boxStyle.BorderForeground(lipgloss.Color("#7D56F4"))
			}

			b.WriteString(boxStyle.Render(value))
			b.WriteString("\n\n")
		}
	}

	b.WriteString("Sections:\n\n")

	for i, item := range formHandler.Items {
		if item.Type == "section" || item.Type == "bullet" || item.Type == "toggle" || item.Type == "bullet_header" {
			cursor := "  "
			if i == formHandler.FocusIndex {
				cursor = "► "
			}

			if item.Type == "section" {
				// Section header with prominent styling
				sectionHeaderStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#7D56F4")).
					Background(lipgloss.Color("#2E2E2E")).
					Padding(0, 1).
					Width(width - 8)

				b.WriteString(cursor)
				b.WriteString(sectionHeaderStyle.Render(item.Label))
				b.WriteString("\n")

				for _, opt := range item.Options {
					marker := "○"
					optValue := opt
					if opt == item.Value {
						marker = "●"
					}

					// Show description if available for the variant
					for _, section := range template.Sections {
						if section.ID == item.Key {
							for _, variant := range section.Variants {
								variantDisplayName := variant.Label
								if variantDisplayName == "" {
									variantDisplayName = variant.ID
								}
								if opt == variantDisplayName && variant.Description != "" {
									optValue = fmt.Sprintf("%s - %s", opt, variant.Description)
								}
							}
							break
						}
					}

					b.WriteString(fmt.Sprintf("  %s %s\n", marker, optValue))
				}
				b.WriteString("\n")
			} else if item.Type == "bullet_header" {
				// Bullet section header
				headerStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#FAFAFA")).
					Background(lipgloss.Color("#444444")).
					Padding(0, 1).
					Width(width - 8)

				b.WriteString("\n")
				b.WriteString(headerStyle.Render(item.Label))
				b.WriteString("\n\n")
			} else if item.Type == "bullet" || item.Type == "toggle" {
				marker := "☐"
				if item.Selected {
					marker = "☑"
				}

				style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
				if i == formHandler.FocusIndex {
					style = style.Bold(true).Foreground(lipgloss.Color("#7D56F4"))
				}

				labelText := item.Label
				if item.Type == "toggle" && item.Hint != "" {
					labelText += fmt.Sprintf(" (%s)", item.Hint)
				}

				line := fmt.Sprintf("%s%s %s", cursor, marker, labelText)
				b.WriteString(style.Render(line))
				b.WriteString("\n")
			}
		}
	}

	return containerStyle.Render(b.String())
}

// renderPreview renders the preview section
func (u *UIRenderer) renderPreview(preview string, width, height int) string {
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
	content := previewStyle.Render(preview)

	previewContent := title + "\n" + content
	return containerStyle.Render(previewContent)
}
