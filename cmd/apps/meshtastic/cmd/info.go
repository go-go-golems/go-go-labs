package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/client"
	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
)

var (
	outputJSON bool
	outputYAML bool
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display comprehensive device information",
	Long: `Display comprehensive device information including hardware details, 
firmware version, network status, and more.

Examples:
  meshtastic info
  meshtastic info --json
  meshtastic info --yaml`,
	RunE: runInfo,
}

type DeviceInfo struct {
	NodeID          string           `json:"node_id" yaml:"node_id"`
	User            *UserInfo        `json:"user,omitempty" yaml:"user,omitempty"`
	Hardware        string           `json:"hardware" yaml:"hardware"`
	Firmware        string           `json:"firmware" yaml:"firmware"`
	Region          string           `json:"region,omitempty" yaml:"region,omitempty"`
	ModemPreset     string           `json:"modem_preset,omitempty" yaml:"modem_preset,omitempty"`
	HardwareDetails *HardwareDetails `json:"hardware_details,omitempty" yaml:"hardware_details,omitempty"`
	Network         *NetworkInfo     `json:"network,omitempty" yaml:"network,omitempty"`
}

type UserInfo struct {
	LongName  string `json:"long_name" yaml:"long_name"`
	ShortName string `json:"short_name" yaml:"short_name"`
	Licensed  bool   `json:"licensed,omitempty" yaml:"licensed,omitempty"`
}

type HardwareDetails struct {
	Battery            string `json:"battery,omitempty" yaml:"battery,omitempty"`
	Voltage            string `json:"voltage,omitempty" yaml:"voltage,omitempty"`
	ChannelUtilization string `json:"channel_utilization,omitempty" yaml:"channel_utilization,omitempty"`
	AirTime            string `json:"air_time,omitempty" yaml:"air_time,omitempty"`
	Uptime             string `json:"uptime,omitempty" yaml:"uptime,omitempty"`
	Temperature        string `json:"temperature,omitempty" yaml:"temperature,omitempty"`
}

type NetworkInfo struct {
	MeshID     string `json:"mesh_id" yaml:"mesh_id"`
	NodesCount int    `json:"nodes_count" yaml:"nodes_count"`
	Channels   int    `json:"channels" yaml:"channels"`
}

func runInfo(cmd *cobra.Command, args []string) error {
	// Create client and connect
	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	// Get device information
	deviceInfo, err := gatherDeviceInfo(client)
	if err != nil {
		return errors.Wrap(err, "failed to gather device information")
	}

	// Output in requested format
	if outputJSON {
		data, err := json.MarshalIndent(deviceInfo, "", "  ")
		if err != nil {
			return errors.Wrap(err, "failed to marshal JSON")
		}
		fmt.Println(string(data))
		return nil
	}

	if outputYAML {
		data, err := yaml.Marshal(deviceInfo)
		if err != nil {
			return errors.Wrap(err, "failed to marshal YAML")
		}
		fmt.Print(string(data))
		return nil
	}

	// Default text output
	displayDeviceInfo(deviceInfo)
	return nil
}

func createAndConnectClient() (*client.RobustMeshtasticClient, error) {
	config := &client.Config{
		DevicePath:  globalConfig.Port,
		Timeout:     globalConfig.Timeout,
		DebugSerial: globalConfig.DebugSerial,
		HexDump:     globalConfig.HexDump,
	}

	meshtasticClient, err := client.NewRobustMeshtasticClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create robust client")
	}

	// Connect to the device
	ctx, cancel := context.WithTimeout(context.Background(), globalConfig.Timeout)
	defer cancel()

	if err := meshtasticClient.Connect(ctx); err != nil {
		meshtasticClient.Close()
		return nil, errors.Wrap(err, "failed to connect to device")
	}

	if !meshtasticClient.IsConnected() {
		meshtasticClient.Close()
		return nil, errors.New("device connection failed")
	}

	return meshtasticClient, nil
}

func gatherDeviceInfo(client *client.RobustMeshtasticClient) (*DeviceInfo, error) {
	// Get basic device information
	myInfo := client.GetMyInfo()
	if myInfo == nil {
		return nil, errors.New("failed to get device info")
	}

	nodes := client.GetNodes()
	channels := client.GetChannels()
	config := client.GetConfig()

	// Build device info structure
	deviceInfo := &DeviceInfo{
		NodeID:   fmt.Sprintf("!%08x", myInfo.MyNodeNum),
		Hardware: "Unknown",
		Firmware: "Unknown",
		Network: &NetworkInfo{
			MeshID:     fmt.Sprintf("!%08x", myInfo.MyNodeNum),
			NodesCount: len(nodes),
			Channels:   len(channels),
		},
	}

	// Get user info from our node
	if node, exists := nodes[myInfo.MyNodeNum]; exists {
		user := node.GetUser()
		if user != nil {
			deviceInfo.User = &UserInfo{
				LongName:  user.LongName,
				ShortName: user.ShortName,
				Licensed:  user.IsLicensed,
			}
			if user.HwModel != pb.HardwareModel_UNSET {
				deviceInfo.Hardware = user.HwModel.String()
			}
		}
	}

	// Get region and modem preset from config
	if config != nil {
		lora := config.GetLora()
		if lora != nil {
			deviceInfo.Region = lora.Region.String()
			deviceInfo.ModemPreset = lora.ModemPreset.String()
		}
	}

	return deviceInfo, nil
}

func displayDeviceInfo(info *DeviceInfo) {
	fmt.Printf("Device Information:\n")
	fmt.Printf("  Node ID: %s\n", info.NodeID)

	if info.User != nil {
		fmt.Printf("  User: %s (%s)\n", info.User.LongName, info.User.ShortName)
	}

	fmt.Printf("  Hardware: %s\n", info.Hardware)
	fmt.Printf("  Firmware: %s\n", info.Firmware)

	if info.Region != "" {
		fmt.Printf("  Region: %s\n", info.Region)
	}

	if info.ModemPreset != "" {
		fmt.Printf("  Modem Preset: %s\n", info.ModemPreset)
	}

	if info.HardwareDetails != nil {
		fmt.Printf("\nHardware Details:\n")
		if info.HardwareDetails.Battery != "" {
			fmt.Printf("  Battery: %s\n", info.HardwareDetails.Battery)
		}
		if info.HardwareDetails.Voltage != "" {
			fmt.Printf("  Voltage: %s\n", info.HardwareDetails.Voltage)
		}
		if info.HardwareDetails.ChannelUtilization != "" {
			fmt.Printf("  Channel Utilization: %s\n", info.HardwareDetails.ChannelUtilization)
		}
		if info.HardwareDetails.AirTime != "" {
			fmt.Printf("  Air Time: %s\n", info.HardwareDetails.AirTime)
		}
	}

	if info.Network != nil {
		fmt.Printf("\nNetwork:\n")
		fmt.Printf("  Mesh ID: %s\n", info.Network.MeshID)
		fmt.Printf("  Nodes in mesh: %d\n", info.Network.NodesCount)
		fmt.Printf("  Channels: %d\n", info.Network.Channels)
	}
}

// Helper functions for marshaling
func marshalJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func marshalYAML(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func init() {
	infoCmd.Flags().BoolVar(&outputJSON, "json", false, "Output in JSON format")
	infoCmd.Flags().BoolVar(&outputYAML, "yaml", false, "Output in YAML format")
}
