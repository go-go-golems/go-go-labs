package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
)

// deviceCmd represents the device command
var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Device management commands",
	Long: `Device management commands for controlling Meshtastic devices.
Includes reboot, shutdown, factory reset, and configuration operations.`,
}

// deviceRebootCmd represents the device reboot command
var deviceRebootCmd = &cobra.Command{
	Use:   "reboot",
	Short: "Reboot the device",
	Long: `Reboot the connected Meshtastic device.
This will cause the device to restart and reconnect.

Examples:
  meshtastic device reboot
  meshtastic device reboot --confirm`,
	RunE: runDeviceReboot,
}

// deviceShutdownCmd represents the device shutdown command
var deviceShutdownCmd = &cobra.Command{
	Use:   "shutdown",
	Short: "Shutdown the device",
	Long: `Shutdown the connected Meshtastic device.
This will power off the device completely.

Examples:
  meshtastic device shutdown
  meshtastic device shutdown --confirm`,
	RunE: runDeviceShutdown,
}

// deviceFactoryResetCmd represents the device factory-reset command
var deviceFactoryResetCmd = &cobra.Command{
	Use:   "factory-reset",
	Short: "Factory reset the device",
	Long: `Factory reset the connected Meshtastic device.
This will erase all configuration and return the device to default settings.

WARNING: This operation is irreversible!

Examples:
  meshtastic device factory-reset
  meshtastic device factory-reset --confirm`,
	RunE: runDeviceFactoryReset,
}

// deviceSetOwnerCmd represents the device set-owner command
var deviceSetOwnerCmd = &cobra.Command{
	Use:   "set-owner [LONG_NAME] [SHORT_NAME]",
	Short: "Set device owner information",
	Long: `Set the owner information for the connected Meshtastic device.
The long name can be up to 39 characters, and the short name up to 4 characters.

Examples:
  meshtastic device set-owner "John's T-Beam" "J123"
  meshtastic device set-owner --long-name "Station Alpha" --short-name "ALPH"`,
	Args: cobra.RangeArgs(0, 2),
	RunE: runDeviceSetOwner,
}

// deviceSetTimeCmd represents the device set-time command
var deviceSetTimeCmd = &cobra.Command{
	Use:   "set-time [TIMESTAMP]",
	Short: "Set device time",
	Long: `Set the time on the connected Meshtastic device.
If no timestamp is provided, the current system time will be used.

Examples:
  meshtastic device set-time
  meshtastic device set-time 1640995200
  meshtastic device set-time --now`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runDeviceSetTime,
}

// deviceMetadataCmd represents the device metadata command
var deviceMetadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Get device metadata",
	Long: `Get detailed metadata information from the connected Meshtastic device.
This includes hardware information, capabilities, and device-specific data.

Examples:
  meshtastic device metadata
  meshtastic device metadata --json`,
	RunE: runDeviceMetadata,
}

// Device management flags
var (
	deviceConfirm   bool
	deviceOTA       bool
	deviceNow       bool
	deviceLongName  string
	deviceShortName string
	deviceLicensed  bool
)

func runDeviceReboot(cmd *cobra.Command, args []string) error {
	if !deviceConfirm {
		fmt.Print("Are you sure you want to reboot the device? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Reboot cancelled.")
			return nil
		}
	}

	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	log.Info().Msg("Sending reboot command to device")

	// Create admin message for reboot
	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_RebootSeconds{
			RebootSeconds: 1, // Reboot in 1 second
		},
	}

	// Send admin message
	_, err = client.SendAdminMessage(adminMsg)
	if err != nil {
		return errors.Wrap(err, "failed to send reboot command")
	}

	fmt.Println("✓ Reboot command sent successfully")
	fmt.Println("Device will reboot in 1 second...")

	return nil
}

func runDeviceShutdown(cmd *cobra.Command, args []string) error {
	if !deviceConfirm {
		fmt.Print("Are you sure you want to shutdown the device? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Shutdown cancelled.")
			return nil
		}
	}

	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	log.Info().Msg("Sending shutdown command to device")

	// Create admin message for shutdown
	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_ShutdownSeconds{
			ShutdownSeconds: 1, // Shutdown in 1 second
		},
	}

	// Send admin message
	_, err = client.SendAdminMessage(adminMsg)
	if err != nil {
		return errors.Wrap(err, "failed to send shutdown command")
	}

	fmt.Println("✓ Shutdown command sent successfully")
	fmt.Println("Device will shutdown in 1 second...")

	return nil
}

func runDeviceFactoryReset(cmd *cobra.Command, args []string) error {
	if !deviceConfirm {
		fmt.Print("⚠️  WARNING: Factory reset will erase ALL configuration!\n")
		fmt.Print("Are you absolutely sure you want to continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Factory reset cancelled.")
			return nil
		}
	}

	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	log.Info().Msg("Sending factory reset command to device")

	// Create admin message for factory reset
	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_FactoryResetDevice{
			FactoryResetDevice: 1, // Factory reset signal
		},
	}

	// Send admin message
	_, err = client.SendAdminMessage(adminMsg)
	if err != nil {
		return errors.Wrap(err, "failed to send factory reset command")
	}

	fmt.Println("✓ Factory reset command sent successfully")
	fmt.Println("Device will reset to factory defaults...")

	return nil
}

func runDeviceSetOwner(cmd *cobra.Command, args []string) error {
	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	// Parse arguments and flags
	longName := deviceLongName
	shortName := deviceShortName

	if len(args) >= 1 && longName == "" {
		longName = args[0]
	}
	if len(args) >= 2 && shortName == "" {
		shortName = args[1]
	}

	// Validate names
	if longName == "" {
		return errors.New("long name is required")
	}
	if len(longName) > 39 {
		return errors.New("long name must be 39 characters or less")
	}
	if len(shortName) > 4 {
		return errors.New("short name must be 4 characters or less")
	}

	log.Info().Str("long_name", longName).Str("short_name", shortName).Msg("Setting device owner")

	// Create user object
	user := &pb.User{
		LongName:   longName,
		ShortName:  shortName,
		IsLicensed: deviceLicensed,
	}

	// Create admin message
	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_SetOwner{
			SetOwner: user,
		},
	}

	// Send admin message
	_, err = client.SendAdminMessage(adminMsg)
	if err != nil {
		return errors.Wrap(err, "failed to set owner")
	}

	fmt.Printf("✓ Device owner set successfully\n")
	fmt.Printf("  Long Name: %s\n", longName)
	fmt.Printf("  Short Name: %s\n", shortName)
	if deviceLicensed {
		fmt.Printf("  Licensed: Yes\n")
	}

	return nil
}

func runDeviceSetTime(cmd *cobra.Command, args []string) error {
	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	var timestamp uint32

	if deviceNow || len(args) == 0 {
		// Use current time
		timestamp = uint32(time.Now().Unix())
	} else {
		// Parse provided timestamp
		ts, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			return errors.Wrap(err, "invalid timestamp format")
		}
		timestamp = uint32(ts)
	}

	log.Info().Uint32("timestamp", timestamp).Msg("Setting device time")

	// Create admin message
	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_SetTimeOnly{
			SetTimeOnly: timestamp,
		},
	}

	// Send admin message
	_, err = client.SendAdminMessage(adminMsg)
	if err != nil {
		return errors.Wrap(err, "failed to set time")
	}

	timeStr := time.Unix(int64(timestamp), 0).Format(time.RFC3339)
	fmt.Printf("✓ Device time set successfully\n")
	fmt.Printf("  Timestamp: %d\n", timestamp)
	fmt.Printf("  Time: %s\n", timeStr)

	return nil
}

func runDeviceMetadata(cmd *cobra.Command, args []string) error {
	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	log.Info().Msg("Requesting device metadata")

	// Create admin message to get device metadata
	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_GetDeviceMetadataRequest{
			GetDeviceMetadataRequest: true,
		},
	}

	// Send admin message
	response, err := client.SendAdminMessage(adminMsg)
	if err != nil {
		return errors.Wrap(err, "failed to get device metadata")
	}

	// Check if we got a metadata response
	if response.GetGetDeviceMetadataResponse() == nil {
		return errors.New("no metadata response received")
	}

	metadata := response.GetGetDeviceMetadataResponse()

	// Output in requested format
	if outputJSON {
		data, err := marshalJSON(metadata)
		if err != nil {
			return errors.Wrap(err, "failed to marshal JSON")
		}
		fmt.Println(string(data))
		return nil
	}

	if outputYAML {
		data, err := marshalYAML(metadata)
		if err != nil {
			return errors.Wrap(err, "failed to marshal YAML")
		}
		fmt.Print(string(data))
		return nil
	}

	// Default text output
	displayDeviceMetadata(metadata)
	return nil
}

func displayDeviceMetadata(metadata *pb.DeviceMetadata) {
	fmt.Printf("Device Metadata:\n")
	fmt.Printf("  Firmware Version: %s\n", metadata.FirmwareVersion)
	fmt.Printf("  Device State Version: %d\n", metadata.DeviceStateVersion)
	fmt.Printf("  Can Shutdown: %t\n", metadata.CanShutdown)
	fmt.Printf("  Has WiFi: %t\n", metadata.HasWifi)
	fmt.Printf("  Has Bluetooth: %t\n", metadata.HasBluetooth)
	fmt.Printf("  Has Ethernet: %t\n", metadata.HasEthernet)
	fmt.Printf("  Role: %s\n", metadata.Role.String())
	fmt.Printf("  Position Flags: %d\n", metadata.PositionFlags)
	fmt.Printf("  Hardware Model: %s\n", metadata.HwModel.String())
	fmt.Printf("  Has Remote Hardware: %t\n", metadata.HasRemoteHardware)
}

func init() {
	// Add device management flags
	deviceRebootCmd.Flags().BoolVar(&deviceConfirm, "confirm", false, "Skip confirmation prompt")
	deviceRebootCmd.Flags().BoolVar(&deviceOTA, "ota", false, "Reboot into OTA mode")

	deviceShutdownCmd.Flags().BoolVar(&deviceConfirm, "confirm", false, "Skip confirmation prompt")

	deviceFactoryResetCmd.Flags().BoolVar(&deviceConfirm, "confirm", false, "Skip confirmation prompt")

	deviceSetOwnerCmd.Flags().StringVar(&deviceLongName, "long-name", "", "Full device name (up to 39 characters)")
	deviceSetOwnerCmd.Flags().StringVar(&deviceShortName, "short-name", "", "Short device name (up to 4 characters)")
	deviceSetOwnerCmd.Flags().BoolVar(&deviceLicensed, "licensed", false, "Mark as licensed operator")

	deviceSetTimeCmd.Flags().BoolVar(&deviceNow, "now", false, "Use current system time")

	deviceMetadataCmd.Flags().BoolVar(&outputJSON, "json", false, "Output in JSON format")
	deviceMetadataCmd.Flags().BoolVar(&outputYAML, "yaml", false, "Output in YAML format")

	// Add subcommands to device
	deviceCmd.AddCommand(deviceRebootCmd)
	deviceCmd.AddCommand(deviceShutdownCmd)
	deviceCmd.AddCommand(deviceFactoryResetCmd)
	deviceCmd.AddCommand(deviceSetOwnerCmd)
	deviceCmd.AddCommand(deviceSetTimeCmd)
	deviceCmd.AddCommand(deviceMetadataCmd)

	// Add device command to root
	rootCmd.AddCommand(deviceCmd)
}
