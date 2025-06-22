package main

import (
	"fmt"
	"time"
)

// StateManager handles template selection state and updates
type StateManager struct {
	Template  *TemplateDefinition
	Selection *SelectionState
	Renderer  *PromptRenderer
	Preview   string
}

// NewStateManager creates a new state manager
func NewStateManager(template *TemplateDefinition, renderer *PromptRenderer) *StateManager {
	selection := CreateDefaultSelection(template)
	selection.Timestamp = time.Now()

	sm := &StateManager{
		Template:  template,
		Selection: selection,
		Renderer:  renderer,
	}

	sm.UpdatePreview()
	return sm
}

// UpdatePreview regenerates the prompt preview
func (s *StateManager) UpdatePreview() {
	preview, err := s.Renderer.RenderPrompt(s.Template, s.Selection)
	if err != nil {
		s.Preview = fmt.Sprintf("❌ Error rendering prompt: %v\n\nPlease check your template configuration and try again.", err)
	} else {
		s.Preview = preview
	}
}

// UpdateVariable updates a template variable value
func (s *StateManager) UpdateVariable(name, value string) {
	s.Selection.Variables[name] = value
	s.UpdatePreview()
}

// CycleSectionVariant cycles through section variants
func (s *StateManager) CycleSectionVariant(sectionID string) {
	sectionSelection := s.Selection.Sections[sectionID]

	// Find current variant index
	var currentIndex int
	for _, section := range s.Template.Sections {
		if section.ID == sectionID {
			for i, variant := range section.Variants {
				if variant.ID == sectionSelection.Variant {
					currentIndex = i
					break
				}
			}

			// Cycle to next variant
			nextIndex := (currentIndex + 1) % len(section.Variants)
			nextVariant := section.Variants[nextIndex]
			sectionSelection.Variant = nextVariant.ID
			// Preserve bullet selections across variant changes so users don't lose work (keys now namespaced by variant).
			sectionSelection.VariantEnabled = false // Reset toggle state only

			// If the newly selected variant is of type bullets and its keys are not present yet, default them to on.
			if nextVariant.Type == "bullets" {
				if sectionSelection.SelectedBullets == nil {
					sectionSelection.SelectedBullets = make(map[string]bool)
				}
				for i := range nextVariant.Bullets {
					key := fmt.Sprintf("%s_%d", nextVariant.ID, i)
					if _, ok := sectionSelection.SelectedBullets[key]; !ok {
						sectionSelection.SelectedBullets[key] = true
					}
				}
			}

			s.Selection.Sections[sectionID] = sectionSelection
			break
		}
	}

	s.UpdatePreview()
}

// ToggleBullet toggles a bullet selection. Selection is tracked per variant.
func (s *StateManager) ToggleBullet(sectionID, variantID, bulletKey string) {
	sectionSelection := s.Selection.Sections[sectionID]

	// Initialize map if needed
	if sectionSelection.SelectedBullets == nil {
		sectionSelection.SelectedBullets = make(map[string]bool)
	}

	fullKey := variantID + "_" + bulletKey

	// Toggle bullet selection for this variant
	sectionSelection.SelectedBullets[fullKey] = !sectionSelection.SelectedBullets[fullKey]
	s.Selection.Sections[sectionID] = sectionSelection

	s.UpdatePreview()
}

// ToggleVariant toggles a variant on/off for toggle-type variants
func (s *StateManager) ToggleVariant(sectionID, variantID string) {
	sectionSelection := s.Selection.Sections[sectionID]
	sectionSelection.VariantEnabled = !sectionSelection.VariantEnabled
	s.Selection.Sections[sectionID] = sectionSelection

	s.UpdatePreview()
}

// HandleToggle handles toggling based on form item type
func (s *StateManager) HandleToggle(item *FormItem) {
	switch item.Type {
	case "bullet":
		s.ToggleBullet(item.SectionID, item.VariantID, item.Key)
	case "toggle":
		s.ToggleVariant(item.SectionID, item.VariantID)
	}
}

// GetTemplate returns the template definition
func (s *StateManager) GetTemplate() *TemplateDefinition {
	return s.Template
}

// GetSelection returns the current selection state
func (s *StateManager) GetSelection() *SelectionState {
	return s.Selection
}

// GetPreview returns the current preview
func (s *StateManager) GetPreview() string {
	return s.Preview
}
