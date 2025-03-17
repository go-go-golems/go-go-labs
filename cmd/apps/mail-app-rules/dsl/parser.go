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

	// Validate the rule using the Validate method
	if err := rule.Validate(); err != nil {
		return nil, err
	}

	// Set default values if needed
	if rule.Output.Format == "" {
		rule.Output.Format = "text"
	}

	return &rule, nil
}
