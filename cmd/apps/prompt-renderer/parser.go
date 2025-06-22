package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// ParseDSLFile loads and validates a DSL file from the given path
func ParseDSLFile(path string) (*DSLFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read DSL file: %s", path)
	}

	var dslFile DSLFile
	if err := yaml.Unmarshal(data, &dslFile); err != nil {
		return nil, errors.Wrapf(err, "❌ Failed to parse YAML in file: %s. Please check YAML syntax (indentation, colons, quotes)", path)
	}

	if err := validateDSLFile(&dslFile); err != nil {
		return nil, errors.Wrapf(err, "validation failed for file: %s", path)
	}

	return &dslFile, nil
}

// validateDSLFile performs validation on the parsed DSL structure
func validateDSLFile(dsl *DSLFile) error {
	if dsl.Version == 0 {
		return errors.New("version field is required")
	}

	if dsl.Version != 1 {
		return fmt.Errorf("unsupported DSL version: %d (supported: 1)", dsl.Version)
	}

	if len(dsl.Templates) == 0 {
		return errors.New("at least one template is required")
	}

	templateIDs := make(map[string]bool)
	for i, template := range dsl.Templates {
		if err := validateTemplate(&template, i); err != nil {
			return errors.Wrapf(err, "template %d validation failed", i)
		}

		if templateIDs[template.ID] {
			return fmt.Errorf("duplicate template ID: %s", template.ID)
		}
		templateIDs[template.ID] = true
	}

	return nil
}

// validateTemplate validates a single template definition
func validateTemplate(template *TemplateDefinition, index int) error {
	if template.ID == "" {
		return fmt.Errorf("template at index %d missing required ID field", index)
	}

	if template.Label == "" {
		return fmt.Errorf("template '%s' missing required label field", template.ID)
	}

	if len(template.Sections) == 0 {
		return fmt.Errorf("template '%s' must have at least one section", template.ID)
	}

	// Validate variables
	for varName, varConfig := range template.Variables {
		if varConfig.Type != "" && varConfig.Type != "text" {
			return fmt.Errorf("template '%s' variable '%s' has unsupported type: %s", template.ID, varName, varConfig.Type)
		}
	}

	// Validate sections
	sectionIDs := make(map[string]bool)
	for i, section := range template.Sections {
		if err := validateSection(&section, template.ID, i); err != nil {
			return err
		}

		if sectionIDs[section.ID] {
			return fmt.Errorf("template '%s' has duplicate section ID: %s", template.ID, section.ID)
		}
		sectionIDs[section.ID] = true
	}

	return nil
}

// validateSection validates a single section definition
func validateSection(section *SectionDefinition, templateID string, index int) error {
	if section.ID == "" {
		return fmt.Errorf("template '%s' section at index %d missing required ID field", templateID, index)
	}

	if len(section.Variants) == 0 {
		return fmt.Errorf("template '%s' section '%s' must have at least one variant", templateID, section.ID)
	}

	// Validate variants
	variantIDs := make(map[string]bool)
	for i, variant := range section.Variants {
		if err := validateVariant(&variant, templateID, section.ID, i); err != nil {
			return err
		}

		if variantIDs[variant.ID] {
			return fmt.Errorf("template '%s' section '%s' has duplicate variant ID: %s", templateID, section.ID, variant.ID)
		}
		variantIDs[variant.ID] = true
	}

	return nil
}

// validateVariant validates a single variant definition
func validateVariant(variant *VariantDefinition, templateID, sectionID string, index int) error {
	if variant.ID == "" {
		return fmt.Errorf("template '%s' section '%s' variant at index %d missing required ID field", templateID, sectionID, index)
	}

	if variant.Type != "text" && variant.Type != "bullets" && variant.Type != "toggle" {
		return fmt.Errorf("template '%s' section '%s' variant '%s' has invalid type: %s (must be 'text', 'bullets', or 'toggle')", templateID, sectionID, variant.ID, variant.Type)
	}

	if variant.Type == "text" && variant.Content == "" {
		return fmt.Errorf("template '%s' section '%s' variant '%s' of type 'text' requires content field", templateID, sectionID, variant.ID)
	}

	if variant.Type == "toggle" && variant.Content == "" {
		return fmt.Errorf("template '%s' section '%s' variant '%s' of type 'toggle' requires content field", templateID, sectionID, variant.ID)
	}

	if variant.Type == "bullets" && len(variant.Bullets) == 0 {
		return fmt.Errorf("template '%s' section '%s' variant '%s' of type 'bullets' requires at least one bullet", templateID, sectionID, variant.ID)
	}

	return nil
}

// LoadDefaultDSLFile attempts to load a DSL file from common locations
func LoadDefaultDSLFile() (*DSLFile, error) {
	// Try current directory first
	if _, err := os.Stat("templates.yml"); err == nil {
		return ParseDSLFile("templates.yml")
	}

	// Try XDG data directory
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			xdgDataHome = filepath.Join(homeDir, ".local", "share")
		}
	}

	if xdgDataHome != "" {
		dslPath := filepath.Join(xdgDataHome, "prompt-builder", "templates.yml")
		if _, err := os.Stat(dslPath); err == nil {
			return ParseDSLFile(dslPath)
		}
	}

	return nil, errors.New("no DSL file found in current directory or XDG data directory")
}
