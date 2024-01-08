package main

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

import (
	"github.com/spf13/viper"
)

func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("$HOME/.rollbar")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("Error reading config file")
		// Handle error
	}
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List occurrences in a Rollbar project",
	Run: func(cmd *cobra.Command, args []string) {
		ListOccurrences()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
