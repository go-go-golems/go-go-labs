package dsl

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseRuleFile parses a YAML rule file into a Rule struct
func ParseRuleFile(filename string) (*Rule, error) {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read rule file: %w", err)
	}

	// Parse YAML
	return ParseRuleString(string(data))
}

// ParseRuleString parses a YAML string into a Rule struct
func ParseRuleString(yamlStr string) (*Rule, error) {
	var rule Rule

	// Parse YAML into Rule struct
	err := yaml.Unmarshal([]byte(yamlStr), &rule)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate basic requirements
	if err := validateRule(&rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// validateRule performs basic validation on the rule
func validateRule(rule *Rule) error {
	// Check if search section exists
	if rule.Search.From == "" && rule.Search.Since == "" &&
		rule.Search.Before == "" && rule.Search.On == "" &&
		rule.Search.WithinDays == 0 {
		return fmt.Errorf("rule must contain at least one search criterion")
	}

	// Check if output format is valid
	switch rule.Output.Format {
	case "text", "json", "table", "":
		// Valid formats (empty defaults to text)
		if rule.Output.Format == "" {
			rule.Output.Format = "text"
		}
	default:
		return fmt.Errorf("invalid output format: %s", rule.Output.Format)
	}

	// Check if at least one field is specified
	if len(rule.Output.Fields) == 0 {
		return fmt.Errorf("output must specify at least one field")
	}

	return nil
}
