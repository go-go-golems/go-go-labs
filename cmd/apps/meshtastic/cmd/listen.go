package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/client"
	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
)

// listenCmd represents the listen command
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen for incoming messages",
	Long:  `Listen for incoming messages from the Meshtastic device.`,
	RunE:  runListen,
}

func runListen(cmd *cobra.Command, args []string) error {
	log.Info().Str("port", globalConfig.Port).Msg("Starting to listen for messages")

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
	connectCtx, connectCancel := context.WithTimeout(context.Background(), globalConfig.Timeout)
	defer connectCancel()

	if err := meshtasticClient.Connect(connectCtx); err != nil {
		return errors.Wrap(err, "failed to connect to device")
	}
	defer meshtasticClient.Disconnect()

	// Check connection status
	if !meshtasticClient.IsConnected() {
		return errors.New("device not connected")
	}

	fmt.Printf("Listening for messages on port: %s\n", globalConfig.Port)
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
}

func init() {
	rootCmd.AddCommand(listenCmd)
}
