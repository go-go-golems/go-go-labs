package main

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

// TODO(manuel, 2023-10-04) Write proper documentation, add a select ticket command or something
// ways of using
// ❯ jq '[.[] | select(.created_at < "2020-01-01T00:00:00Z")]' /ttmp/backup/tickets.json > /ttmp/tickets-2019.json

type ZendeskConfig struct {
	Domain   string
	Email    string
	Password string
	ApiToken string
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "zendesk",
		Short: "Zendesk fetcher",
	}

	helpSystem := help.NewHelpSystem()

	helpSystem.SetupCobraRootCommand(rootCmd)

	getTicketsCommand, err := NewGetTicketsCommand()
	cobra.CheckErr(err)
	cobraCommand, err := cli.BuildCobraCommandFromGlazeCommand(getTicketsCommand)
	cobra.CheckErr(err)
	rootCmd.AddCommand(cobraCommand)

	deleteTicketsCommand, err := NewDeleteTicketsCommand()
	cobra.CheckErr(err)
	cobraCommand, err = cli.BuildCobraCommandFromWriterCommand(deleteTicketsCommand)
	cobra.CheckErr(err)
	rootCmd.AddCommand(cobraCommand)

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
