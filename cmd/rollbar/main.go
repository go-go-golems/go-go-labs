package main

import "github.com/spf13/cobra"

func main() {
	LoadConfig()
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

var rootCmd = &cobra.Command{
	Use:   "rollbar",
	Short: "Rollbar CLI",
}
