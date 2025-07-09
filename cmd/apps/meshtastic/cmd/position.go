package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/client"
	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
)

// positionCmd represents the position command
var positionCmd = &cobra.Command{
	Use:   "position",
	Short: "Manage GPS position",
	Long:  `Manage GPS position information for the Meshtastic device.`,
}

// positionGetCmd represents the position get command
var positionGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current position",
	Long:  `Get the current GPS position from the Meshtastic device.`,
	RunE:  runPositionGet,
}

// positionSetCmd represents the position set command
var positionSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set fixed position",
	Long:  `Set a fixed GPS position for the Meshtastic device.`,
	RunE:  runPositionSet,
}

// positionClearCmd represents the position clear command
var positionClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear fixed position",
	Long:  `Clear the fixed GPS position and return to GPS mode.`,
	RunE:  runPositionClear,
}

// positionRequestCmd represents the position request command
var positionRequestCmd = &cobra.Command{
	Use:   "request",
	Short: "Request position from node",
	Long:  `Request position information from other nodes in the mesh.`,
	RunE:  runPositionRequest,
}

// positionBroadcastCmd represents the position broadcast command
var positionBroadcastCmd = &cobra.Command{
	Use:   "broadcast",
	Short: "Broadcast current position",
	Long:  `Broadcast the current position to the mesh network.`,
	RunE:  runPositionBroadcast,
}

var (
	latFlag             float64
	lonFlag             float64
	altFlag             int32
	fixedFlag           bool
	allFlag             bool
	destNodeFlag        string
	positionTimeoutFlag time.Duration
)

func runPositionGet(cmd *cobra.Command, args []string) error {
	log.Info().Str("port", globalConfig.Port).Msg("Getting current position")

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

	// Get my node info
	myInfo := meshtasticClient.GetMyInfo()
	if myInfo == nil {
		return errors.New("failed to get device info")
	}

	// Get nodes to find position
	nodes := meshtasticClient.GetNodes()
	if nodes == nil {
		return errors.New("failed to get nodes")
	}

	myNode, exists := nodes[myInfo.GetMyNodeNum()]
	if !exists {
		return errors.New("own node not found in nodes list")
	}

	position := myNode.GetPosition()
	if position == nil {
		fmt.Println("No position information available")
		return nil
	}

	// Display position information
	fmt.Println("Current Position:")

	if position.LatitudeI != nil {
		lat := float64(*position.LatitudeI) * 1e-7
		latDir := "N"
		if lat < 0 {
			latDir = "S"
			lat = -lat
		}
		fmt.Printf("  Latitude: %.6f°%s\n", lat, latDir)
	}

	if position.LongitudeI != nil {
		lon := float64(*position.LongitudeI) * 1e-7
		lonDir := "E"
		if lon < 0 {
			lonDir = "W"
			lon = -lon
		}
		fmt.Printf("  Longitude: %.6f°%s\n", lon, lonDir)
	}

	if position.Altitude != nil {
		fmt.Printf("  Altitude: %dm\n", *position.Altitude)
	}

	// Source information
	source := "Unknown"
	switch position.GetLocationSource() {
	case pb.Position_LOC_UNSET:
		source = "Unset"
	case pb.Position_LOC_MANUAL:
		source = "Manual"
	case pb.Position_LOC_INTERNAL:
		source = "Internal GPS"
	case pb.Position_LOC_EXTERNAL:
		source = "External GPS"
	}
	fmt.Printf("  Source: %s\n", source)

	// Precision information
	if position.GetPDOP() > 0 {
		fmt.Printf("  Precision: PDOP %.1f\n", float64(position.GetPDOP())/100.0)
	}

	// Timestamp
	if position.GetTimestamp() > 0 {
		timestamp := time.Unix(int64(position.GetTimestamp()), 0)
		fmt.Printf("  Last Updated: %s\n", timestamp.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func runPositionSet(cmd *cobra.Command, args []string) error {
	if latFlag == 0 && lonFlag == 0 {
		return errors.New("latitude and longitude are required (use --lat and --lon flags)")
	}

	// Validate coordinates
	if latFlag < -90 || latFlag > 90 {
		return errors.New("latitude must be between -90 and 90 degrees")
	}
	if lonFlag < -180 || lonFlag > 180 {
		return errors.New("longitude must be between -180 and 180 degrees")
	}

	log.Info().
		Str("port", globalConfig.Port).
		Float64("lat", latFlag).
		Float64("lon", lonFlag).
		Int32("alt", altFlag).
		Bool("fixed", fixedFlag).
		Msg("Setting position")

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

	// Create position message
	latI := int32(latFlag * 1e7)
	lonI := int32(lonFlag * 1e7)

	position := &pb.Position{
		LatitudeI:      &latI,
		LongitudeI:     &lonI,
		Time:           uint32(time.Now().Unix()),
		LocationSource: pb.Position_LOC_MANUAL,
	}

	if altFlag != 0 {
		position.Altitude = &altFlag
	}

	// Create admin message to set fixed position
	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_SetFixedPosition{
			SetFixedPosition: position,
		},
	}

	// Send admin message
	_, err = meshtasticClient.SendAdminMessage(adminMsg)
	if err != nil {
		return errors.Wrap(err, "failed to set position")
	}

	fmt.Printf("Position set successfully:\n")
	fmt.Printf("  Latitude: %.6f°\n", latFlag)
	fmt.Printf("  Longitude: %.6f°\n", lonFlag)
	if altFlag != 0 {
		fmt.Printf("  Altitude: %dm\n", altFlag)
	}
	fmt.Printf("  Source: Manual (Fixed)\n")

	return nil
}

func runPositionClear(cmd *cobra.Command, args []string) error {
	log.Info().Str("port", globalConfig.Port).Msg("Clearing fixed position")

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

	// Create admin message to remove fixed position
	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_RemoveFixedPosition{
			RemoveFixedPosition: true,
		},
	}

	// Send admin message
	_, err = meshtasticClient.SendAdminMessage(adminMsg)
	if err != nil {
		return errors.Wrap(err, "failed to clear position")
	}

	fmt.Println("Fixed position cleared successfully")
	fmt.Println("Device will now use GPS for position information")

	return nil
}

func runPositionRequest(cmd *cobra.Command, args []string) error {
	var dest uint32
	var err error

	if allFlag {
		dest = client.BROADCAST_ADDR
	} else if destNodeFlag != "" {
		dest, err = parseDestination(destNodeFlag)
		if err != nil {
			return errors.Wrap(err, "invalid destination node")
		}
	} else {
		return errors.New("either --dest or --all flag is required")
	}

	log.Info().
		Str("port", globalConfig.Port).
		Uint32("dest", dest).
		Msg("Requesting position")

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

	// Create position request packet
	packet := &pb.MeshPacket{
		To:      dest,
		WantAck: true,
		PayloadVariant: &pb.MeshPacket_Decoded{
			Decoded: &pb.Data{
				Portnum:   pb.PortNum_POSITION_APP,
				Payload:   []byte{}, // Empty payload for position request
				RequestId: uint32(time.Now().Unix()),
			},
		},
	}

	// Send position request
	err = meshtasticClient.SendMessage(packet)
	if err != nil {
		return errors.Wrap(err, "failed to send position request")
	}

	if dest == client.BROADCAST_ADDR {
		fmt.Println("Position request sent to all nodes")
	} else {
		fmt.Printf("Position request sent to node %s\n", destNodeFlag)
	}

	// Set up position handler to receive responses
	if positionTimeoutFlag > 0 {
		fmt.Printf("Waiting for responses (timeout: %v)...\n", positionTimeoutFlag)

		responseCtx, responseCancel := context.WithTimeout(context.Background(), positionTimeoutFlag)
		defer responseCancel()

		positionReceived := make(chan *pb.Position, 10)
		meshtasticClient.SetOnPosition(func(position *pb.Position) {
			select {
			case positionReceived <- position:
			default:
				log.Warn().Msg("Position response buffer full")
			}
		})

		// Wait for responses
		for {
			select {
			case <-responseCtx.Done():
				fmt.Println("Timeout reached")
				return nil
			case position := <-positionReceived:
				fmt.Printf("Position received:\n")
				if position.LatitudeI != nil {
					lat := float64(*position.LatitudeI) * 1e-7
					fmt.Printf("  Latitude: %.6f°\n", lat)
				}
				if position.LongitudeI != nil {
					lon := float64(*position.LongitudeI) * 1e-7
					fmt.Printf("  Longitude: %.6f°\n", lon)
				}
				if position.Altitude != nil {
					fmt.Printf("  Altitude: %dm\n", *position.Altitude)
				}
				if !allFlag {
					// Single node request, return after first response
					return nil
				}
			}
		}
	}

	return nil
}

func runPositionBroadcast(cmd *cobra.Command, args []string) error {
	log.Info().Str("port", globalConfig.Port).Msg("Broadcasting current position")

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

	// Get current position
	myInfo := meshtasticClient.GetMyInfo()
	if myInfo == nil {
		return errors.New("failed to get device info")
	}

	nodes := meshtasticClient.GetNodes()
	if nodes == nil {
		return errors.New("failed to get nodes")
	}

	myNode, exists := nodes[myInfo.GetMyNodeNum()]
	if !exists {
		return errors.New("own node not found in nodes list")
	}

	position := myNode.GetPosition()
	if position == nil {
		return errors.New("no position information available to broadcast")
	}

	// Create position broadcast packet
	positionBytes, err := proto.Marshal(position)
	if err != nil {
		return errors.Wrap(err, "failed to marshal position")
	}

	packet := &pb.MeshPacket{
		To:      client.BROADCAST_ADDR,
		Channel: 0,
		PayloadVariant: &pb.MeshPacket_Decoded{
			Decoded: &pb.Data{
				Portnum: pb.PortNum_POSITION_APP,
				Payload: positionBytes,
			},
		},
	}

	// Send position broadcast
	err = meshtasticClient.SendMessage(packet)
	if err != nil {
		return errors.Wrap(err, "failed to broadcast position")
	}

	fmt.Println("Position broadcast sent successfully")

	// Display broadcasted position
	if position.LatitudeI != nil {
		lat := float64(*position.LatitudeI) * 1e-7
		fmt.Printf("  Latitude: %.6f°\n", lat)
	}
	if position.LongitudeI != nil {
		lon := float64(*position.LongitudeI) * 1e-7
		fmt.Printf("  Longitude: %.6f°\n", lon)
	}
	if position.Altitude != nil {
		fmt.Printf("  Altitude: %dm\n", *position.Altitude)
	}

	return nil
}

// parseDMS parses degrees, minutes, seconds format
func parseDMS(dms string) (float64, error) {
	// Simple implementation - just handle decimal degrees for now
	return strconv.ParseFloat(dms, 64)
}

// validateCoordinates validates GPS coordinates
func validateCoordinates(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return errors.New("latitude must be between -90 and 90 degrees")
	}
	if lon < -180 || lon > 180 {
		return errors.New("longitude must be between -180 and 180 degrees")
	}
	return nil
}

func init() {
	// Add flags for position set
	positionSetCmd.Flags().Float64Var(&latFlag, "lat", 0, "Latitude in decimal degrees")
	positionSetCmd.Flags().Float64Var(&lonFlag, "lon", 0, "Longitude in decimal degrees")
	positionSetCmd.Flags().Int32Var(&altFlag, "alt", 0, "Altitude in meters")
	positionSetCmd.Flags().BoolVar(&fixedFlag, "fixed", true, "Mark position as fixed (not GPS)")
	positionSetCmd.MarkFlagRequired("lat")
	positionSetCmd.MarkFlagRequired("lon")

	// Add flags for position request
	positionRequestCmd.Flags().StringVar(&destNodeFlag, "dest", "", "Destination node ID, name, or number")
	positionRequestCmd.Flags().BoolVar(&allFlag, "all", false, "Request from all nodes")
	positionRequestCmd.Flags().DurationVar(&positionTimeoutFlag, "timeout", 30*time.Second, "Request timeout")
	positionRequestCmd.MarkFlagsMutuallyExclusive("dest", "all")

	// Add subcommands
	positionCmd.AddCommand(positionGetCmd)
	positionCmd.AddCommand(positionSetCmd)
	positionCmd.AddCommand(positionClearCmd)
	positionCmd.AddCommand(positionRequestCmd)
	positionCmd.AddCommand(positionBroadcastCmd)

	// Add to root command
	rootCmd.AddCommand(positionCmd)
}
