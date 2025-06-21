package main

import "time"

// DSLFile represents the root structure of a YAML DSL file
type DSLFile struct {
	Version   int                `yaml:"version"`
	Globals   *GlobalConfig      `yaml:"globals,omitempty"`
	Templates []TemplateDefinition `yaml:"templates"`
}

// GlobalConfig contains global defaults
type GlobalConfig struct {
	ModelFallback string `yaml:"model_fallback,omitempty"`
	BulletPrefix  string `yaml:"bullet_prefix,omitempty"`
}

// TemplateDefinition represents a single prompt template
type TemplateDefinition struct {
	ID        string                    `yaml:"id"`
	Label     string                    `yaml:"label"`
	Model     string                    `yaml:"model,omitempty"`
	Variables map[string]VariableConfig `yaml:"variables,omitempty"`
	Sections  []SectionDefinition       `yaml:"sections"`
}

// VariableConfig defines a template variable
type VariableConfig struct {
	Hint string `yaml:"hint"`
	Type string `yaml:"type"` // Currently only "text" supported
}

// SectionDefinition represents a section with variants
type SectionDefinition struct {
	ID       string             `yaml:"id"`
	Variants []VariantDefinition `yaml:"variants"`
}

// VariantDefinition represents a variant within a section
type VariantDefinition struct {
	ID      string       `yaml:"id"`
	Type    string       `yaml:"type"` // "text" or "bullets"
	Content string       `yaml:"content,omitempty"`
	Groups  []BulletGroup `yaml:"groups,omitempty"`
}

// BulletGroup represents a group of bullet points
type BulletGroup struct {
	ID      string   `yaml:"id"`
	Bullets []string `yaml:"bullets"`
}

// SelectionState represents the current user selections for a template
type SelectionState struct {
	TemplateID string                       `yaml:"template_id"`
	Timestamp  time.Time                    `yaml:"timestamp"`
	Variables  map[string]string            `yaml:"variables,omitempty"`
	Sections   map[string]SectionSelection  `yaml:"sections,omitempty"`
}

// SectionSelection represents selections for a specific section
type SectionSelection struct {
	Variant string   `yaml:"variant"`
	Groups  []string `yaml:"groups,omitempty"`
}

// AppState represents the current application state
type AppState int

const (
	StateTemplateList AppState = iota
	StateTemplateConfig
)
