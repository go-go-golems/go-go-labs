package main

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

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

	getTicketsCommand, err := NewGetTicketsCommand()
	cobra.CheckErr(err)
	cobraCommand, err := cli.BuildCobraCommandFromGlazeCommand(getTicketsCommand)
	cobra.CheckErr(err)
	rootCmd.AddCommand(cobraCommand)

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
