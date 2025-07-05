package cmd

import (
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/pkg/ui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea"
)

func NewStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the film development calculator TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().Msg("Starting film development calculator TUI")
			
			model := ui.NewAppModel()
			p := tea.NewProgram(model, tea.WithAltScreen())
			
			if _, err := p.Run(); err != nil {
				log.Error().Err(err).Msg("Failed to run TUI")
				return err
			}
			
			return nil
		},
	}
	
	return cmd
}
