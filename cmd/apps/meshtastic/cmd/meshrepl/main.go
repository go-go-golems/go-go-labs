package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/client"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/deviceadapter"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/events"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/meshbus"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/serial/discovery"
)

var (
	// Global CLI arguments
	args struct {
		LogLevel   string `help:"Log level (debug, info, warn, error)" default:"info"`
		DevicePath string `help:"Device path (auto-discover if not specified)" short:"d"`
		Timeout    int    `help:"Connection timeout in seconds" default:"30"`
	}

	// Global state
	bus           *meshbus.Bus
	adapter       *deviceadapter.DeviceAdapter
	eventListener *EventListener
	handlerAdded  bool // Track if we've added the handler

	// Active connections tracking
	activeConnections = make(map[string]bool)
	connectionsMu     sync.RWMutex

	// Colors for output
	colorSuccess = color.New(color.FgGreen).SprintFunc()
	colorError   = color.New(color.FgRed).SprintFunc()
	colorWarning = color.New(color.FgYellow).SprintFunc()
	colorInfo    = color.New(color.FgCyan).SprintFunc()
	colorDevice  = color.New(color.FgMagenta).SprintFunc()
	colorCommand = color.New(color.FgBlue).SprintFunc()
)

// EventListener handles real-time event display
type EventListener struct {
	bus       *meshbus.Bus
	ctx       context.Context
	cancel    context.CancelFunc
	listening bool
}

// NewEventListener creates a new event listener
func NewEventListener(bus *meshbus.Bus) *EventListener {
	ctx, cancel := context.WithCancel(context.Background())
	return &EventListener{
		bus:    bus,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts listening for events
func (el *EventListener) Start() error {
	if el.listening {
		return nil
	}

	// Only add handler once since Watermill doesn't support removing handlers
	if !handlerAdded {
		if err := el.bus.AddHandler(
			"repl_event_listener",
			"broadcast.*",
			el.handleEvent,
		); err != nil {
			return errors.Wrap(err, "failed to add event handler")
		}
		handlerAdded = true
	}

	el.listening = true
	return nil
}

// Stop stops listening for events
func (el *EventListener) Stop() {
	if !el.listening {
		return
	}

	// Note: We can't actually remove handlers from Watermill once added
	// We just mark as not listening so handleEvent will ignore messages
	el.listening = false
}

// handleEvent handles incoming events
func (el *EventListener) handleEvent(msg *message.Message) error {
	// Skip if not listening
	if !el.listening {
		return nil
	}

	// Parse envelope
	var envelope events.Envelope
	if err := json.Unmarshal(msg.Payload, &envelope); err != nil {
		return errors.Wrap(err, "failed to parse envelope")
	}

	// Format and display event
	el.displayEvent(&envelope)

	return nil
}

// displayEvent displays an event to the user
func (el *EventListener) displayEvent(envelope *events.Envelope) {
	timestamp := envelope.Timestamp.Format("15:04:05")
	deviceID := envelope.DeviceID
	if deviceID == "" {
		deviceID = "unknown"
	}

	switch envelope.Type {
	case events.EventDeviceConnected:
		var event events.DeviceConnectedEvent
		if err := envelope.GetData(&event); err == nil {
			fmt.Printf("[%s] %s %s connected to %s\n",
				timestamp,
				colorSuccess("✓"),
				colorDevice(deviceID),
				event.DevicePath,
			)
		}

	case events.EventDeviceDisconnected:
		var event events.DeviceDisconnectedEvent
		if err := envelope.GetData(&event); err == nil {
			fmt.Printf("[%s] %s %s disconnected: %s\n",
				timestamp,
				colorError("✗"),
				colorDevice(deviceID),
				event.Reason,
			)
		}

	case events.EventMeshPacketRx:
		var event events.MeshPacketRxEvent
		if err := envelope.GetData(&event); err == nil {
			packet := event.Packet
			if packet.GetDecoded() != nil {
				fmt.Printf("[%s] %s Message from %08x to %08x: %s\n",
					timestamp,
					colorInfo("📬"),
					packet.From,
					packet.To,
					string(packet.GetDecoded().Payload),
				)
			} else {
				fmt.Printf("[%s] %s Packet from %08x to %08x (encrypted)\n",
					timestamp,
					colorInfo("📦"),
					packet.From,
					packet.To,
				)
			}
		}

	case events.EventNodeInfoUpdated:
		var event events.NodeInfoUpdatedEvent
		if err := envelope.GetData(&event); err == nil {
			status := colorInfo("updated")
			if event.IsNew {
				status = colorSuccess("new")
			}
			fmt.Printf("[%s] %s Node %08x %s: %s\n",
				timestamp,
				colorInfo("🔄"),
				event.NodeInfo.Num,
				status,
				event.NodeInfo.User.ShortName,
			)
		}

	case events.EventTelemetryReceived:
		var event events.TelemetryReceivedEvent
		if err := envelope.GetData(&event); err == nil {
			fmt.Printf("[%s] %s Telemetry from %08x\n",
				timestamp,
				colorInfo("📊"),
				event.NodeID,
			)
		}

	case events.EventPositionUpdated:
		var event events.PositionUpdatedEvent
		if err := envelope.GetData(&event); err == nil {
			if event.Position != nil {
				lat := float64(event.Position.GetLatitudeI()) / 1e7
				lon := float64(event.Position.GetLongitudeI()) / 1e7
				fmt.Printf("[%s] %s Position from %08x: %.6f,%.6f\n",
					timestamp,
					colorInfo("📍"),
					event.NodeID,
					lat,
					lon,
				)
			}
		}

	case events.EventResponseSuccess:
		var event events.ResponseSuccessEvent
		if err := envelope.GetData(&event); err == nil {
			fmt.Printf("[%s] %s Command succeeded\n",
				timestamp,
				colorSuccess("✓"),
			)
		}

	case events.EventResponseError:
		var event events.ResponseErrorEvent
		if err := envelope.GetData(&event); err == nil {
			fmt.Printf("[%s] %s Command failed: %s\n",
				timestamp,
				colorError("✗"),
				event.Error,
			)
		}

	default:
		fmt.Printf("[%s] %s Event: %s\n",
			timestamp,
			colorInfo("•"),
			envelope.Type,
		)
	}
}

func main() {
	kong.Parse(&args)

	// Setup logging
	setupLogging()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info().Str("signal", sig.String()).Msg("Shutdown signal received")
		fmt.Printf("\n%s Received signal %s, shutting down gracefully...\n", colorWarning("⚠"), sig.String())
		cancel()
	}()

	// Initialize and start the REPL
	if err := initializeREPL(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize REPL")
	}

	// Start REPL
	if err := startREPL(ctx); err != nil {
		log.Fatal().Err(err).Msg("REPL failed")
	}
}

func setupLogging() {
	// Parse log level
	level, err := zerolog.ParseLevel(args.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Configure zerolog
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	log.Info().Str("log_level", level.String()).Msg("Logging configured")
}

func initializeREPL(ctx context.Context) error {
	fmt.Println(colorInfo("🚀 Meshtastic Event-Driven REPL"))
	fmt.Println(colorInfo("Type 'help' for available commands"))
	fmt.Println()

	// Create event bus
	busConfig := meshbus.DefaultConfig()
	var err error
	bus, err = meshbus.NewBus(busConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create event bus")
	}

	// Start event bus
	if err := bus.Start(); err != nil {
		return errors.Wrap(err, "failed to start event bus")
	}

	// Create event listener
	eventListener = NewEventListener(bus)
	if err := eventListener.Start(); err != nil {
		return errors.Wrap(err, "failed to start event listener")
	}

	log.Info().Msg("Event bus and listener initialized")
	return nil
}

func startREPL(ctx context.Context) error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			return cleanup()
		default:
			// Show prompt
			fmt.Printf("%s ", colorCommand("meshtastic>"))

			// Read input
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					return errors.Wrap(err, "failed to read input")
				}
				// EOF
				return cleanup()
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			// Process command
			if err := processCommand(ctx, input); err != nil {
				fmt.Printf("%s: %v\n", colorError("Error"), err)
			}
		}
	}
}

func processCommand(ctx context.Context, input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]
	cmdArgs := parts[1:]

	switch command {
	case "help", "h":
		return cmdHelp(cmdArgs)
	case "connect":
		return cmdConnect(ctx, cmdArgs)
	case "disconnect":
		return cmdDisconnect(cmdArgs)
	case "send":
		return cmdSend(cmdArgs)
	case "listen":
		return cmdListen(cmdArgs)
	case "nodes":
		return cmdNodes(cmdArgs)
	case "status":
		return cmdStatus(cmdArgs)
	case "quit", "exit", "q":
		return cmdQuit(cmdArgs)
	default:
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", command)
	}
}

func cmdHelp(cmdArgs []string) error {
	fmt.Println(colorInfo("Available commands:"))
	fmt.Println()
	fmt.Printf("  %s                 Connect to device (auto-discover if no path specified)\n", colorCommand("connect [device_path]"))
	fmt.Printf("  %s              Disconnect from device\n", colorCommand("disconnect"))
	fmt.Printf("  %s         Send text message to node\n", colorCommand("send <node_id> <message>"))
	fmt.Printf("  %s              Toggle event listening\n", colorCommand("listen"))
	fmt.Printf("  %s                   Show node information\n", colorCommand("nodes"))
	fmt.Printf("  %s                  Show device and adapter status\n", colorCommand("status"))
	fmt.Printf("  %s                    Show this help message\n", colorCommand("help"))
	fmt.Printf("  %s                    Exit the REPL\n", colorCommand("quit"))
	fmt.Println()
	fmt.Println(colorInfo("Examples:"))
	fmt.Printf("  %s\n", colorCommand("connect /dev/ttyACM0"))
	fmt.Printf("  %s\n", colorCommand("send 0x12345678 Hello world"))
	fmt.Printf("  %s\n", colorCommand("send 0xFFFFFFFF Broadcast message"))
	fmt.Println()
	return nil
}

func cmdConnect(ctx context.Context, cmdArgs []string) error {
	if adapter != nil && adapter.IsRunning() {
		return errors.New("already connected to a device")
	}

	var devicePath string
	if len(cmdArgs) > 0 {
		devicePath = cmdArgs[0]
	} else {
		// Auto-discover device
		fmt.Printf("%s Auto-discovering device...\n", colorInfo("🔍"))
		discovered, err := discovery.FindBestMeshtasticPort()
		if err != nil {
			return errors.Wrap(err, "failed to auto-discover device")
		}
		devicePath = discovered
	}

	// Check if already connecting to this device
	connectionsMu.Lock()
	if activeConnections[devicePath] {
		connectionsMu.Unlock()
		return errors.New("already connecting to this device")
	}
	activeConnections[devicePath] = true
	connectionsMu.Unlock()

	fmt.Printf("Connecting to device: %s...\n", colorDevice(devicePath))

	// Start connection process in background
	go func() {
		defer func() {
			connectionsMu.Lock()
			delete(activeConnections, devicePath)
			connectionsMu.Unlock()
		}()

		// For now, we'll create a simple client that delays the connection
		// This is a workaround since the current client architecture opens
		// the serial port immediately during construction

		// Sleep briefly to ensure we return to the prompt first
		time.Sleep(100 * time.Millisecond)

		// Create context with timeout
		connectCtx, cancel := context.WithTimeout(ctx, time.Duration(args.Timeout)*time.Second)
		defer cancel()

		// Create client config
		config := &client.Config{
			DevicePath:  devicePath,
			Timeout:     time.Duration(args.Timeout) * time.Second,
			DebugSerial: args.LogLevel == "debug",
			HexDump:     args.LogLevel == "debug",
		}

		// Create robust client - this will fail immediately if device doesn't exist
		robustClient, err := client.NewRobustMeshtasticClient(config)
		if err != nil {
			fmt.Printf("%s Failed to create client: %v\n", colorError("✗"), err)
			return
		}

		// Create device adapter
		deviceID := fmt.Sprintf("dev_%s", strings.ReplaceAll(strings.TrimPrefix(devicePath, "/dev/"), "/", "_"))
		newAdapter := deviceadapter.NewDeviceAdapter(deviceID, robustClient, bus)

		// Start adapter with timeout context
		if err := newAdapter.Start(connectCtx); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("%s Connection timed out after %ds\n", colorError("✗"), args.Timeout)
			} else {
				fmt.Printf("%s Failed to connect: %v\n", colorError("✗"), err)
			}
			return
		}

		// Set global adapter reference
		adapter = newAdapter
		fmt.Printf("%s Connected to %s\n", colorSuccess("✓"), colorDevice(devicePath))
	}()

	return nil
}

func cmdDisconnect(cmdArgs []string) error {
	if adapter == nil || !adapter.IsRunning() {
		return errors.New("not connected to a device")
	}

	fmt.Printf("Disconnecting from device...\n")

	if err := adapter.Stop(); err != nil {
		return errors.Wrap(err, "failed to stop adapter")
	}

	adapter = nil
	fmt.Printf("%s Disconnected\n", colorSuccess("✓"))
	return nil
}

func cmdSend(cmdArgs []string) error {
	if adapter == nil || !adapter.IsRunning() {
		return errors.New("not connected to a device")
	}

	if len(cmdArgs) < 2 {
		return errors.New("usage: send <node_id> <message>")
	}

	// Parse node ID
	nodeIDStr := cmdArgs[0]
	var nodeID uint64
	var err error

	if strings.HasPrefix(nodeIDStr, "0x") {
		nodeID, err = strconv.ParseUint(nodeIDStr[2:], 16, 32)
	} else {
		nodeID, err = strconv.ParseUint(nodeIDStr, 10, 32)
	}

	if err != nil {
		return errors.Wrap(err, "invalid node ID")
	}

	// Join message parts
	messageText := strings.Join(cmdArgs[1:], " ")

	// Create command event
	commandEvent := &events.SendTextCommandEvent{
		Text:        messageText,
		Destination: uint32(nodeID),
		Channel:     0,
	}

	// Create envelope
	envelope, err := events.NewEnvelope(
		events.EventCommandSendText,
		adapter.GetDeviceID(),
		events.SourceREPL,
		commandEvent,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create command envelope")
	}

	// Set correlation ID
	correlationID := uuid.New().String()
	envelope.WithCorrelationID(correlationID)

	// Marshal envelope
	payload, err := envelope.ToJSON()
	if err != nil {
		return errors.Wrap(err, "failed to marshal envelope")
	}

	// Create watermill message
	msg := message.NewMessage(uuid.New().String(), payload)
	msg.Metadata.Set("event_type", events.EventCommandSendText)
	msg.Metadata.Set("device_id", adapter.GetDeviceID())
	msg.Metadata.Set("correlation_id", correlationID)

	// Publish command
	topic := meshbus.BuildTopicName(meshbus.TopicCommandSendText, adapter.GetDeviceID())
	if err := bus.Publisher().Publish(topic, msg); err != nil {
		return errors.Wrap(err, "failed to publish command")
	}

	fmt.Printf("%s Sent message to %s: %s\n", colorSuccess("✓"), colorDevice(nodeIDStr), messageText)
	return nil
}

func cmdListen(cmdArgs []string) error {
	if eventListener == nil {
		return errors.New("event listener not initialized")
	}

	if eventListener.listening {
		eventListener.Stop()
		fmt.Printf("%s Event listening stopped\n", colorWarning("⏸"))
	} else {
		if err := eventListener.Start(); err != nil {
			return errors.Wrap(err, "failed to start event listener")
		}
		fmt.Printf("%s Event listening started\n", colorSuccess("▶"))
	}

	return nil
}

func cmdNodes(cmdArgs []string) error {
	if adapter == nil || !adapter.IsRunning() {
		return errors.New("not connected to a device")
	}

	// Request node info
	commandEvent := &events.RequestInfoCommandEvent{
		NodeID:   0, // All nodes
		InfoType: "nodes",
	}

	// Create envelope
	envelope, err := events.NewEnvelope(
		events.EventCommandRequestInfo,
		adapter.GetDeviceID(),
		events.SourceREPL,
		commandEvent,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create command envelope")
	}

	// Set correlation ID
	correlationID := uuid.New().String()
	envelope.WithCorrelationID(correlationID)

	// Marshal envelope
	payload, err := envelope.ToJSON()
	if err != nil {
		return errors.Wrap(err, "failed to marshal envelope")
	}

	// Create watermill message
	msg := message.NewMessage(uuid.New().String(), payload)
	msg.Metadata.Set("event_type", events.EventCommandRequestInfo)
	msg.Metadata.Set("device_id", adapter.GetDeviceID())
	msg.Metadata.Set("correlation_id", correlationID)

	// Publish command
	topic := meshbus.BuildTopicName(meshbus.TopicCommandRequestInfo, adapter.GetDeviceID())
	if err := bus.Publisher().Publish(topic, msg); err != nil {
		return errors.Wrap(err, "failed to publish command")
	}

	fmt.Printf("%s Requested node information\n", colorSuccess("✓"))
	return nil
}

func cmdStatus(cmdArgs []string) error {
	fmt.Printf("%s System Status:\n", colorInfo("📊"))
	fmt.Printf("  Event Bus: %s\n", colorSuccess("running"))
	fmt.Printf("  Event Listener: %s\n", getStatusColor(eventListener.listening))

	if adapter != nil && adapter.IsRunning() {
		fmt.Printf("  Device: %s (%s)\n", colorSuccess("connected"), colorDevice(adapter.GetDevicePath()))
		fmt.Printf("  Device ID: %s\n", adapter.GetDeviceID())

		// Show adapter statistics
		stats := adapter.GetStatistics()
		fmt.Printf("  Messages Received: %d\n", stats.MessagesReceived)
		fmt.Printf("  Messages Sent: %d\n", stats.MessagesSent)
		fmt.Printf("  Events Published: %d\n", stats.EventsPublished)
		fmt.Printf("  Commands Processed: %d\n", stats.CommandsProcessed)
		fmt.Printf("  Errors: %d\n", stats.Errors)
		fmt.Printf("  Uptime: %s\n", time.Since(stats.UptimeStart).Round(time.Second))
		fmt.Printf("  Last Activity: %s\n", stats.LastActivity.Format("15:04:05"))
	} else {
		fmt.Printf("  Device: %s\n", colorError("disconnected"))
	}

	return nil
}

func cmdQuit(cmdArgs []string) error {
	fmt.Printf("%s Goodbye!\n", colorInfo("👋"))
	os.Exit(0)
	return nil
}

func cleanup() error {
	fmt.Println("\nShutting down...")

	// Stop event listener first
	if eventListener != nil {
		log.Info().Msg("Stopping event listener...")
		eventListener.Stop()
	}

	// Stop adapter
	if adapter != nil {
		log.Info().Msg("Stopping device adapter...")
		if err := adapter.Stop(); err != nil {
			log.Error().Err(err).Msg("Error stopping adapter")
		}
		adapter = nil
	}

	// Stop bus last
	if bus != nil {
		log.Info().Msg("Stopping event bus...")
		if err := bus.Stop(); err != nil {
			log.Error().Err(err).Msg("Error stopping bus")
		}
		bus = nil
	}

	fmt.Println("Shutdown complete")
	return nil
}

func getStatusColor(running bool) string {
	if running {
		return colorSuccess("running")
	}
	return colorError("stopped")
}
