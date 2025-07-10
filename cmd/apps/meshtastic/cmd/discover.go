package cmd

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/serial/discovery"
)

var (
	serialOnly bool
	tcpOnly    bool
	bleOnly    bool
)

// discoverCmd represents the discover command
var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover available Meshtastic devices",
	Long: `Discover available Meshtastic devices on all interfaces.

Examples:
  meshtastic discover
  meshtastic discover --serial-only
  meshtastic discover --timeout 5s`,
	RunE: runDiscover,
}

func runDiscover(cmd *cobra.Command, args []string) error {
	log.Info().Msg("Discovering Meshtastic devices...")

	// Configure discovery options
	config := discovery.DiscoveryConfig{
		Timeout:    timeout,
		SerialOnly: serialOnly,
		TCPOnly:    tcpOnly,
		BLEOnly:    bleOnly,
	}

	// Discover devices
	devices, err := discovery.DiscoverMeshtasticDevices(config)
	if err != nil {
		return errors.Wrap(err, "failed to discover devices")
	}

	if len(devices) == 0 {
		fmt.Println("No Meshtastic devices found")
		return nil
	}

	// Display discovered devices
	fmt.Printf("Discovered devices:\n")

	// Group devices by type
	serialDevices := make([]discovery.DiscoveredDevice, 0)
	networkDevices := make([]discovery.DiscoveredDevice, 0)
	bleDevices := make([]discovery.DiscoveredDevice, 0)

	for _, device := range devices {
		switch device.Type {
		case discovery.DeviceTypeSerial:
			serialDevices = append(serialDevices, device)
		case discovery.DeviceTypeNetwork:
			networkDevices = append(networkDevices, device)
		case discovery.DeviceTypeBLE:
			bleDevices = append(bleDevices, device)
		}
	}

	// Display serial devices
	if len(serialDevices) > 0 {
		fmt.Printf("Serial:\n")
		for _, device := range serialDevices {
			nodeID := ""
			if device.NodeID != 0 {
				nodeID = fmt.Sprintf(" (!%08x)", device.NodeID)
			}
			fmt.Printf("  %s - %s%s\n", device.Port, device.Description, nodeID)
		}
	}

	// Display network devices
	if len(networkDevices) > 0 {
		fmt.Printf("Network:\n")
		for _, device := range networkDevices {
			nodeID := ""
			if device.NodeID != 0 {
				nodeID = fmt.Sprintf(" (!%08x)", device.NodeID)
			}
			fmt.Printf("  %s - %s%s\n", device.Host, device.Description, nodeID)
		}
	}

	// Display BLE devices
	if len(bleDevices) > 0 {
		fmt.Printf("BLE:\n")
		for _, device := range bleDevices {
			nodeID := ""
			if device.NodeID != 0 {
				nodeID = fmt.Sprintf(" (!%08x)", device.NodeID)
			}
			fmt.Printf("  %s - %s%s\n", device.Address, device.Description, nodeID)
		}
	}

	return nil
}

func init() {
	discoverCmd.Flags().BoolVar(&serialOnly, "serial-only", false, "Only scan serial ports")
	discoverCmd.Flags().BoolVar(&tcpOnly, "tcp-only", false, "Only scan TCP/IP network")
	discoverCmd.Flags().BoolVar(&bleOnly, "ble-only", false, "Only scan BLE devices")
	discoverCmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "Discovery timeout")
}
