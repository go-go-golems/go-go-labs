package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/client"
	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
)

var (
	lastMessageSender  uint32
	lastMessageChannel uint32
)

// messageCmd represents the message command
var messageCmd = &cobra.Command{
	Use:   "message",
	Short: "Send and receive messages",
	Long:  `Send and receive text messages through the Meshtastic device.`,
}

// messageSendCmd represents the message send command
var messageSendCmd = &cobra.Command{
	Use:   "send [message]",
	Short: "Send a text message",
	Long:  `Send a text message to the mesh network or specific destination.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runMessageSend,
}

// messageListenCmd represents the message listen command
var messageListenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen for incoming messages",
	Long:  `Listen for incoming messages from the Meshtastic device.`,
	RunE:  runMessageListen,
}

// messageReplyCmd represents the message reply command
var messageReplyCmd = &cobra.Command{
	Use:   "reply [message]",
	Short: "Reply to the last received message",
	Long:  `Reply to the last received message.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runMessageReply,
}

// messagePrivateCmd represents the message private command
var messagePrivateCmd = &cobra.Command{
	Use:   "private [message]",
	Short: "Send a private message",
	Long:  `Send a private message to a specific destination.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runMessagePrivate,
}

var (
	destFlag     string
	channelFlag  uint32
	wantAckFlag  bool
	hopLimitFlag uint32
	timeoutFlag  time.Duration
	fromFlag     string
	jsonFlag     bool
)

func runMessageSend(cmd *cobra.Command, args []string) error {
	message := args[0]

	// Parse destination
	dest, err := parseDestination(destFlag)
	if err != nil {
		return errors.Wrap(err, "invalid destination")
	}

	log.Info().
		Str("port", globalConfig.Port).
		Str("message", message).
		Str("dest", destFlag).
		Uint32("channel", channelFlag).
		Bool("want_ack", wantAckFlag).
		Msg("Sending message")

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

	// Create message packet
	packet := &pb.MeshPacket{
		To:       dest,
		Channel:  channelFlag,
		WantAck:  wantAckFlag,
		HopLimit: hopLimitFlag,
		PayloadVariant: &pb.MeshPacket_Decoded{
			Decoded: &pb.Data{
				Portnum: pb.PortNum_TEXT_MESSAGE_APP,
				Payload: []byte(message),
			},
		},
	}

	// Send message
	err = meshtasticClient.SendMessage(packet)
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	if dest == client.BROADCAST_ADDR {
		fmt.Printf("Message sent to mesh network: %s\n", message)
	} else {
		fmt.Printf("Message sent to %s: %s\n", destFlag, message)
	}

	return nil
}

func runMessageListen(cmd *cobra.Command, args []string) error {
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
	if channelFlag > 0 {
		fmt.Printf("Filtering for channel: %d\n", channelFlag)
	}
	if fromFlag != "" {
		fmt.Printf("Filtering for sender: %s\n", fromFlag)
	}
	fmt.Println("Press Ctrl+C to stop")

	// Parse from filter
	var fromNodeID uint32
	if fromFlag != "" {
		fromNodeID, err = parseDestination(fromFlag)
		if err != nil {
			return errors.Wrap(err, "invalid from filter")
		}
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if timeoutFlag > 0 {
		var timeoutCancel context.CancelFunc
		ctx, timeoutCancel = context.WithTimeout(ctx, timeoutFlag)
		defer timeoutCancel()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Set up message handler
	messageReceived := make(chan *pb.MeshPacket, 100)
	meshtasticClient.SetOnMessage(func(packet *pb.MeshPacket) {
		// Apply filters
		if channelFlag > 0 && packet.GetChannel() != channelFlag {
			return
		}
		if fromNodeID > 0 && packet.GetFrom() != fromNodeID {
			return
		}

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
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				fmt.Printf("\nTimeout reached after %v\n", timeoutFlag)
			}
			return nil
		case <-sigChan:
			fmt.Println("\nShutting down...")
			return nil
		case packet := <-messageReceived:
			decoded := packet.GetDecoded()
			if decoded != nil && decoded.GetPortnum() == pb.PortNum_TEXT_MESSAGE_APP {
				// Update last message info for reply functionality
				lastMessageSender = packet.GetFrom()
				lastMessageChannel = packet.GetChannel()

				timestamp := time.Now().Format("2006-01-02 15:04:05")
				senderInfo := fmt.Sprintf("Node %d", packet.GetFrom())

				// Look up sender name from node info
				if nodes := meshtasticClient.GetNodes(); nodes != nil {
					if node, exists := nodes[packet.GetFrom()]; exists && node.GetUser() != nil {
						senderInfo = fmt.Sprintf("%s (%s)", node.GetUser().GetLongName(), senderInfo)
					}
				}

				channelInfo := ""
				if packet.GetChannel() > 0 {
					channelInfo = fmt.Sprintf(" [Ch%d]", packet.GetChannel())
				}

				if packet.GetTo() != client.BROADCAST_ADDR {
					channelInfo += " [Private]"
				}

				if jsonFlag {
					fmt.Printf(`{"timestamp": "%s", "from": %d, "channel": %d, "message": "%s"}`+"\n",
						timestamp, packet.GetFrom(), packet.GetChannel(), string(decoded.GetPayload()))
				} else {
					fmt.Printf("[%s] %s%s: %s\n", timestamp, senderInfo, channelInfo, string(decoded.GetPayload()))
				}
			}
		}
	}
}

func runMessageReply(cmd *cobra.Command, args []string) error {
	message := args[0]

	if lastMessageSender == 0 {
		return errors.New("no previous message to reply to")
	}

	log.Info().
		Str("port", globalConfig.Port).
		Str("message", message).
		Uint32("to", lastMessageSender).
		Uint32("channel", lastMessageChannel).
		Msg("Replying to message")

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

	// Create reply packet
	packet := &pb.MeshPacket{
		To:      lastMessageSender,
		Channel: lastMessageChannel,
		WantAck: wantAckFlag,
		PayloadVariant: &pb.MeshPacket_Decoded{
			Decoded: &pb.Data{
				Portnum: pb.PortNum_TEXT_MESSAGE_APP,
				Payload: []byte(message),
			},
		},
	}

	// Send reply
	err = meshtasticClient.SendMessage(packet)
	if err != nil {
		return errors.Wrap(err, "failed to send reply")
	}

	fmt.Printf("Reply sent to node %d: %s\n", lastMessageSender, message)
	return nil
}

func runMessagePrivate(cmd *cobra.Command, args []string) error {
	message := args[0]

	if destFlag == "" {
		return errors.New("destination is required for private messages (use --dest flag)")
	}

	// Parse destination
	dest, err := parseDestination(destFlag)
	if err != nil {
		return errors.Wrap(err, "invalid destination")
	}

	if dest == client.BROADCAST_ADDR {
		return errors.New("cannot send private message to broadcast address")
	}

	log.Info().
		Str("port", globalConfig.Port).
		Str("message", message).
		Str("dest", destFlag).
		Msg("Sending private message")

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

	// Create private message packet
	packet := &pb.MeshPacket{
		To:      dest,
		Channel: 0, // Private messages go on channel 0
		WantAck: wantAckFlag,
		PayloadVariant: &pb.MeshPacket_Decoded{
			Decoded: &pb.Data{
				Portnum: pb.PortNum_TEXT_MESSAGE_APP,
				Payload: []byte(message),
			},
		},
	}

	// Send private message
	err = meshtasticClient.SendMessage(packet)
	if err != nil {
		return errors.Wrap(err, "failed to send private message")
	}

	fmt.Printf("Private message sent to %s: %s\n", destFlag, message)
	return nil
}

// parseDestination parses various destination formats
func parseDestination(dest string) (uint32, error) {
	if dest == "" {
		return client.BROADCAST_ADDR, nil
	}

	// Handle broadcast
	if strings.ToLower(dest) == "broadcast" || dest == "all" {
		return client.BROADCAST_ADDR, nil
	}

	// Handle hex node ID format (!a4c138f4)
	if strings.HasPrefix(dest, "!") {
		hexStr := dest[1:]
		if len(hexStr) != 8 {
			return 0, errors.New("node ID must be 8 hex characters after !")
		}
		nodeID, err := strconv.ParseUint(hexStr, 16, 32)
		if err != nil {
			return 0, errors.Wrap(err, "invalid hex node ID")
		}
		return uint32(nodeID), nil
	}

	// Handle decimal node ID
	if nodeID, err := strconv.ParseUint(dest, 10, 32); err == nil {
		return uint32(nodeID), nil
	}

	// Handle hex without ! prefix
	if nodeID, err := strconv.ParseUint(dest, 16, 32); err == nil {
		return uint32(nodeID), nil
	}

	// TODO: Handle node name lookup
	return 0, errors.New("invalid destination format. Use node ID (decimal/hex) or !hex format")
}

func init() {
	// Add flags for message send
	messageSendCmd.Flags().StringVar(&destFlag, "dest", "", "Destination node ID, name, or number")
	messageSendCmd.Flags().Uint32Var(&channelFlag, "channel", 0, "Channel to send on (default: 0)")
	messageSendCmd.Flags().BoolVar(&wantAckFlag, "want-ack", false, "Request acknowledgment")
	messageSendCmd.Flags().Uint32Var(&hopLimitFlag, "hop-limit", 3, "Maximum hop limit")

	// Add flags for message listen
	messageListenCmd.Flags().Uint32Var(&channelFlag, "channel", 0, "Listen on specific channel (0 for all)")
	messageListenCmd.Flags().StringVar(&fromFlag, "from", "", "Only show messages from specific node")
	messageListenCmd.Flags().DurationVar(&timeoutFlag, "timeout", 0, "Listen timeout (0 for infinite)")
	messageListenCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output messages in JSON format")

	// Add flags for message reply
	messageReplyCmd.Flags().BoolVar(&wantAckFlag, "want-ack", false, "Request acknowledgment")

	// Add flags for message private
	messagePrivateCmd.Flags().StringVar(&destFlag, "dest", "", "Destination node ID, name, or number (required)")
	messagePrivateCmd.Flags().BoolVar(&wantAckFlag, "want-ack", false, "Request acknowledgment")
	messagePrivateCmd.MarkFlagRequired("dest")

	// Add subcommands
	messageCmd.AddCommand(messageSendCmd)
	messageCmd.AddCommand(messageListenCmd)
	messageCmd.AddCommand(messageReplyCmd)
	messageCmd.AddCommand(messagePrivateCmd)

	// Add to root command
	rootCmd.AddCommand(messageCmd)
}
