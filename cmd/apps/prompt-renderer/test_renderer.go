package main

import (
	"fmt"
	"log"
)

// TestRenderer tests the core rendering functionality
func TestRenderer() {
	// Load DSL file
	dslFile, err := ParseDSLFile("templates.yml")
	if err != nil {
		log.Fatalf("Failed to load DSL: %v", err)
	}

	fmt.Printf("Loaded %d templates\n", len(dslFile.Templates))

	// Test with "with-context" template to check toggle functionality
	var template *TemplateDefinition
	for i := range dslFile.Templates {
		if dslFile.Templates[i].ID == "with-context" {
			template = &dslFile.Templates[i]
			break
		}
	}
	if template == nil {
		template = &dslFile.Templates[0] // fallback
	}
	fmt.Printf("Testing template: %s\n", template.Label)

	// Create renderer
	renderer := NewPromptRenderer(dslFile)

	// Create test selection
	selection := CreateDefaultSelection(template)
	selection.Variables["code_snippet"] = "func main() {\n    fmt.Println(\"Hello, World!\")\n}"

	// Test toggle functionality
	if sectionSelection, exists := selection.Sections["context_request"]; exists {
		sectionSelection.VariantEnabled = true // Enable the toggle
		selection.Sections["context_request"] = sectionSelection
		fmt.Printf("Enabled toggle for context_request: %v\n", sectionSelection.VariantEnabled)
	}

	// Also test bullet selection if present
	if sectionSelection, exists := selection.Sections["review_aspects"]; exists {
		sectionSelection.SelectedBullets = map[string]bool{
			"0": true, // Code quality and readability
			"1": true, // Best practices adherence
			"3": true, // Security vulnerabilities
		}
		selection.Sections["review_aspects"] = sectionSelection
	}

	// Render prompt
	prompt, err := renderer.RenderPrompt(template, selection)
	if err != nil {
		log.Fatalf("Failed to render prompt: %v", err)
	}

	fmt.Println("\n--- Rendered Prompt ---")
	fmt.Println(prompt)
	fmt.Println("--- End Prompt ---\n")

	fmt.Println("Renderer test completed successfully!")
}
