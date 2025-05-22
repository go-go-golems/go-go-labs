package application

import (
	"context"
	"strings"

	"github.com/go-go-golems/go-go-labs/cmd/apps/create-pull-request/pkg/domain"
	"github.com/go-go-golems/go-go-labs/cmd/apps/create-pull-request/pkg/infrastructure/filesystem"
	"github.com/go-go-golems/go-go-labs/cmd/apps/create-pull-request/pkg/infrastructure/git"
	"github.com/go-go-golems/go-go-labs/cmd/apps/create-pull-request/pkg/infrastructure/github"
	"github.com/go-go-golems/go-go-labs/cmd/apps/create-pull-request/pkg/infrastructure/llm"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// CreatePullRequestConfig holds all configuration options for PR creation
type CreatePullRequestConfig struct {
	Description           string
	Branch                string
	IssueID               string
	Title                 string
	OutputFile            string
	NonInteractive        bool
	TUI                   bool

	DiffFile              string
	FromClipboard         bool
	DiffExclusions        []string
	DiffContextSize       int
	DiffOnlyPaths         []string
	DiffNoTests           bool
	DiffNoPackage         bool

	CommitsFile           string
	NoCommits             bool

	LlmCommand            string
	LlmStyle              string
	LlmParams             map[string]string
	CodeContextFiles      []string
	AdditionalSystemPrompt string
	AdditionalUserPrompts []string
	ContextFiles          []string
}

// PullRequestService orchestrates the creation of pull requests
type PullRequestService struct {
	gitAdapter       git.Adapter
	llmAdapter       llm.Adapter
	githubCliAdapter github.Adapter
	fsAdapter        filesystem.Adapter
}

// NewPullRequestService creates a new service with the given adapters
func NewPullRequestService(
	gitAdapter git.Adapter,
	llmAdapter llm.Adapter,
	githubCliAdapter github.Adapter,
	fsAdapter filesystem.Adapter,
) *PullRequestService {
	return &PullRequestService{
		gitAdapter:       gitAdapter,
		llmAdapter:       llmAdapter,
		githubCliAdapter: githubCliAdapter,
		fsAdapter:        fsAdapter,
	}
}

// CreatePullRequest orchestrates the PR creation process
func (s *PullRequestService) CreatePullRequest(
	ctx context.Context,
	config CreatePullRequestConfig,
) (*domain.PullRequestSpec, string, error) {
	var diffContent string
	var commitMessages []string
	var issueTitle, issueBody string
	var err error

	// 1. Get Diff
	if config.FromClipboard {
		diffContent, err = s.gitAdapter.GetDiffFromClipboard(ctx)
	} else if config.DiffFile != "" {
		diffContentBytes, err := s.fsAdapter.ReadFile(config.DiffFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to read diff file")
		}
		diffContent = string(diffContentBytes)
	} else {
		// Auto-generate diff
		err = s.gitAdapter.FetchOrigin(ctx)
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to fetch from origin")
		}
		diffContent, err = s.gitAdapter.GetDiff(
			ctx,
			config.Branch,
			config.DiffExclusions,
			config.DiffContextSize,
			config.DiffOnlyPaths,
			config.DiffNoTests,
			config.DiffNoPackage,
		)
	}
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to get diff")
	}

	// 2. Get Commit History
	var commitHistoryStr string
	if !config.NoCommits {
		if config.CommitsFile != "" {
			commitsContentBytes, err := s.fsAdapter.ReadFile(config.CommitsFile)
			if err != nil {
				return nil, "", errors.Wrap(err, "failed to read commits file")
			}
			commitMessages = parseCommitMessages(string(commitsContentBytes))
		} else {
			commitMessages, err = s.gitAdapter.GetCommitHistory(ctx, config.Branch, "HEAD")
			if err != nil {
				return nil, "", errors.Wrap(err, "failed to get commit history")
			}
		}
		commitHistoryStr = joinCommitMessages(commitMessages)
	}

	// 3. Get Issue Details (if issue ID provided)
	otherContext := make(map[string]string)
	if config.IssueID != "" {
		issueTitle, issueBody, err = s.githubCliAdapter.GetIssueDetails(ctx, config.IssueID)
		if err != nil {
			// Log warning but proceed
			// In a real implementation, we would log this
		} else {
			otherContext["issue_title"] = issueTitle
			otherContext["issue_body"] = issueBody
		}
	}

	// 4. User interaction for diff (if not non-interactive and not TUI)
	if !config.NonInteractive && !config.TUI {
		// Show token count - in a real implementation, we would use these
		_, _ = s.llmAdapter.CountTokens(ctx, config.LlmCommand, diffContent)
		_, _ = s.llmAdapter.CountTokens(ctx, config.LlmCommand, commitHistoryStr)

		// In a real implementation, we would show these counts to the user
		// and offer options to view/edit the diff
	}

	// 5. Prepare LLM prompt parameters
	llmPromptParams := map[string]interface{}{
		"description":           config.Description,
		"user_title_suggestion": config.Title,
		"llm_style":            config.LlmStyle,
		"raw_params":           config.LlmParams,
		"additional_system_prompt": config.AdditionalSystemPrompt,
		"additional_user_prompts":  config.AdditionalUserPrompts,
	}

	// Add content from context files
	for _, file := range config.CodeContextFiles {
		content, err := s.fsAdapter.ReadFile(file)
		if err == nil {
			otherContext[file] = string(content)
		}
	}

	for _, file := range config.ContextFiles {
		content, err := s.fsAdapter.ReadFile(file)
		if err == nil {
			otherContext[file] = string(content)
		}
	}

	// 6. Call LLM
	rawYaml, err := s.llmAdapter.GeneratePullRequestSpec(
		ctx,
		config.LlmCommand,
		llmPromptParams,
		diffContent,
		commitHistoryStr,
		otherContext,
	)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to generate PR description with LLM")
	}

	err = s.fsAdapter.WriteFile(config.OutputFile, []byte(rawYaml))
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to write output YAML")
	}

	// 7. Parse YAML to domain.PullRequestSpec
	var prSpec domain.PullRequestSpec
	err = yaml.Unmarshal([]byte(rawYaml), &prSpec)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to parse YAML output")
	}

	currentPRTitle := prSpec.Title
	currentPRBody := prSpec.Body

	// 8. User interaction for YAML (if not non-interactive and not TUI)
	if !config.NonInteractive && !config.TUI {
		// In a real implementation, we would prompt the user to review and edit the YAML

		// For now, we'll just use the parsed YAML directly
	}

	// 9. Create PR on GitHub
	prURL, err := s.githubCliAdapter.CreatePullRequest(ctx, currentPRTitle, currentPRBody)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to create PR on GitHub")
	}

	return &prSpec, prURL, nil
}

// Helper functions

// parseCommitMessages splits commit messages by line
func parseCommitMessages(content string) []string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}

	return result
}

// joinCommitMessages joins commit messages into a single string
func joinCommitMessages(messages []string) string {
	return strings.Join(messages, "\n")
}