package main

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"

	"github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

type CreatePRCommand struct {
	*cmds.CommandDescription
}

type CreatePRSettings struct {
	Description  string               `glazed.parameter:"description"`
	BaseBranch   string               `glazed.parameter:"base-branch"`
	Branch       string               `glazed.parameter:"branch"`
	ExcludeFiles []string             `glazed.parameter:"exclude"`
	LongContext  bool                 `glazed.parameter:"long"`
	ShortContext bool                 `glazed.parameter:"short"`
	IncludePaths string               `glazed.parameter:"only"`
	NoTests      bool                 `glazed.parameter:"no-tests"`
	NoPackage    bool                 `glazed.parameter:"no-package"`
	Issue        string               `glazed.parameter:"issue"`
	IssueFile    *parameters.FileData `glazed.parameter:"issue-file"`
}

func NewCreatePRCommand() (*CreatePRCommand, error) {
	return &CreatePRCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create-pr",
			cmds.WithShort("Create a pull request"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"description",
					parameters.ParameterTypeString,
					parameters.WithRequired(true),
					parameters.WithHelp("Description of the pull request"),
				),
				parameters.NewParameterDefinition(
					"base-branch",
					parameters.ParameterTypeString,
					parameters.WithHelp("Base branch name"),
					parameters.WithDefault("origin/main"),
				),
				parameters.NewParameterDefinition(
					"branch",
					parameters.ParameterTypeString,
					parameters.WithHelp("Branch to compare"),
				),
				parameters.NewParameterDefinition(
					"exclude",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Files to exclude (comma-separated)"),
				),
				parameters.NewParameterDefinition(
					"long",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Use long context (10 lines)"),
				),
				parameters.NewParameterDefinition(
					"short",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Use short context (1 line)"),
				),
				parameters.NewParameterDefinition(
					"only",
					parameters.ParameterTypeString,
					parameters.WithHelp("Only include specified paths"),
				),
				parameters.NewParameterDefinition(
					"no-tests",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Exclude test files"),
				),
				parameters.NewParameterDefinition(
					"no-package",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Exclude package files"),
				),
				parameters.NewParameterDefinition(
					"issue",
					parameters.ParameterTypeString,
					parameters.WithHelp("GitHub issue number or description"),
				),
				parameters.NewParameterDefinition(
					"issue-file",
					parameters.ParameterTypeFile,
					parameters.WithHelp("File containing GitHub issue description"),
				),
			),
		),
	}, nil
}

func (c *CreatePRCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	s := &CreatePRSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Create temporary files for diff and issue
	diffFile, err := os.CreateTemp("", "diff")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(diffFile.Name())

	issueFile, err := os.CreateTemp("", "issue")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(issueFile.Name())

	// Write issue content to the temporary file
	if s.Issue != "" {
		if err := os.WriteFile(issueFile.Name(), []byte(s.Issue), 0644); err != nil {
			return fmt.Errorf("failed to write issue to file: %w", err)
		}
	} else if s.IssueFile != nil {
		if err := os.WriteFile(issueFile.Name(), []byte(s.IssueFile.Content), 0644); err != nil {
			return fmt.Errorf("failed to write issue file content: %w", err)
		}
	}

	// Run git-diff.sh
	if err := runGitDiffSh(s, diffFile.Name()); err != nil {
		return err
	}

	// Run the pinocchio command
	if err := runPinocchio(s.Description, diffFile.Name(), issueFile.Name()); err != nil {
		return err
	}

	return nil
}

func runGitDiffSh(s *CreatePRSettings, outputFile string) error {
	args := []string{"get", "git-diff.sh", "--"}

	if s.Branch != "" {
		args = append(args, "-b", s.Branch)
	} else {
		args = append(args, "-b", s.BaseBranch)
	}

	if len(s.ExcludeFiles) > 0 {
		args = append(args, "-e", fmt.Sprintf("%s", s.ExcludeFiles))
	}

	if s.LongContext {
		args = append(args, "-l")
	} else if s.ShortContext {
		args = append(args, "-s")
	}

	if s.IncludePaths != "" {
		args = append(args, "-o", s.IncludePaths)
	}

	if s.NoTests {
		args = append(args, "--no-tests")
	}

	if s.NoPackage {
		args = append(args, "--no-package")
	}

	cmd := exec.Command("prompto", args...)
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func(outFile *os.File) {
		_ = outFile.Close()
	}(outFile)
	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	log.Info().Msgf("Running command: %s", cmd.String())

	return cmd.Run()
}

func runPinocchio(description, diffFile, issueFile string) error {
	cmd := exec.Command("pinocchio", "code", "create-pull-request",
		"--diff", diffFile,
		"--description", description,
		"--issue", issueFile,
		"--interactive=false",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var rootCmd = &cobra.Command{
	Use:   "pr-creator",
	Short: "Create pull requests",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := pkg.InitLogger()
		cobra.CheckErr(err)
	},
}

func initRootCmd() (*help.HelpSystem, error) {
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	err := pkg.InitViper("pr-creator", rootCmd)
	if err != nil {
		return nil, err
	}

	err = pkg.InitLogger()
	if err != nil {
		return nil, err
	}

	return helpSystem, nil
}

func registerCommands(helpSystem *help.HelpSystem) error {
	createPRCommand, err := NewCreatePRCommand()
	if err != nil {
		return err
	}
	cobraCreatePRCommand, err := cli.BuildCobraCommandFromCommand(createPRCommand)
	if err != nil {
		return err
	}
	rootCmd.AddCommand(cobraCreatePRCommand)

	return nil
}

func main() {
	helpSystem, err := initRootCmd()
	cobra.CheckErr(err)

	err = registerCommands(helpSystem)
	cobra.CheckErr(err)

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
