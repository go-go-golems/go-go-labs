package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/client"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/ui/model"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the TUI interface",
	Long:  `Launch the terminal user interface for interactive Meshtastic device management.`,
	RunE:  runTUI,
}

func runTUI(cmd *cobra.Command, args []string) error {
	log.Info().Str("port", globalConfig.Port).Msg("Starting TUI")

	// Create robust client
	config := &client.Config{
		DevicePath:  globalConfig.Port,
		Timeout:     globalConfig.Timeout,
		DebugSerial: globalConfig.DebugSerial,
		HexDump:     globalConfig.HexDump,
	}

	meshtasticClient, err := client.NewRobustMeshtasticClient(config)
	if err != nil {
		return errors.Wrap(err, "failed to create robust client")
	}

	// Connect to the device (non-blocking for TUI)
	connectCtx, connectCancel := context.WithTimeout(context.Background(), globalConfig.Timeout)
	defer connectCancel()

	if err := meshtasticClient.Connect(connectCtx); err != nil {
		log.Warn().Str("port", globalConfig.Port).Err(err).Msg("Failed to connect to device, but launching TUI anyway")
	}

	// Start heartbeat for robust connection
	meshtasticClient.StartHeartbeat()

	// Launch TUI (client will be cleaned up by the TUI model)
	return launchTUI(meshtasticClient)
}

func launchTUI(client *client.RobustMeshtasticClient) error {
	// Create the root model and inject the client
	m := model.NewRootModelWithClient(client)

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the program with proper options
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithContext(ctx),
	)

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run the program in a goroutine
	var runErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if err := m.Cleanup(); err != nil {
				log.Error().Err(err).Msg("Failed to cleanup TUI")
			}
		}()
		if _, err := p.Run(); err != nil {
			runErr = errors.Wrap(err, "failed to run TUI")
		}
	}()

	// Wait for either completion or signal
	select {
	case <-done:
		return runErr
	case <-sigChan:
		log.Info().Msg("Received shutdown signal, cleaning up...")
		cancel()
		p.Quit()
		<-done
		return runErr
	}
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
