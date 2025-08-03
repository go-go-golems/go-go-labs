package cmd

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed docs/README.md
var readmeContent string

//go:embed docs/AGENT_GUIDE.md
var agentGuideContent string

//go:embed docs/QUICK_START.md
var quickStartContent string

//go:embed docs/WORKFLOW.md
var workflowContent string

//go:embed docs/SETUP.md
var setupContent string

//go:embed docs/COMMANDS.md
var commandsContent string

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Access embedded documentation",
	Long:  "View README, agent guide, and other documentation directly from the CLI",
}

var readmeCmd = &cobra.Command{
	Use:   "readme",
	Short: "Show the complete README documentation",
	Long:  "Display the full README with system overview, features, and usage examples",
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, _ := cmd.Flags().GetBool("raw")
		return displayDoc("README", readmeContent, raw)
	},
}

var agentGuideCmd = &cobra.Command{
	Use:   "agent-guide",
	Short: "Show the agent work guide",
	Long:  "Display the concise agent work guide with essential commands and workflow",
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, _ := cmd.Flags().GetBool("raw")
		return displayDoc("Agent Guide", agentGuideContent, raw)
	},
}

var quickStartCmd = &cobra.Command{
	Use:   "quick-start",
	Short: "Show quick start guide",
	Long:  "Display essential commands to get started with the system",
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, _ := cmd.Flags().GetBool("raw")
		return displayDoc("Quick Start", quickStartContent, raw)
	},
}

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Show typical agent workflow",
	Long:  "Display the step-by-step workflow for agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, _ := cmd.Flags().GetBool("raw")
		return displayDoc("Agent Workflow", workflowContent, raw)
	},
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Complete project initialization guide",
	Long:  "Step-by-step guide to initialize a new project with proper structure",
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, _ := cmd.Flags().GetBool("raw")
		return displayDoc("Project Setup Guide", setupContent, raw)
	},
}

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "Show all available commands summary",
	Long:  "Display a summary of all available commands organized by category",
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, _ := cmd.Flags().GetBool("raw")
		return displayDoc("Commands Reference", commandsContent, raw)
	},
}

func displayDoc(title, content string, raw bool) error {
	if raw {
		fmt.Print(content)
		return nil
	}

	// Format for terminal display
	fmt.Printf("═══ %s ═══\n\n", strings.ToUpper(title))

	// Add some basic formatting for better readability
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Skip the top-level title since we already showed it
		if strings.HasPrefix(line, "# ") && strings.Contains(line, title) {
			continue
		}
		fmt.Println(line)
	}

	fmt.Printf("\n═══ END %s ═══\n", strings.ToUpper(title))
	return nil
}

func init() {
	rootCmd.AddCommand(docsCmd)

	// Add subcommands
	docsCmd.AddCommand(readmeCmd)
	docsCmd.AddCommand(agentGuideCmd)
	docsCmd.AddCommand(quickStartCmd)
	docsCmd.AddCommand(workflowCmd)
	docsCmd.AddCommand(setupCmd)
	docsCmd.AddCommand(commandsCmd)

	// Add flags for raw output
	for _, cmd := range []*cobra.Command{readmeCmd, agentGuideCmd, quickStartCmd, workflowCmd, setupCmd, commandsCmd} {
		cmd.Flags().Bool("raw", false, "Output raw markdown without formatting")
	}
}
