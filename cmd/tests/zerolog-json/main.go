// https://www.phind.com/search?cache=vhy8whukmiiay54aublf9ykp

package main

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type User struct {
	Name  string
	Email string
	Age   int
}

func main() {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	//log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true})

	// Create a user object
	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	// Log various types of messages
	log.Info().Msg("This is an info message")
	log.Warn().Int("count", 3).Msg("This is a warning with a count")
	log.Error().Str("error", "file not found").Msg("An error occurred")
	log.Debug().Str("key", "value").Msg("Debug message with key-value")
	log.Info().Time("now", time.Now()).Msg("Current time")
	log.Info().Interface("user", user).Msg("User information")
}
