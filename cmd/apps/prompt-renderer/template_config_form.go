package main

import (
	"fmt"
	"sort"
)

// FormOption represents a selectable option in a section selector.
// ID is the internal identifier (e.g. variant ID) while Label is what gets
// rendered in the UI. We also keep the Description around so it can be shown
// next to the label without having to look it up again at render time.
type FormOption struct {
	ID          string
	Label       string
	Description string
}

// FormItem represents a configurable item in the form
type FormItem struct {
	Type        string // "variable", "section", "bullet", "toggle", "bullet_header"
	Key         string // variable name, section ID, bullet index, or variant ID
	SectionID   string // for bullet items
	VariantID   string // for bullet items
	BulletIndex int    // for bullet items
	Label       string
	Value       string
	Hint        string
	Options     []FormOption // for sections
	Selected    bool         // for bullets and toggles
}

// FormHandler manages form structure and navigation
type FormHandler struct {
	Items      []FormItem
	FocusIndex int
}

// NewFormHandler creates a new form handler
func NewFormHandler() *FormHandler {
	return &FormHandler{
		Items:      []FormItem{},
		FocusIndex: 0,
	}
}

// RebuildFormItems constructs the form items list from template and selection
func (f *FormHandler) RebuildFormItems(template *TemplateDefinition, selection *SelectionState) {
	f.Items = []FormItem{}

	// Add variables
	varNames := make([]string, 0, len(template.Variables))
	for name := range template.Variables {
		varNames = append(varNames, name)
	}
	sort.Strings(varNames) // Sort for consistent ordering

	for _, name := range varNames {
		varConfig := template.Variables[name]
		value := ""
		if val, exists := selection.Variables[name]; exists {
			value = val
		}

		f.Items = append(f.Items, FormItem{
			Type:  "variable",
			Key:   name,
			Label: name,
			Value: value,
			Hint:  varConfig.Hint,
		})
	}

	// Add sections and their variants
	for _, section := range template.Sections {
		sectionSelection := selection.Sections[section.ID]

		// Add section variant selector (only if multiple variants)
		if len(section.Variants) > 1 {
			options := make([]FormOption, len(section.Variants))
			for i, variant := range section.Variants {
				label := variant.Label
				if label == "" {
					label = variant.ID
				}
				options[i] = FormOption{
					ID:          variant.ID,
					Label:       label,
					Description: variant.Description,
				}
			}

			sectionLabel := section.Label
			if sectionLabel == "" {
				sectionLabel = section.ID
			}

			f.Items = append(f.Items, FormItem{
				Type:    "section",
				Key:     section.ID,
				Label:   sectionLabel,
				Value:   sectionSelection.Variant,
				Options: options,
			})
		}

		// Add items for current variant
		for _, variant := range section.Variants {
			if variant.ID == sectionSelection.Variant {
				switch variant.Type {
				case "toggle":
					// Add toggle for this variant
					variantLabel := variant.Label
					if variantLabel == "" {
						variantLabel = variant.ID
					}

					f.Items = append(f.Items, FormItem{
						Type:      "toggle",
						Key:       variant.ID,
						SectionID: section.ID,
						VariantID: variant.ID,
						Label:     variantLabel,
						Selected:  sectionSelection.VariantEnabled,
						Hint:      variant.Description, // Store description in hint for display
					})
				case "bullets":
					// Add section header for bullets
					sectionLabel := section.Label
					if sectionLabel == "" {
						sectionLabel = section.ID
					}

					variantLabel := variant.Label
					if variantLabel == "" {
						variantLabel = variant.ID
					}

					bulletSectionTitle := fmt.Sprintf("%s (%s)", sectionLabel, variantLabel)
					if variant.Description != "" {
						bulletSectionTitle += " - " + variant.Description
					}

					f.Items = append(f.Items, FormItem{
						Type:  "bullet_header",
						Key:   fmt.Sprintf("%s_%s_header", section.ID, variant.ID),
						Label: bulletSectionTitle,
					})

					// Add individual bullets
					for i, bullet := range variant.Bullets {
						bulletKey := fmt.Sprintf("%d", i)
						selected := false
						if sectionSelection.SelectedBullets != nil {
							selected = sectionSelection.SelectedBullets[fmt.Sprintf("%s_%s", variant.ID, bulletKey)]
						}

						f.Items = append(f.Items, FormItem{
							Type:        "bullet",
							Key:         bulletKey,
							SectionID:   section.ID,
							VariantID:   variant.ID,
							BulletIndex: i,
							Label:       bullet,
							Selected:    selected,
						})
					}
				}
			}
		}
	}
}

// NavigateUp moves focus up in the form
func (f *FormHandler) NavigateUp() {
	if len(f.Items) > 0 && f.FocusIndex > 0 {
		f.FocusIndex--
	}
}

// NavigateDown moves focus down in the form
func (f *FormHandler) NavigateDown() {
	if len(f.Items) > 0 && f.FocusIndex < len(f.Items)-1 {
		f.FocusIndex++
	}
}

// NavigateNext moves to next item (with wraparound)
func (f *FormHandler) NavigateNext() {
	if len(f.Items) > 0 {
		f.FocusIndex = (f.FocusIndex + 1) % len(f.Items)
	}
}

// NavigatePrev moves to previous item (with wraparound)
func (f *FormHandler) NavigatePrev() {
	if len(f.Items) > 0 {
		f.FocusIndex = (f.FocusIndex - 1 + len(f.Items)) % len(f.Items)
	}
}

// GetFocusedItem returns the currently focused form item
func (f *FormHandler) GetFocusedItem() *FormItem {
	if len(f.Items) > 0 && f.FocusIndex >= 0 && f.FocusIndex < len(f.Items) {
		return &f.Items[f.FocusIndex]
	}
	return nil
}

// UpdateItemValue updates the value of the focused item
func (f *FormHandler) UpdateItemValue(value string) {
	if item := f.GetFocusedItem(); item != nil && item.Type == "variable" {
		f.Items[f.FocusIndex].Value = value
	}
}
