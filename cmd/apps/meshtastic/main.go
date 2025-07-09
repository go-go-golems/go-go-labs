package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/cmd"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/client"
	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/ui/model"
)

var (
	port        string
	logLevel    string
	timeout     time.Duration
	debugSerial bool
	hexDump     bool
)

var rootCmd = &cobra.Command{
	Use:   "meshtastic-tui",
	Short: "A TUI for Meshtastic devices",
	Long:  `A terminal user interface for interacting with Meshtastic devices via serial connection.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return setupLogging(logLevel)
	},
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show device information",
	Long:  `Display information about the connected Meshtastic device.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info().Str("port", port).Msg("Getting device info")

		// Create robust client
		config := &client.Config{
			DevicePath:  port,
			Timeout:     timeout,
			DebugSerial: debugSerial,
			HexDump:     hexDump,
		}

		meshtasticClient, err := client.NewRobustMeshtasticClient(config)
		if err != nil {
			return errors.Wrap(err, "failed to create robust client")
		}
		defer meshtasticClient.Close()

		// Connect to the device
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := meshtasticClient.Connect(ctx); err != nil {
			return errors.Wrap(err, "failed to connect to device")
		}
		defer meshtasticClient.Disconnect()

		// Check connection status
		if !meshtasticClient.IsConnected() {
			fmt.Printf("Device not connected on port: %s\n", port)
			return nil
		}

		fmt.Printf("Device info for port: %s\n", port)
		fmt.Println("Status: Connected")

		// Get device info
		myInfo := meshtasticClient.GetMyInfo()
		if myInfo != nil {
			fmt.Printf("Node ID: %d\n", myInfo.GetMyNodeNum())
			fmt.Printf("Reboot Count: %d\n", myInfo.GetRebootCount())
			fmt.Printf("Min App Version: %d\n", myInfo.GetMinAppVersion())
			fmt.Printf("PIO Environment: %s\n", myInfo.GetPioEnv())
		}

		// Get node information
		nodes := meshtasticClient.GetNodes()
		if len(nodes) > 0 {
			fmt.Printf("\nKnown nodes (%d):\n", len(nodes))
			for nodeNum, node := range nodes {
				user := node.GetUser()
				fmt.Printf("  %d: %s (%s)\n", nodeNum, user.GetLongName(), user.GetShortName())
			}
		}

		// Get channels
		channels := meshtasticClient.GetChannels()
		if len(channels) > 0 {
			fmt.Printf("\nChannels (%d):\n", len(channels))
			for idx, channel := range channels {
				settings := channel.GetSettings()
				fmt.Printf("  %d: %s\n", idx, settings.GetName())
			}
		}

		return nil
	},
}

var sendCmd = &cobra.Command{
	Use:   "send [message]",
	Short: "Send a text message",
	Long:  `Send a text message through the Meshtastic device.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		message := args[0]
		log.Info().Str("port", port).Str("message", message).Msg("Sending message")

		// Create robust client
		config := &client.Config{
			DevicePath:  port,
			Timeout:     timeout,
			DebugSerial: debugSerial,
			HexDump:     hexDump,
		}

		meshtasticClient, err := client.NewRobustMeshtasticClient(config)
		if err != nil {
			return errors.Wrap(err, "failed to create robust client")
		}
		defer meshtasticClient.Close()

		// Connect to the device
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := meshtasticClient.Connect(ctx); err != nil {
			return errors.Wrap(err, "failed to connect to device")
		}
		defer meshtasticClient.Disconnect()

		// Check connection status
		if !meshtasticClient.IsConnected() {
			return errors.New("device not connected")
		}

		// Send message (broadcast to all)
		err = meshtasticClient.SendText(message, client.BROADCAST_ADDR)
		if err != nil {
			return errors.Wrap(err, "failed to send message")
		}

		fmt.Printf("Message sent successfully: %s\n", message)
		return nil
	},
}

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen for incoming messages",
	Long:  `Listen for incoming messages from the Meshtastic device.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info().Str("port", port).Msg("Starting to listen for messages")

		// Create robust client
		config := &client.Config{
			DevicePath:  port,
			Timeout:     timeout,
			DebugSerial: debugSerial,
			HexDump:     hexDump,
		}

		meshtasticClient, err := client.NewRobustMeshtasticClient(config)
		if err != nil {
			return errors.Wrap(err, "failed to create robust client")
		}
		defer meshtasticClient.Close()

		// Connect to the device
		connectCtx, connectCancel := context.WithTimeout(context.Background(), timeout)
		defer connectCancel()

		if err := meshtasticClient.Connect(connectCtx); err != nil {
			return errors.Wrap(err, "failed to connect to device")
		}
		defer meshtasticClient.Disconnect()

		// Check connection status
		if !meshtasticClient.IsConnected() {
			return errors.New("device not connected")
		}

		fmt.Printf("Listening for messages on port: %s\n", port)
		fmt.Println("Press Ctrl+C to stop")

		// Set up signal handling for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Set up message handler
		messageReceived := make(chan *pb.MeshPacket, 100)
		meshtasticClient.SetOnMessage(func(packet *pb.MeshPacket) {
			select {
			case messageReceived <- packet:
			default:
				log.Warn().Msg("Message buffer full, dropping message")
			}
		})

		// Listen for messages
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-sigChan:
				fmt.Println("\nShutting down...")
				return nil
			case packet := <-messageReceived:
				decoded := packet.GetDecoded()
				if decoded != nil && decoded.GetPortnum() == pb.PortNum_TEXT_MESSAGE_APP {
					timestamp := time.Now().Format("15:04:05")
					fmt.Printf("[%s] From %d: %s\n", timestamp, packet.GetFrom(), string(decoded.GetPayload()))
				}
			}
		}
	},
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the TUI interface",
	Long:  `Launch the terminal user interface for interactive Meshtastic device management.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info().Str("port", port).Msg("Starting TUI")

		// Create robust client
		config := &client.Config{
			DevicePath:  port,
			Timeout:     timeout,
			DebugSerial: debugSerial,
			HexDump:     hexDump,
		}

		meshtasticClient, err := client.NewRobustMeshtasticClient(config)
		if err != nil {
			return errors.Wrap(err, "failed to create robust client")
		}

		// Connect to the device (non-blocking for TUI)
		connectCtx, connectCancel := context.WithTimeout(context.Background(), timeout)
		defer connectCancel()

		if err := meshtasticClient.Connect(connectCtx); err != nil {
			log.Warn().Str("port", port).Err(err).Msg("Failed to connect to device, but launching TUI anyway")
		}

		// Start heartbeat for robust connection
		meshtasticClient.StartHeartbeat()

		// Launch TUI (client will be cleaned up by the TUI model)
		return launchTUI(meshtasticClient)
	},
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

func setupLogging(level string) error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Configure console writer for human-readable output
	output := zerolog.ConsoleWriter{Out: os.Stderr}
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	// Set log level
	switch strings.ToLower(level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		return errors.Errorf("invalid log level: %s", level)
	}

	return nil
}

func init() {
	// Add persistent flags
	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "/dev/ttyUSB0", "Serial port for Meshtastic device")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 10*time.Second, "Connection timeout")
	rootCmd.PersistentFlags().BoolVar(&debugSerial, "debug-serial", false, "Enable verbose serial communication logging")
	rootCmd.PersistentFlags().BoolVar(&hexDump, "hex-dump", false, "Enable hex dump logging of raw serial data")

	// Add subcommands
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(listenCmd)
	rootCmd.AddCommand(tuiCmd)
}

func main() {
	cmd.Execute()
}
