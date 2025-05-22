package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Adapter defines the interface for LLM operations
type Adapter interface {
	// GeneratePullRequestSpec calls the LLM to generate PR details
	GeneratePullRequestSpec(ctx context.Context, llmCommand string, promptParams map[string]interface{}, diffContent string, commitHistoryContent string, otherContext map[string]string) (string, error)
	// CountTokens calls the LLM's token counter for given text
	CountTokens(ctx context.Context, llmCommand string, textContent string) (int, error)
}

// PinocchioAdapter provides an implementation that calls pinocchio
type PinocchioAdapter struct{}

// GeneratePullRequestSpec implements the Adapter interface
func (a *PinocchioAdapter) GeneratePullRequestSpec(ctx context.Context, llmCommand string, promptParams map[string]interface{}, diffContent string, commitHistoryContent string, otherContext map[string]string) (string, error) {
	if llmCommand == "" {
		llmCommand = "pinocchio code create-pull-request"
	}

	// Prepare command parts
	cmdParts := strings.Split(llmCommand, " ")

	// Build the command with appropriate flags
	args := make([]string, 0, len(cmdParts)-1)
	if len(cmdParts) > 1 {
		args = append(args, cmdParts[1:]...)
	}

	// Add description if provided
	if description, ok := promptParams["description"].(string); ok && description != "" {
		args = append(args, "--description", description)
	}

	// Add user title suggestion if provided
	if title, ok := promptParams["user_title_suggestion"].(string); ok && title != "" {
		args = append(args, "--title", title)
	}

	// Add style if provided
	if style, ok := promptParams["llm_style"].(string); ok && style != "" {
		args = append(args, "--style", style)
	}

	// Add additional system prompt if provided
	if systemPrompt, ok := promptParams["additional_system_prompt"].(string); ok && systemPrompt != "" {
		args = append(args, "--system-prompt", systemPrompt)
	}

	// Add additional user prompts if provided
	if userPrompts, ok := promptParams["additional_user_prompts"].([]string); ok {
		for _, prompt := range userPrompts {
			args = append(args, "--user-prompt", prompt)
		}
	}

	// Add raw params if provided
	if rawParams, ok := promptParams["raw_params"].(map[string]string); ok {
		for k, v := range rawParams {
			args = append(args, fmt.Sprintf("--%s=%s", k, v))
		}
	}

	// Execute the command
	cmd := exec.CommandContext(ctx, cmdParts[0], args...)

	// Set diff content as stdin
	input := diffContent
	if commitHistoryContent != "" {
		input += "\n\n" + commitHistoryContent
	}

	// Add other context
	if len(otherContext) > 0 {
		contextJSON, err := json.Marshal(otherContext)
		if err == nil {
			args = append(args, "--context", string(contextJSON))
		}
	}

	// Set stdin with the diff and commit history
	cmd.Stdin = strings.NewReader(input)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		return "", errors.Wrapf(err, "failed to run LLM command: %s", stderr.String())
	}

	return stdout.String(), nil
}

// CountTokens implements the Adapter interface
func (a *PinocchioAdapter) CountTokens(ctx context.Context, llmCommand string, textContent string) (int, error) {
	// For now, we'll implement a simple estimation (1 token ≈ 4 chars)
	// In a real implementation, you would call the LLM's token counter
	return len(textContent) / 4, nil
}

// NewPinocchioAdapter creates a new adapter for pinocchio
func NewPinocchioAdapter() *PinocchioAdapter {
	return &PinocchioAdapter{}
}

// MockAdapter provides a mock implementation for testing
type MockAdapter struct {
	MockYAMLOutput string
	MockTokenCount int
	MockError      error
}

// GeneratePullRequestSpec implements the Adapter interface with mock data
func (m *MockAdapter) GeneratePullRequestSpec(ctx context.Context, llmCommand string, promptParams map[string]interface{}, diffContent string, commitHistoryContent string, otherContext map[string]string) (string, error) {
	if m.MockError != nil {
		return "", m.MockError
	}
	return m.MockYAMLOutput, nil
}

// CountTokens implements the Adapter interface with mock data
func (m *MockAdapter) CountTokens(ctx context.Context, llmCommand string, textContent string) (int, error) {
	if m.MockError != nil {
		return 0, m.MockError
	}
	return m.MockTokenCount, nil
}

// NewMockAdapter creates a new mock adapter with sample data
func NewMockAdapter() *MockAdapter {
	return &MockAdapter{
		MockYAMLOutput: `title: "feat: implement hello world functionality"
body: |
  This PR adds a simple Hello World implementation.
  
  - Added fmt import
  - Implemented print statement
  
  Resolves: #42
changelog: "Added Hello World functionality"
release_notes:
  title: "Hello World Feature"
  body: "Users can now see a Hello World message when running the application."
`,
		MockTokenCount: 150,
	}
}