package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
)

// PromptRenderer handles the assembly of prompts from templates and selections
type PromptRenderer struct {
	dslFile *DSLFile
}

// NewPromptRenderer creates a new prompt renderer
func NewPromptRenderer(dslFile *DSLFile) *PromptRenderer {
	return &PromptRenderer{
		dslFile: dslFile,
	}
}

// RenderPrompt assembles a final prompt from a template and user selections
func (r *PromptRenderer) RenderPrompt(templateDef *TemplateDefinition, selection *SelectionState) (string, error) {
	var prompt strings.Builder

	// Process each section in order
	for _, section := range templateDef.Sections {
		sectionSelection, hasSelection := selection.Sections[section.ID]

		// Default to first variant if no selection
		selectedVariantID := section.Variants[0].ID
		if hasSelection && sectionSelection.Variant != "" {
			selectedVariantID = sectionSelection.Variant
		}

		// Find the selected variant
		var selectedVariant *VariantDefinition
		for _, variant := range section.Variants {
			if variant.ID == selectedVariantID {
				selectedVariant = &variant
				break
			}
		}

		if selectedVariant == nil {
			return "", fmt.Errorf("variant '%s' not found in section '%s'", selectedVariantID, section.ID)
		}

		// Process the variant
		sectionContent, err := r.renderVariant(selectedVariant, sectionSelection, templateDef, selection)
		if err != nil {
			return "", errors.Wrapf(err, "failed to render variant '%s' in section '%s'", selectedVariantID, section.ID)
		}

		if sectionContent != "" {
			prompt.WriteString(sectionContent)
			prompt.WriteString("\n\n")
		}
	}

	// Perform variable substitution
	renderedPrompt, err := r.substituteVariables(prompt.String(), templateDef, selection)
	if err != nil {
		return "", errors.Wrap(err, "failed to substitute variables")
	}

	// Clean up excessive whitespace
	return r.cleanupPrompt(renderedPrompt), nil
}

// renderVariant processes a single variant
func (r *PromptRenderer) renderVariant(variant *VariantDefinition, sectionSelection SectionSelection, templateDef *TemplateDefinition, selection *SelectionState) (string, error) {
	switch variant.Type {
	case "text":
		return variant.Content, nil
	case "toggle":
		if sectionSelection.VariantEnabled {
			return variant.Content, nil
		}
		return "", nil
	case "bullets":
		return r.renderBullets(variant, sectionSelection), nil
	default:
		return "", fmt.Errorf("unsupported variant type: %s", variant.Type)
	}
}

// renderBullets processes bullet-type variants
func (r *PromptRenderer) renderBullets(variant *VariantDefinition, sectionSelection SectionSelection) string {
	if len(variant.Bullets) == 0 {
		return ""
	}

	var bullets strings.Builder
	bulletPrefix := r.getBulletPrefix()

	// Check if variant has content with {{.}} placeholder
	if variant.Content != "" {
		// Collect selected bullets
		var selectedBullets []string
		for i, bullet := range variant.Bullets {
			bulletKey := fmt.Sprintf("%d", i)
			if sectionSelection.SelectedBullets != nil && sectionSelection.SelectedBullets[bulletKey] {
				selectedBullets = append(selectedBullets, bulletPrefix+bullet)
			}
		}

		// Replace {{.}} with bullet list
		content := strings.ReplaceAll(variant.Content, "{{.}}", strings.Join(selectedBullets, "\n"))
		return content
	}

	// Default: just render selected bullets
	for i, bullet := range variant.Bullets {
		bulletKey := fmt.Sprintf("%d", i)
		if sectionSelection.SelectedBullets != nil && sectionSelection.SelectedBullets[bulletKey] {
			bullets.WriteString(bulletPrefix)
			bullets.WriteString(bullet)
			bullets.WriteString("\n")
		}
	}

	return strings.TrimSuffix(bullets.String(), "\n")
}

// getBulletPrefix returns the bullet prefix to use
func (r *PromptRenderer) getBulletPrefix() string {
	if r.dslFile.Globals != nil && r.dslFile.Globals.BulletPrefix != "" {
		return r.dslFile.Globals.BulletPrefix
	}
	return "- "
}

// substituteVariables performs variable substitution using Go templates with sprig functions
func (r *PromptRenderer) substituteVariables(content string, templateDef *TemplateDefinition, selection *SelectionState) (string, error) {
	// Prepare variables map with defaults for missing variables
	variables := make(map[string]string)

	// Set default values for all defined variables
	for varName, varConfig := range templateDef.Variables {
		defaultValue := fmt.Sprintf("DEFAULT_%s", strings.ToUpper(varName))
		if varConfig.Hint != "" {
			defaultValue = fmt.Sprintf("DEFAULT_%s", strings.ToUpper(varName))
		}
		variables[varName] = defaultValue
	}

	// Override with actual values from selection
	for varName, value := range selection.Variables {
		// Handle file references
		if strings.HasPrefix(value, "@") {
			filename := strings.TrimPrefix(value, "@")
			fileContent, err := r.readFileContent(filename)
			if err != nil {
				// Use error message as value to show the problem
				variables[varName] = fmt.Sprintf("ERROR_READING_FILE: %s", err.Error())
			} else {
				variables[varName] = fileContent
			}
		} else {
			variables[varName] = value
		}
	}

	// Create template with sprig functions
	tmpl, err := template.New("prompt").Funcs(sprig.TxtFuncMap()).Parse(content)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse template")
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", errors.Wrap(err, "failed to execute template")
	}

	return buf.String(), nil
}

// readFileContent reads content from a file, with error handling
func (r *PromptRenderer) readFileContent(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read file: %s", filename)
	}
	return string(data), nil
}

// cleanupPrompt removes excessive whitespace and normalizes line endings
func (r *PromptRenderer) cleanupPrompt(prompt string) string {
	// Remove trailing whitespace from lines
	lines := strings.Split(prompt, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	prompt = strings.Join(lines, "\n")

	// Remove excessive blank lines (more than 2 consecutive)
	re := regexp.MustCompile(`\n{3,}`)
	prompt = re.ReplaceAllString(prompt, "\n\n")

	// Trim leading and trailing whitespace
	return strings.TrimSpace(prompt)
}

// CreateDefaultSelection creates a default selection state for a template
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

	// Initialize sections with first variant selected
	for _, section := range templateDef.Sections {
		if len(section.Variants) > 0 {
			sectionSelection := SectionSelection{
				Variant:         section.Variants[0].ID,
				SelectedBullets: make(map[string]bool),
				VariantEnabled:  false, // Default toggles to off
			}

			// Select all bullets by default so users can deselect as needed.
			for i := range section.Variants[0].Bullets {
				bulletKey := fmt.Sprintf("%d", i)
				sectionSelection.SelectedBullets[bulletKey] = true
			}

			selection.Sections[section.ID] = sectionSelection
		}
	}

	return selection
}
