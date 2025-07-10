package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/client"
	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/serial/discovery"
)

var (
	port           string
	host           string
	bleAddress     string
	bleScan        bool
	timeout        time.Duration
	connectTimeout time.Duration
)

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a Meshtastic device",
	Long: `Connect to a Meshtastic device using various connection methods.

Examples:
  meshtastic connect --port /dev/ttyUSB0
  meshtastic connect --host 192.168.1.100
  meshtastic connect --ble-scan`,
	RunE: runConnect,
}

func runConnect(cmd *cobra.Command, args []string) error {
	if bleScan {
		return runBLEScan(cmd, args)
	}

	// Determine connection method
	var config *client.Config
	if host != "" {
		// TODO: Implement TCP connection
		return errors.New("TCP connection not yet implemented")
	} else if bleAddress != "" {
		// TODO: Implement BLE connection
		return errors.New("BLE connection not yet implemented")
	} else {
		// Serial connection
		if port == "" {
			// Auto-discover serial port
			log.Info().Msg("Auto-discovering Meshtastic devices...")
			devices, err := discovery.DiscoverMeshtasticDevices(discovery.DiscoveryConfig{
				Timeout:    timeout,
				SerialOnly: true,
			})
			if err != nil {
				return errors.Wrap(err, "failed to discover devices")
			}

			if len(devices) == 0 {
				return errors.New("no Meshtastic devices found")
			}

			// Use the first discovered device
			port = devices[0].Port
			log.Info().Str("port", port).Msg("Using auto-discovered device")
		}

		config = &client.Config{
			DevicePath:  port,
			Timeout:     connectTimeout,
			DebugSerial: globalConfig.DebugSerial,
			HexDump:     globalConfig.HexDump,
		}
	}

	// Create robust client
	meshtasticClient, err := client.NewRobustMeshtasticClient(config)
	if err != nil {
		return errors.Wrap(err, "failed to create robust client")
	}
	defer meshtasticClient.Close()

	// Connect to the device
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	log.Info().Str("device", config.DevicePath).Msg("Connecting to Meshtastic device")
	if err := meshtasticClient.Connect(ctx); err != nil {
		return errors.Wrap(err, "failed to connect to device")
	}
	defer meshtasticClient.Disconnect()

	// Check connection status
	if !meshtasticClient.IsConnected() {
		return errors.New("device connection failed")
	}

	// Get device info
	myInfo := meshtasticClient.GetMyInfo()
	nodes := meshtasticClient.GetNodes()

	// Display connection info
	fmt.Printf("Connected to: %s\n", config.DevicePath)
	if myInfo != nil {
		fmt.Printf("Node ID: !%08x\n", myInfo.MyNodeNum)

		// Find our node info for more details
		if node, exists := nodes[myInfo.MyNodeNum]; exists {
			user := node.GetUser()
			if user != nil {
				if user.LongName != "" {
					fmt.Printf("User: %s (%s)\n", user.LongName, user.ShortName)
				}
				if user.HwModel != pb.HardwareModel_UNSET {
					fmt.Printf("Hardware: %s\n", user.HwModel.String())
				}
			}
		}
	}

	fmt.Printf("Status: Connected\n")
	return nil
}

func runBLEScan(cmd *cobra.Command, args []string) error {
	// TODO: Implement BLE scanning
	return errors.New("BLE scanning not yet implemented")
}

func init() {
	connectCmd.Flags().StringVarP(&port, "port", "p", "", "Serial port (e.g., /dev/ttyUSB0, COM3)")
	connectCmd.Flags().StringVar(&host, "host", "", "TCP/IP host (e.g., 192.168.1.100, meshtastic.local)")
	connectCmd.Flags().StringVar(&bleAddress, "ble-address", "", "BLE device address")
	connectCmd.Flags().BoolVar(&bleScan, "ble-scan", false, "Scan for BLE devices")
	connectCmd.Flags().DurationVar(&connectTimeout, "timeout", 30*time.Second, "Connection timeout")
}
