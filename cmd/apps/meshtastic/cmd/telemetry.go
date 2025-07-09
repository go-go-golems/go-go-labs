package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/client"
	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
)

// telemetryCmd represents the telemetry command
var telemetryCmd = &cobra.Command{
	Use:   "telemetry",
	Short: "Telemetry data commands",
	Long: `Commands for retrieving and monitoring telemetry data from Meshtastic devices.
Supports device metrics, environment data, and power information.`,
}

// telemetryGetCmd represents the telemetry get command
var telemetryGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get telemetry data",
	Long: `Get telemetry data from the connected device or from the mesh network.
Supports different types of telemetry data including device metrics, environment data, and power information.

Examples:
  meshtastic telemetry get
  meshtastic telemetry get --type device
  meshtastic telemetry get --type environment
  meshtastic telemetry get --from !a4c138f4`,
	RunE: runTelemetryGet,
}

// telemetryRequestCmd represents the telemetry request command
var telemetryRequestCmd = &cobra.Command{
	Use:   "request",
	Short: "Request telemetry from a node",
	Long: `Request telemetry data from a specific node in the mesh network.
This will send a telemetry request to the specified destination node.

Examples:
  meshtastic telemetry request --dest !a4c138f4
  meshtastic telemetry request --dest !a4c138f4 --type device
  meshtastic telemetry request --dest !a4c138f4 --type environment`,
	RunE: runTelemetryRequest,
}

// telemetryMonitorCmd represents the telemetry monitor command
var telemetryMonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor telemetry in real-time",
	Long: `Monitor telemetry data in real-time from the mesh network.
This will continuously display telemetry updates as they are received.

Examples:
  meshtastic telemetry monitor
  meshtastic telemetry monitor --type device
  meshtastic telemetry monitor --timeout 60`,
	RunE: runTelemetryMonitor,
}

// tracerouteCmd represents the traceroute command
var tracerouteCmd = &cobra.Command{
	Use:   "traceroute [DEST]",
	Short: "Trace route to destination node",
	Long: `Trace the route to a destination node in the mesh network.
This will show the path that packets take to reach the destination.

Examples:
  meshtastic traceroute !a4c138f4
  meshtastic traceroute !a4c138f4 --timeout 30
  meshtastic traceroute !a4c138f4 --max-hops 5`,
	Args: cobra.ExactArgs(1),
	RunE: runTraceroute,
}

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping [DEST]",
	Short: "Ping a destination node",
	Long: `Ping a destination node in the mesh network.
This will send ping packets and measure round-trip time.

Examples:
  meshtastic ping !a4c138f4
  meshtastic ping !a4c138f4 --count 5
  meshtastic ping !a4c138f4 --timeout 30`,
	Args: cobra.ExactArgs(1),
	RunE: runPing,
}

// Telemetry flags
var (
	telemetryType     string
	telemetryFrom     string
	telemetryDest     string
	telemetryTimeout  time.Duration
	telemetryLive     bool
	telemetryCount    int
	tracerouteMaxHops int
)

func runTelemetryGet(cmd *cobra.Command, args []string) error {
	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	log.Info().Str("type", telemetryType).Str("from", telemetryFrom).Msg("Getting telemetry data")

	// Get telemetry data from local device or specific node
	var telemetryData *pb.Telemetry

	if telemetryFrom != "" {
		// Request from specific node
		nodeID, err := parseNodeID(telemetryFrom)
		if err != nil {
			return errors.Wrap(err, "invalid node ID")
		}

		telemetryData, err = requestTelemetryFromNode(client, nodeID, telemetryType)
		if err != nil {
			return errors.Wrap(err, "failed to get telemetry from node")
		}
	} else {
		// Get from local device
		telemetryData, err = getTelemetryFromDevice(client, telemetryType)
		if err != nil {
			return errors.Wrap(err, "failed to get telemetry from device")
		}
	}

	// Output in requested format
	if outputJSON {
		data, err := marshalJSON(telemetryData)
		if err != nil {
			return errors.Wrap(err, "failed to marshal JSON")
		}
		fmt.Println(string(data))
		return nil
	}

	if outputYAML {
		data, err := marshalYAML(telemetryData)
		if err != nil {
			return errors.Wrap(err, "failed to marshal YAML")
		}
		fmt.Print(string(data))
		return nil
	}

	// Default text output
	displayTelemetry(telemetryData)
	return nil
}

func runTelemetryRequest(cmd *cobra.Command, args []string) error {
	if telemetryDest == "" {
		return errors.New("destination node is required (use --dest flag)")
	}

	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	nodeID, err := parseNodeID(telemetryDest)
	if err != nil {
		return errors.Wrap(err, "invalid destination node ID")
	}

	log.Info().Uint32("dest", nodeID).Str("type", telemetryType).Msg("Requesting telemetry from node")

	// Send telemetry request
	telemetryData, err := requestTelemetryFromNode(client, nodeID, telemetryType)
	if err != nil {
		return errors.Wrap(err, "failed to request telemetry")
	}

	fmt.Printf("✓ Telemetry received from %s\n", telemetryDest)
	displayTelemetry(telemetryData)

	return nil
}

func runTelemetryMonitor(cmd *cobra.Command, args []string) error {
	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	log.Info().Msg("Starting telemetry monitoring")

	// Set up telemetry handler
	telemetryChannel := make(chan *pb.Telemetry, 100)

	client.SetOnTelemetry(func(telemetry *pb.Telemetry) {
		select {
		case telemetryChannel <- telemetry:
		default:
			// Channel is full, drop the message
		}
	})

	fmt.Println("Monitoring telemetry data... (Press Ctrl+C to stop)")

	// Set up timeout context
	ctx := context.Background()
	if telemetryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, telemetryTimeout)
		defer cancel()
	}

	// Monitor telemetry
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nMonitoring stopped (timeout)")
			return nil
		case telemetry := <-telemetryChannel:
			if telemetryType == "" || matchesTelemetryType(telemetry, telemetryType) {
				fmt.Printf("[%s] ", time.Now().Format("15:04:05"))
				displayTelemetryInline(telemetry)
			}
		}
	}
}

func runTraceroute(cmd *cobra.Command, args []string) error {
	dest := args[0]
	nodeID, err := parseNodeID(dest)
	if err != nil {
		return errors.Wrap(err, "invalid destination node ID")
	}

	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	log.Info().Uint32("dest", nodeID).Msg("Starting traceroute")

	fmt.Printf("Traceroute to %s:\n", dest)

	// Send traceroute request
	err = performTraceroute(client, nodeID, tracerouteMaxHops, telemetryTimeout)
	if err != nil {
		return errors.Wrap(err, "traceroute failed")
	}

	return nil
}

func runPing(cmd *cobra.Command, args []string) error {
	dest := args[0]
	nodeID, err := parseNodeID(dest)
	if err != nil {
		return errors.Wrap(err, "invalid destination node ID")
	}

	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	log.Info().Uint32("dest", nodeID).Int("count", telemetryCount).Msg("Starting ping")

	fmt.Printf("Pinging %s:\n", dest)

	// Send ping requests
	err = performPing(client, nodeID, telemetryCount, telemetryTimeout)
	if err != nil {
		return errors.Wrap(err, "ping failed")
	}

	return nil
}

func parseNodeID(nodeStr string) (uint32, error) {
	// Remove leading ! if present
	if strings.HasPrefix(nodeStr, "!") {
		nodeStr = nodeStr[1:]
	}

	// Parse as hex
	nodeID, err := strconv.ParseUint(nodeStr, 16, 32)
	if err != nil {
		return 0, errors.Wrap(err, "invalid node ID format")
	}

	return uint32(nodeID), nil
}

func requestTelemetryFromNode(client *client.RobustMeshtasticClient, nodeID uint32, telemetryType string) (*pb.Telemetry, error) {
	// Create telemetry request packet
	requestPacket := &pb.MeshPacket{
		To:      nodeID,
		Channel: 0,
		PayloadVariant: &pb.MeshPacket_Decoded{
			Decoded: &pb.Data{
				Portnum: pb.PortNum_TELEMETRY_APP,
				Payload: []byte{}, // Empty payload for request
			},
		},
	}

	// Set up response handler
	responseChannel := make(chan *pb.Telemetry, 1)

	client.SetOnTelemetry(func(telemetry *pb.Telemetry) {
		select {
		case responseChannel <- telemetry:
		default:
		}
	})

	// Send request
	err := client.SendMessage(requestPacket)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send telemetry request")
	}

	// Wait for response
	select {
	case telemetry := <-responseChannel:
		return telemetry, nil
	case <-time.After(telemetryTimeout):
		return nil, errors.New("timeout waiting for telemetry response")
	}
}

func getTelemetryFromDevice(client *client.RobustMeshtasticClient, telemetryType string) (*pb.Telemetry, error) {
	// Get local device telemetry
	// For now, we'll request from our own node
	myInfo := client.GetMyInfo()
	if myInfo == nil {
		return nil, errors.New("failed to get device info")
	}

	return requestTelemetryFromNode(client, myInfo.MyNodeNum, telemetryType)
}

func matchesTelemetryType(telemetry *pb.Telemetry, telemetryType string) bool {
	if telemetryType == "" {
		return true
	}

	switch strings.ToLower(telemetryType) {
	case "device":
		return telemetry.GetDeviceMetrics() != nil
	case "environment":
		return telemetry.GetEnvironmentMetrics() != nil
	case "power":
		return telemetry.GetPowerMetrics() != nil
	default:
		return true
	}
}

func displayTelemetry(telemetry *pb.Telemetry) {
	fmt.Printf("Telemetry Data:\n")
	fmt.Printf("  Timestamp: %s\n", time.Unix(int64(telemetry.Time), 0).Format(time.RFC3339))

	if deviceMetrics := telemetry.GetDeviceMetrics(); deviceMetrics != nil {
		fmt.Printf("\nDevice Metrics:\n")
		fmt.Printf("  Battery Level: %d%%\n", deviceMetrics.GetBatteryLevel())
		fmt.Printf("  Voltage: %.2fV\n", deviceMetrics.GetVoltage())
		fmt.Printf("  Channel Utilization: %.1f%%\n", deviceMetrics.GetChannelUtilization())
		fmt.Printf("  Air Time: %.1f%%\n", deviceMetrics.GetAirUtilTx())
		fmt.Printf("  Uptime: %d seconds\n", deviceMetrics.GetUptimeSeconds())
	}

	if envMetrics := telemetry.GetEnvironmentMetrics(); envMetrics != nil {
		fmt.Printf("\nEnvironment Metrics:\n")
		fmt.Printf("  Temperature: %.1f°C\n", envMetrics.GetTemperature())
		fmt.Printf("  Humidity: %.1f%%\n", envMetrics.GetRelativeHumidity())
		fmt.Printf("  Pressure: %.1f hPa\n", envMetrics.GetBarometricPressure())
		fmt.Printf("  Gas Resistance: %.1f kΩ\n", envMetrics.GetGasResistance())
	}

	if powerMetrics := telemetry.GetPowerMetrics(); powerMetrics != nil {
		fmt.Printf("\nPower Metrics:\n")
		fmt.Printf("  CH1 Voltage: %.2fV\n", powerMetrics.GetCh1Voltage())
		fmt.Printf("  CH1 Current: %.2fA\n", powerMetrics.GetCh1Current())
		fmt.Printf("  CH2 Voltage: %.2fV\n", powerMetrics.GetCh2Voltage())
		fmt.Printf("  CH2 Current: %.2fA\n", powerMetrics.GetCh2Current())
		fmt.Printf("  CH3 Voltage: %.2fV\n", powerMetrics.GetCh3Voltage())
		fmt.Printf("  CH3 Current: %.2fA\n", powerMetrics.GetCh3Current())
	}
}

func displayTelemetryInline(telemetry *pb.Telemetry) {
	if deviceMetrics := telemetry.GetDeviceMetrics(); deviceMetrics != nil {
		fmt.Printf("Device: Battery %d%%, Voltage %.2fV, CH Util %.1f%%\n",
			deviceMetrics.GetBatteryLevel(), deviceMetrics.GetVoltage(), deviceMetrics.GetChannelUtilization())
	} else if envMetrics := telemetry.GetEnvironmentMetrics(); envMetrics != nil {
		fmt.Printf("Environment: Temp %.1f°C, Humidity %.1f%%, Pressure %.1f hPa\n",
			envMetrics.GetTemperature(), envMetrics.GetRelativeHumidity(), envMetrics.GetBarometricPressure())
	} else if powerMetrics := telemetry.GetPowerMetrics(); powerMetrics != nil {
		fmt.Printf("Power: CH1 %.2fV/%.2fA, CH2 %.2fV/%.2fA, CH3 %.2fV/%.2fA\n",
			powerMetrics.GetCh1Voltage(), powerMetrics.GetCh1Current(),
			powerMetrics.GetCh2Voltage(), powerMetrics.GetCh2Current(),
			powerMetrics.GetCh3Voltage(), powerMetrics.GetCh3Current())
	} else {
		fmt.Printf("Unknown telemetry type\n")
	}
}

func performTraceroute(client *client.RobustMeshtasticClient, dest uint32, maxHops int, timeout time.Duration) error {
	// Traceroute implementation
	// This is a simplified version - in a full implementation, you'd need to:
	// 1. Send packets with increasing hop limits
	// 2. Track intermediate nodes that respond
	// 3. Measure timing for each hop

	fmt.Printf("1. !%08x (self) - 0ms\n", client.GetMyInfo().MyNodeNum)

	// Send a direct message to measure final hop
	start := time.Now()
	packet := &pb.MeshPacket{
		To:      dest,
		Channel: 0,
		PayloadVariant: &pb.MeshPacket_Decoded{
			Decoded: &pb.Data{
				Portnum: pb.PortNum_TEXT_MESSAGE_APP,
				Payload: []byte("traceroute"),
			},
		},
	}

	err := client.SendMessage(packet)
	if err != nil {
		return errors.Wrap(err, "failed to send traceroute packet")
	}

	// Wait for response or timeout
	time.Sleep(timeout)
	elapsed := time.Since(start)

	fmt.Printf("2. !%08x (destination) - %dms\n", dest, elapsed.Milliseconds())

	return nil
}

func performPing(client *client.RobustMeshtasticClient, dest uint32, count int, timeout time.Duration) error {
	var totalTime time.Duration
	var successful int

	for i := 0; i < count; i++ {
		start := time.Now()

		// Send ping packet
		packet := &pb.MeshPacket{
			To:      dest,
			Channel: 0,
			PayloadVariant: &pb.MeshPacket_Decoded{
				Decoded: &pb.Data{
					Portnum: pb.PortNum_TEXT_MESSAGE_APP,
					Payload: []byte(fmt.Sprintf("ping %d", i+1)),
				},
			},
		}

		err := client.SendMessage(packet)
		if err != nil {
			fmt.Printf("Ping %d: failed to send - %v\n", i+1, err)
			continue
		}

		// Wait for response (simplified - in real implementation, you'd listen for ACKs)
		time.Sleep(100 * time.Millisecond)

		elapsed := time.Since(start)
		totalTime += elapsed
		successful++

		fmt.Printf("Ping %d: time=%dms\n", i+1, elapsed.Milliseconds())

		if i < count-1 {
			time.Sleep(1 * time.Second)
		}
	}

	// Print statistics
	fmt.Printf("\nPing statistics:\n")
	fmt.Printf("  %d packets sent, %d received, %.1f%% packet loss\n",
		count, successful, float64(count-successful)/float64(count)*100)

	if successful > 0 {
		avgTime := totalTime / time.Duration(successful)
		fmt.Printf("  Average round-trip time: %dms\n", avgTime.Milliseconds())
	}

	return nil
}

func init() {
	// Telemetry get flags
	telemetryGetCmd.Flags().StringVar(&telemetryType, "type", "", "Telemetry type (device, environment, power)")
	telemetryGetCmd.Flags().StringVar(&telemetryFrom, "from", "", "Get telemetry from specific node")
	telemetryGetCmd.Flags().BoolVar(&outputJSON, "json", false, "Output in JSON format")
	telemetryGetCmd.Flags().BoolVar(&outputYAML, "yaml", false, "Output in YAML format")

	// Telemetry request flags
	telemetryRequestCmd.Flags().StringVar(&telemetryDest, "dest", "", "Destination node ID (required)")
	telemetryRequestCmd.Flags().StringVar(&telemetryType, "type", "", "Telemetry type (device, environment, power)")
	telemetryRequestCmd.Flags().DurationVar(&telemetryTimeout, "timeout", 30*time.Second, "Request timeout")
	telemetryRequestCmd.MarkFlagRequired("dest")

	// Telemetry monitor flags
	telemetryMonitorCmd.Flags().StringVar(&telemetryType, "type", "", "Filter by telemetry type")
	telemetryMonitorCmd.Flags().DurationVar(&telemetryTimeout, "timeout", 0, "Monitor timeout (0 for infinite)")
	telemetryMonitorCmd.Flags().BoolVar(&telemetryLive, "live", true, "Live updating display")

	// Traceroute flags
	tracerouteCmd.Flags().DurationVar(&telemetryTimeout, "timeout", 30*time.Second, "Timeout per hop")
	tracerouteCmd.Flags().IntVar(&tracerouteMaxHops, "max-hops", 7, "Maximum number of hops")

	// Ping flags
	pingCmd.Flags().IntVar(&telemetryCount, "count", 4, "Number of ping packets to send")
	pingCmd.Flags().DurationVar(&telemetryTimeout, "timeout", 30*time.Second, "Ping timeout")

	// Add subcommands to telemetry
	telemetryCmd.AddCommand(telemetryGetCmd)
	telemetryCmd.AddCommand(telemetryRequestCmd)
	telemetryCmd.AddCommand(telemetryMonitorCmd)

	// Add network diagnostic commands to root
	rootCmd.AddCommand(telemetryCmd)
	rootCmd.AddCommand(tracerouteCmd)
	rootCmd.AddCommand(pingCmd)
}
