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

	// Test with first template
	template := &dslFile.Templates[0]
	fmt.Printf("Testing template: %s\n", template.Label)

	// Create renderer
	renderer := NewPromptRenderer(dslFile)

	// Create test selection
	selection := CreateDefaultSelection(template)
	selection.Variables["code_snippet"] = "func main() {\n    fmt.Println(\"Hello, World!\")\n}"
	selection.Variables["language"] = "go"

	// Select some bullet groups
	if sectionSelection, exists := selection.Sections["review_aspects"]; exists {
		sectionSelection.Groups = []string{"quality", "security"}
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
