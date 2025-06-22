package main

import (
	"fmt"
	"log"
)

// TestUIImprovements tests the new UI improvements
func TestUIImprovements() {
	// Load DSL file
	dslFile, err := ParseDSLFile("templates.yml")
	if err != nil {
		log.Fatalf("Failed to load DSL: %v", err)
	}

	fmt.Printf("Testing UI improvements with %d templates\n", len(dslFile.Templates))

	// Test with "code-review" template which has bullets
	var template *TemplateDefinition
	for i := range dslFile.Templates {
		if dslFile.Templates[i].ID == "code-review" {
			template = &dslFile.Templates[i]
			break
		}
	}
	if template == nil {
		log.Fatal("Could not find code-review template")
	}

	fmt.Printf("Testing template: %s\n", template.Label)

	// Create renderer
	renderer := NewPromptRenderer(dslFile)

	// Create default selection to test default bullet selection
	selection := CreateDefaultSelection(template)

	// Show what default selections were made
	fmt.Println("\n=== Default Selections ===")
	for sectionID, sectionSelection := range selection.Sections {
		fmt.Printf("Section: %s, Variant: %s\n", sectionID, sectionSelection.Variant)
		if len(sectionSelection.SelectedBullets) > 0 {
			fmt.Printf("  Default bullets selected: %d\n", len(sectionSelection.SelectedBullets))
			for bulletKey, selected := range sectionSelection.SelectedBullets {
				if selected {
					fmt.Printf("    - Bullet %s: selected\n", bulletKey)
				}
			}
		}
		if sectionSelection.VariantEnabled {
			fmt.Printf("  Toggle enabled: true\n")
		}
	}

	// Set some variables for testing
	selection.Variables["code_snippet"] = "func calculateSum(a, b int) int {\n    return a + b\n}"
	selection.Variables["language"] = "go"

	// Render prompt with defaults
	prompt, err := renderer.RenderPrompt(template, selection)
	if err != nil {
		log.Fatalf("Failed to render prompt: %v", err)
	}

	fmt.Println("\n=== Rendered Prompt with Defaults ===")
	fmt.Println(prompt)
	fmt.Println("=== End Prompt ===\n")

	// Now test section and variant labels
	fmt.Println("=== Testing Labels and Descriptions ===")
	for _, section := range template.Sections {
		fmt.Printf("Section ID: %s", section.ID)
		if section.Label != "" {
			fmt.Printf(", Label: %s", section.Label)
		}
		fmt.Println()

		for _, variant := range section.Variants {
			fmt.Printf("  Variant ID: %s", variant.ID)
			if variant.Label != "" {
				fmt.Printf(", Label: %s", variant.Label)
			}
			if variant.Description != "" {
				fmt.Printf(", Description: %s", variant.Description)
			}
			fmt.Println()
		}
	}

	fmt.Println("UI improvements test completed successfully!")
}
