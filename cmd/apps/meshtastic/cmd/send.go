package cmd

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/client"
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send [message]",
	Short: "Send a text message",
	Long:  `Send a text message through the Meshtastic device.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSend,
}

func runSend(cmd *cobra.Command, args []string) error {
	message := args[0]
	log.Info().Str("port", globalConfig.Port).Str("message", message).Msg("Sending message")

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
	defer meshtasticClient.Close()

	// Connect to the device
	ctx, cancel := context.WithTimeout(context.Background(), globalConfig.Timeout)
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
}

func init() {
	rootCmd.AddCommand(sendCmd)
}
