package main

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "mail-rules",
	Short: "Mail rules management application",
	Long:  `A command line tool for managing mail rules and interacting with email accounts.`,
}

func Execute() error {
	return RootCmd.Execute()
}
