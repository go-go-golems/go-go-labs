package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-go-golems/go-go-labs/cmd/apps/create-pull-request/pkg/application"
	"github.com/go-go-golems/go-go-labs/cmd/apps/create-pull-request/pkg/infrastructure/filesystem"
	"github.com/go-go-golems/go-go-labs/cmd/apps/create-pull-request/pkg/infrastructure/git"
	"github.com/go-go-golems/go-go-labs/cmd/apps/create-pull-request/pkg/infrastructure/github"
	"github.com/go-go-golems/go-go-labs/cmd/apps/create-pull-request/pkg/infrastructure/llm"
	"github.com/spf13/cobra"
)

var (
	// Common flags
	branch         string
	issueID        string
	title          string
	outputFile     string
	nonInteractive bool
	tui            bool

	// Diff customization flags
	diffFile        string
	fromClipboard   bool
	exclude         string
	diffContextSize int
	only            string
	noTests         bool
	noPackage       bool

	// Commit history flags
	commitsFile string
	noCommits   bool

	// LLM customization flags
	llmCommand            string
	llmStyle              string
	llmParams             []string
	codeContextFiles      []string
	additionalSystemPrompt string
	additionalUserPrompts []string
	contextFiles          []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gopr",
	Short: "Go Pull Request - A tool for creating pull requests",
	Long: `A Go tool for creating pull requests with LLM-generated descriptions.

This tool makes it easy to create well-structured pull requests by analyzing
your code changes and commit history, then using an LLM to generate a comprehensive
description.`,
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:     "create [flags] <description>",
	Short:   "Create a pull request with LLM-generated description",
	Aliases: []string{"c"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Parse all flags into the config
		config := application.CreatePullRequestConfig{
			Description:            args[0],
			Branch:                 branch,
			IssueID:                issueID,
			Title:                  title,
			OutputFile:             outputFile,
			NonInteractive:         nonInteractive,
			TUI:                    tui,
			DiffFile:               diffFile,
			FromClipboard:          fromClipboard,
			DiffExclusions:         parseCommaSeparated(exclude),
			DiffContextSize:        diffContextSize,
			DiffOnlyPaths:          parseCommaSeparated(only),
			DiffNoTests:            noTests,
			DiffNoPackage:          noPackage,
			CommitsFile:            commitsFile,
			NoCommits:              noCommits,
			LlmCommand:             llmCommand,
			LlmStyle:               llmStyle,
			LlmParams:              parseKeyValuePairs(llmParams),
			CodeContextFiles:       codeContextFiles,
			AdditionalSystemPrompt: additionalSystemPrompt,
			AdditionalUserPrompts:  additionalUserPrompts,
			ContextFiles:           contextFiles,
		}

		// Initialize the service with adapters
		service := application.NewPullRequestService(
			git.NewMockAdapter(),       // Use mock for prototype
			llm.NewPinocchioAdapter(), // Use real LLM adapter
			github.NewMockAdapter(),    // Use mock for prototype
			filesystem.NewRealAdapter(),
		)

		// Create the pull request
		prSpec, prURL, err := service.CreatePullRequest(context.Background(), config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating pull request: %v\n", err)
			os.Exit(1)
		}

		// Print result
		fmt.Printf("Pull request created successfully!\n")
		fmt.Printf("Title: %s\n", prSpec.Title)
		fmt.Printf("URL: %s\n", prURL)
		fmt.Printf("YAML saved to: %s\n", config.OutputFile)
	},
}

// getDiffCmd represents the get-diff command
var getDiffCmd = &cobra.Command{
	Use:   "get-diff [flags]",
	Short: "Get the git diff for the current branch",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize the git adapter
		gitAdapter := git.NewMockAdapter() // Use mock for prototype

		// Get the diff
		diff, err := gitAdapter.GetDiff(
			context.Background(),
			branch,
			parseCommaSeparated(exclude),
			diffContextSize,
			parseCommaSeparated(only),
			noTests,
			noPackage,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting diff: %v\n", err)
			os.Exit(1)
		}

		// Print the diff
		fmt.Println(diff)
	},
}

// createFromYamlCmd represents the create-from-yaml command
var createFromYamlCmd = &cobra.Command{
	Use:     "create-from-yaml [flags] [<yaml_file_path>]",
	Short:   "Create a pull request from a YAML file",
	Aliases: []string{"cfy"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Determine YAML file path
		yamlFilePath := outputFile // default
		if len(args) > 0 {
			yamlFilePath = args[0]
		}

		// Initialize adapters
		fsAdapter := filesystem.NewRealAdapter()
		githubAdapter := github.NewMockAdapter() // Use mock for prototype

		// Read the YAML file
		_, err := fsAdapter.ReadFile(yamlFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading YAML file: %v\n", err)
			os.Exit(1)
		}
		// In a real implementation, we would parse the YAML here

		// Parse the YAML (In a real implementation, we would parse it properly)
		// For now, we'll just simulate a successful PR creation
		prURL, err := githubAdapter.CreatePullRequest(
			context.Background(),
			"Mock Title from YAML",
			"Mock Body from YAML",
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating PR from YAML: %v\n", err)
			os.Exit(1)
		}

		// Print result
		fmt.Printf("Pull request created successfully from YAML!\n")
		fmt.Printf("URL: %s\n", prURL)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add commands to the root command
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(getDiffCmd)
	rootCmd.AddCommand(createFromYamlCmd)

	// Define flags for the create command
	createCmd.Flags().StringVar(&branch, "branch", "origin/main", "Target branch to diff against")
	createCmd.Flags().StringVar(&issueID, "issue", "", "Issue reference (e.g., number, URL)")
	createCmd.Flags().StringVar(&title, "title", "", "Suggested title for the pull request")
	createCmd.Flags().StringVar(&outputFile, "output-file", "/tmp/pr.yaml", "File to store the generated PR YAML")
	createCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Skip all interactive prompts")
	createCmd.Flags().BoolVar(&tui, "tui", false, "Launch the Terminal User Interface")

	createCmd.Flags().StringVar(&diffFile, "diff-file", "", "Override automatic diff generation by providing a specific diff file")
	createCmd.Flags().BoolVar(&fromClipboard, "from-clipboard", false, "Use diff from clipboard instead of generating it")
	createCmd.Flags().StringVar(&exclude, "exclude", "", "Files/patterns to exclude from the auto-generated diff (comma-separated)")
	createCmd.Flags().IntVar(&diffContextSize, "diff-context-size", 3, "Set context size for git diff -U<int>")
	createCmd.Flags().StringVar(&only, "only", "", "Include specific paths only in the auto-generated diff (comma-separated)")
	createCmd.Flags().BoolVar(&noTests, "no-tests", false, "Exclude common test file patterns")
	createCmd.Flags().BoolVar(&noPackage, "no-package", false, "Exclude common package manager files")

	createCmd.Flags().StringVar(&commitsFile, "commits-file", "", "Override automatic commit history gathering")
	createCmd.Flags().BoolVar(&noCommits, "no-commits", false, "Do not include commit history in the context")

	createCmd.Flags().StringVar(&llmCommand, "llm-command", "pinocchio code create-pull-request", "Command to use for LLM interaction")
	createCmd.Flags().StringVar(&llmStyle, "llm-style", "", "Predefined style for the LLM output")
	createCmd.Flags().StringSliceVar(&llmParams, "llm-param", nil, "Pass additional raw parameters directly to the LLM prompt template (key=value)")
	createCmd.Flags().StringSliceVar(&codeContextFiles, "code-context", nil, "Provide specific code files as additional context to the LLM")
	createCmd.Flags().StringVar(&additionalSystemPrompt, "additional-system-prompt", "", "Additional system prompt for the LLM")
	createCmd.Flags().StringSliceVar(&additionalUserPrompts, "additional-user-prompt", nil, "Additional user prompt content for the LLM")
	createCmd.Flags().StringSliceVar(&contextFiles, "context-files", nil, "Additional arbitrary files to provide as context to the LLM")

	// Define flags for the get-diff command
	getDiffCmd.Flags().StringVar(&branch, "branch", "origin/main", "Target branch to diff against")
	getDiffCmd.Flags().StringVar(&exclude, "exclude", "", "Files/patterns to exclude from the diff (comma-separated)")
	getDiffCmd.Flags().IntVar(&diffContextSize, "diff-context-size", 3, "Set context size for git diff -U<int>")
	getDiffCmd.Flags().StringVar(&only, "only", "", "Include specific paths only in the diff (comma-separated)")
	getDiffCmd.Flags().BoolVar(&noTests, "no-tests", false, "Exclude common test file patterns")
	getDiffCmd.Flags().BoolVar(&noPackage, "no-package", false, "Exclude common package manager files")

	// Define flags for the create-from-yaml command
	createFromYamlCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Skip interactive prompts")
}

// Helper functions

// parseCommaSeparated splits a comma-separated string into a slice
func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}

	return parts
}

// parseKeyValuePairs parses key=value pairs into a map
func parseKeyValuePairs(pairs []string) map[string]string {
	result := make(map[string]string)

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}

	return result
}