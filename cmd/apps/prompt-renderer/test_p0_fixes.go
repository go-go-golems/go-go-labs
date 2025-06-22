package main

import (
	"fmt"
	"log"
)

// TestP0Fixes specifically tests the P0 fixes
func TestP0Fixes() {
	fmt.Println("=== Testing P0 Fixes ===")

	// Load DSL file
	dslFile, err := ParseDSLFile("templates.yml")
	if err != nil {
		log.Fatalf("Failed to load DSL: %v", err)
	}

	fmt.Printf("✓ Loaded %d templates successfully\n", len(dslFile.Templates))

	// Test 1: Check "Code Review with Context" template exists and renders correctly
	var withContextTemplate *TemplateDefinition
	for i := range dslFile.Templates {
		if dslFile.Templates[i].ID == "with-context" {
			withContextTemplate = &dslFile.Templates[i]
			break
		}
	}

	if withContextTemplate == nil {
		log.Fatal("❌ Failed to find 'with-context' template")
	}

	fmt.Printf("✓ Found 'Code Review with Context' template: %s\n", withContextTemplate.Label)

	// Test 2: Test section rendering with toggle functionality
	renderer := NewPromptRenderer(dslFile)
	selection := CreateDefaultSelection(withContextTemplate)

	// Set up test variables
	selection.Variables["code_snippet"] = `func example() {
    fmt.Println("Testing section rendering")
}`

	// Test toggle section rendering - OFF state
	fmt.Println("\n--- Test: Toggle section OFF ---")
	if sectionSelection, exists := selection.Sections["context_request"]; exists {
		sectionSelection.VariantEnabled = false // Disable toggle
		selection.Sections["context_request"] = sectionSelection
	}

	prompt, err := renderer.RenderPrompt(withContextTemplate, selection)
	if err != nil {
		log.Fatalf("❌ Failed to render prompt with toggle OFF: %v", err)
	}
	fmt.Printf("✓ Rendered prompt with toggle OFF (length: %d chars)\n", len(prompt))
	fmt.Println("Preview:", prompt[:min(100, len(prompt))]+"...")

	// Test toggle section rendering - ON state
	fmt.Println("\n--- Test: Toggle section ON ---")
	if sectionSelection, exists := selection.Sections["context_request"]; exists {
		sectionSelection.VariantEnabled = true // Enable toggle
		selection.Sections["context_request"] = sectionSelection
	}

	prompt, err = renderer.RenderPrompt(withContextTemplate, selection)
	if err != nil {
		log.Fatalf("❌ Failed to render prompt with toggle ON: %v", err)
	}
	fmt.Printf("✓ Rendered prompt with toggle ON (length: %d chars)\n", len(prompt))
	fmt.Println("Preview:", prompt[:min(150, len(prompt))]+"...")

	// Test 3: Test bullet section rendering
	fmt.Println("\n--- Test: Bullet section rendering ---")
	var codeReviewTemplate *TemplateDefinition
	for i := range dslFile.Templates {
		if dslFile.Templates[i].ID == "code-review" {
			codeReviewTemplate = &dslFile.Templates[i]
			break
		}
	}

	if codeReviewTemplate != nil {
		bulletSelection := CreateDefaultSelection(codeReviewTemplate)
		bulletSelection.Variables["code_snippet"] = "test code"
		bulletSelection.Variables["language"] = "go"

		// Enable some bullets
		if sectionSelection, exists := bulletSelection.Sections["review_aspects"]; exists {
			sectionSelection.SelectedBullets = map[string]bool{
				"0": true, // Code quality and readability
				"2": true, // Performance considerations
				"3": true, // Security vulnerabilities
			}
			bulletSelection.Sections["review_aspects"] = sectionSelection
		}

		bulletPrompt, err := renderer.RenderPrompt(codeReviewTemplate, bulletSelection)
		if err != nil {
			log.Fatalf("❌ Failed to render bullet prompt: %v", err)
		}
		fmt.Printf("✓ Rendered bullet prompt (length: %d chars)\n", len(bulletPrompt))
	}

	// Test 4: Template config model bounds checking
	fmt.Println("\n--- Test: Template config model bounds ---")
	configModel := NewTemplateConfigModel(withContextTemplate, renderer)
	fmt.Printf("✓ Created config model with %d form items\n", len(configModel.GetFormItems()))

	// Test navigation bounds
	configModel.SetFocusIndex(-1) // Invalid index
	configModel.HandleToggle()    // Should not crash
	fmt.Println("✓ Bounds checking works for negative index")

	configModel.SetFocusIndex(999) // Invalid index
	configModel.HandleToggle()     // Should not crash
	fmt.Println("✓ Bounds checking works for out-of-bounds index")

	fmt.Println("\n=== All P0 Fix Tests Passed! ===")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
