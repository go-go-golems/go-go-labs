package main

import (
	"os"

	"github.com/go-go-golems/go-go-labs/cmd/apps/anki/services"
	"github.com/go-go-golems/go-go-labs/cmd/apps/anki/views"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	ankiService *services.AnkiService
	logLevel    string
)

var rootCmd = &cobra.Command{
	Use:   "anki-viewer",
	Short: "A simple web viewer for Anki decks",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Setup logging
		level, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			return err
		}

		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).Level(level)

		// Initialize services
		ankiService = services.NewAnkiService(log.Logger)

		e := echo.New()
		e.HideBanner = true

		// Middleware
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())

		// Routes
		e.GET("/", handleHome)
		e.GET("/decks", handleListDecks)
		e.GET("/decks/:deckName/cards", handleListCards)
		e.GET("/models", handleListModels)

		log.Info().Msg("Starting server on :8080")
		return e.Start(":8080")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start application")
	}
}

// Temporary handlers until we move them to their own package
func handleHome(c echo.Context) error {
	return c.Redirect(302, "/decks")
}

func handleListDecks(c echo.Context) error {
	log.Debug().Msg("Fetching decks list")
	decks, err := ankiService.GetDecks()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get decks")
		return err
	}

	log.Debug().Strs("decks", decks).Msg("Retrieved decks")
	component := views.DecksList(decks)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func handleListCards(c echo.Context) error {
	deckName := c.Param("deckName")
	log.Debug().Str("deck", deckName).Msg("Fetching cards for deck")

	cards, err := ankiService.GetCardsInDeck(deckName)
	if err != nil {
		log.Error().Err(err).Str("deck", deckName).Msg("Failed to get cards")
		return err
	}

	log.Debug().Str("deck", deckName).Int("cardCount", len(cards)).Msg("Retrieved cards")
	component := views.CardsList(deckName, cards)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func handleListModels(c echo.Context) error {
	log.Debug().Msg("Fetching models list")
	models, err := ankiService.GetModels()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get models")
		return err
	}

	log.Debug().Int("modelCount", len(models)).Msg("Retrieved models")
	component := views.ModelsList(models)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
