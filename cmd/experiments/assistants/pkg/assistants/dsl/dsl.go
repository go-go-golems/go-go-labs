package dsl

import (
	"github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/pkg"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/pkg/assistants"
	"gopkg.in/yaml.v2"
	"net/http"
	"os"
)

// AssistantDSL represents the top-level structure of the YAML DSL for an OpenAI Assistant.
type AssistantDSL struct {
	Assistant       Assistant        `yaml:"assistant"`
	FileDefinitions []FileDefinition `yaml:"file_definitions"`
}

// Assistant represents the details of the Assistant configuration.
type Assistant struct {
	Model        string                 `yaml:"model"`
	Name         string                 `yaml:"name,omitempty"`
	Description  string                 `yaml:"description,omitempty"`
	Instructions string                 `yaml:"instructions,omitempty"`
	Tools        []assistants.Tool      `yaml:"tools,omitempty"`
	Files        []string               `yaml:"files,omitempty"` // Filenames
	Metadata     map[string]interface{} `yaml:"metadata,omitempty"`
}

// FileDefinition represents the definition of a file.
type FileDefinition struct {
	Filename    string `yaml:"filename"`
	Description string `yaml:"description"`
}

// ParseAssistantDSL takes a byte slice of YAML and returns an AssistantDSL structure.
func ParseAssistantDSL(yamlData []byte) (*AssistantDSL, error) {
	var dsl AssistantDSL
	err := yaml.Unmarshal(yamlData, &dsl)
	if err != nil {
		return nil, err
	}
	return &dsl, nil
}

// CreateAssistantAndFiles creates the files and then the assistant,
// returning the assistant on success or the file IDs of created files on error.
func (a *AssistantDSL) CreateAssistantAndFiles(client *http.Client, baseURL, apiKey string) (*assistants.Assistant, []string, error) {
	var createdFileIDs []string

	// Create files and gather IDs
	for _, fileDef := range a.FileDefinitions {
		content, err := os.ReadFile(fileDef.Filename)
		if err != nil {
			return nil, createdFileIDs, err
		}
		createFileReq := pkg.CreateFileRequest{
			File:    content,
			Purpose: "assistants",
		}
		createdFile, err := pkg.CreateFile(client, baseURL, apiKey, createFileReq)
		if err != nil {
			return nil, createdFileIDs, err
		}
		createdFileIDs = append(createdFileIDs, createdFile.ID)
	}

	assistant := assistants.Assistant{
		Object:       "",
		Name:         a.Assistant.Name,
		Description:  a.Assistant.Description,
		Model:        a.Assistant.Model,
		Instructions: a.Assistant.Instructions,
		Tools:        a.Assistant.Tools,
		FileIDs:      createdFileIDs,
		Metadata:     a.Assistant.Metadata,
	}

	// Create the assistant
	createdAssistant, err := assistants.CreateAssistant(apiKey, assistant)
	if err != nil {
		return nil, createdFileIDs, err
	}

	return createdAssistant, createdFileIDs, nil
}
