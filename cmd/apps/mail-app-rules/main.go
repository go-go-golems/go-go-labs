package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-labs/cmd/apps/mail-app-rules/commands"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mail-app-rules",
		Short: "Process mail rules on an IMAP server",
	}

	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	mailRulesCmd, err := commands.NewMailRulesCommand()
	if err != nil {
		fmt.Printf("Error creating mail rules command: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(mailRulesCmd)
	if err != nil {
		fmt.Printf("Error building Cobra command: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(cobraCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
