package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/client"
	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
)

var simpleConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long:  `Manage device configuration including get, set, export, and import operations.`,
}

var simpleConfigGetCmd = &cobra.Command{
	Use:   "get [FIELD]",
	Short: "Get configuration values",
	Long:  `Get configuration values from the device.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSimpleConfigGet,
}

var simpleConfigSetCmd = &cobra.Command{
	Use:   "set FIELD VALUE",
	Short: "Set configuration values",
	Long:  `Set configuration values on the device.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runSimpleConfigSet,
}

var simpleConfigExportCmd = &cobra.Command{
	Use:   "export [FILENAME]",
	Short: "Export configuration to file",
	Long:  `Export device configuration to a file.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSimpleConfigExport,
}

// Simple config command variables
var (
	simpleConfigAll    bool
	simpleConfigFormat string
	simpleConfigStdout bool
)

func init() {
	simpleConfigCmd.AddCommand(simpleConfigGetCmd)
	simpleConfigCmd.AddCommand(simpleConfigSetCmd)
	simpleConfigCmd.AddCommand(simpleConfigExportCmd)

	simpleConfigGetCmd.Flags().BoolVar(&outputJSON, "json", false, "Output in JSON format")
	simpleConfigGetCmd.Flags().BoolVar(&outputYAML, "yaml", false, "Output in YAML format")
	simpleConfigGetCmd.Flags().BoolVar(&simpleConfigAll, "all", false, "Get all configuration")

	simpleConfigExportCmd.Flags().StringVar(&simpleConfigFormat, "format", "yaml", "Output format (yaml, json)")
	simpleConfigExportCmd.Flags().BoolVar(&simpleConfigStdout, "stdout", false, "Output to stdout")
}

func runSimpleConfigGet(cmd *cobra.Command, args []string) error {
	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	config := client.GetConfig()
	if config == nil {
		return errors.New("failed to get device configuration")
	}

	if simpleConfigAll {
		return displaySimpleConfig(config)
	}

	if len(args) == 0 {
		return errors.New("field name required when not using --all")
	}

	field := args[0]
	return displaySimpleConfigField(config, field)
}

func runSimpleConfigSet(cmd *cobra.Command, args []string) error {
	field := args[0]
	value := args[1]

	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	if err := setSimpleConfigField(client, field, value); err != nil {
		return errors.Wrap(err, "failed to set configuration field")
	}

	fmt.Printf("✓ Configuration updated successfully\n")
	return nil
}

func runSimpleConfigExport(cmd *cobra.Command, args []string) error {
	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	config := client.GetConfig()
	if config == nil {
		return errors.New("failed to get device configuration")
	}

	exportData := buildSimpleExportData(config)

	var output []byte
	switch strings.ToLower(simpleConfigFormat) {
	case "json":
		output, err = json.MarshalIndent(exportData, "", "  ")
	case "yaml":
		output, err = yaml.Marshal(exportData)
	default:
		return errors.Errorf("unsupported format: %s", simpleConfigFormat)
	}

	if err != nil {
		return errors.Wrap(err, "failed to format output")
	}

	if simpleConfigStdout || len(args) == 0 {
		fmt.Print(string(output))
		return nil
	}

	filename := args[0]
	if err := os.WriteFile(filename, output, 0644); err != nil {
		return errors.Wrap(err, "failed to write file")
	}

	fmt.Printf("Configuration exported to %s\n", filename)
	return nil
}

func displaySimpleConfig(config *pb.LocalConfig) error {
	data := buildSimpleExportData(config)

	if outputJSON {
		output, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	}

	if outputYAML {
		output, err := yaml.Marshal(data)
		if err != nil {
			return err
		}
		fmt.Print(string(output))
		return nil
	}

	fmt.Printf("Device Configuration:\n")
	if config.GetDevice() != nil {
		fmt.Printf("  Device Role: %s\n", config.GetDevice().GetRole())
	}
	if config.GetLora() != nil {
		fmt.Printf("  LoRa Region: %s\n", config.GetLora().GetRegion())
		fmt.Printf("  LoRa Modem Preset: %s\n", config.GetLora().GetModemPreset())
	}
	if config.GetBluetooth() != nil {
		fmt.Printf("  Bluetooth Enabled: %v\n", config.GetBluetooth().GetEnabled())
	}

	return nil
}

func displaySimpleConfigField(config *pb.LocalConfig, field string) error {
	parts := strings.Split(field, ".")
	if len(parts) != 2 {
		return errors.New("field must be in format 'section.field'")
	}

	section := parts[0]
	fieldName := parts[1]

	var value interface{}
	switch section {
	case "device":
		deviceConfig := config.GetDevice()
		if deviceConfig == nil {
			return errors.New("device config not available")
		}
		switch fieldName {
		case "role":
			value = deviceConfig.GetRole().String()
		case "serial_enabled":
			value = deviceConfig.GetSerialEnabled()
		case "button_gpio":
			value = deviceConfig.GetButtonGpio()
		case "buzzer_gpio":
			value = deviceConfig.GetBuzzerGpio()
		case "disable_triple_click":
			value = deviceConfig.GetDisableTripleClick()
		case "tzdef":
			value = deviceConfig.GetTzdef()
		case "led_heartbeat_disabled":
			value = deviceConfig.GetLedHeartbeatDisabled()
		default:
			return errors.Errorf("unsupported device field: %s", fieldName)
		}
	case "lora":
		loraConfig := config.GetLora()
		if loraConfig == nil {
			return errors.New("lora config not available")
		}
		switch fieldName {
		case "region":
			value = loraConfig.GetRegion().String()
		case "modem_preset":
			value = loraConfig.GetModemPreset().String()
		case "hop_limit":
			value = loraConfig.GetHopLimit()
		case "tx_enabled":
			value = loraConfig.GetTxEnabled()
		case "tx_power":
			value = loraConfig.GetTxPower()
		default:
			return errors.Errorf("unsupported lora field: %s", fieldName)
		}
	case "bluetooth":
		bluetoothConfig := config.GetBluetooth()
		if bluetoothConfig == nil {
			return errors.New("bluetooth config not available")
		}
		switch fieldName {
		case "enabled":
			value = bluetoothConfig.GetEnabled()
		case "mode":
			value = bluetoothConfig.GetMode().String()
		case "fixed_pin":
			value = bluetoothConfig.GetFixedPin()
		default:
			return errors.Errorf("unsupported bluetooth field: %s", fieldName)
		}
	default:
		return errors.Errorf("unsupported section: %s", section)
	}

	if outputJSON {
		result := map[string]interface{}{field: value}
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	}

	if outputYAML {
		result := map[string]interface{}{field: value}
		output, err := yaml.Marshal(result)
		if err != nil {
			return err
		}
		fmt.Print(string(output))
		return nil
	}

	fmt.Printf("%s: %v\n", field, value)
	return nil
}

func setSimpleConfigField(client *client.RobustMeshtasticClient, field, value string) error {
	parts := strings.Split(field, ".")
	if len(parts) != 2 {
		return errors.New("field must be in format 'section.field'")
	}

	section := parts[0]
	fieldName := parts[1]

	config := client.GetConfig()
	if config == nil {
		return errors.New("failed to get current configuration")
	}

	switch section {
	case "device":
		return setSimpleDeviceConfig(client, config, fieldName, value)
	case "lora":
		return setSimpleLoraConfig(client, config, fieldName, value)
	case "bluetooth":
		return setSimpleBluetoothConfig(client, config, fieldName, value)
	default:
		return errors.Errorf("unsupported section: %s", section)
	}
}

func setSimpleDeviceConfig(client *client.RobustMeshtasticClient, config *pb.LocalConfig, fieldName, value string) error {
	deviceConfig := config.GetDevice()
	if deviceConfig == nil {
		deviceConfig = &pb.Config_DeviceConfig{}
	}

	switch fieldName {
	case "role":
		roleValue, err := parseRole(value)
		if err != nil {
			return err
		}
		deviceConfig.Role = roleValue
	case "serial_enabled":
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Wrap(err, "invalid boolean value")
		}
		deviceConfig.SerialEnabled = boolValue
	case "button_gpio":
		intValue, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return errors.Wrap(err, "invalid integer value")
		}
		deviceConfig.ButtonGpio = uint32(intValue)
	case "buzzer_gpio":
		intValue, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return errors.Wrap(err, "invalid integer value")
		}
		deviceConfig.BuzzerGpio = uint32(intValue)
	case "disable_triple_click":
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Wrap(err, "invalid boolean value")
		}
		deviceConfig.DisableTripleClick = boolValue
	case "tzdef":
		deviceConfig.Tzdef = value
	case "led_heartbeat_disabled":
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Wrap(err, "invalid boolean value")
		}
		deviceConfig.LedHeartbeatDisabled = boolValue
	default:
		return errors.Errorf("unsupported device field: %s", fieldName)
	}

	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_SetConfig{
			SetConfig: &pb.Config{
				PayloadVariant: &pb.Config_Device{Device: deviceConfig},
			},
		},
	}

	_, err := client.SendAdminMessage(adminMsg)
	return err
}

func setSimpleLoraConfig(client *client.RobustMeshtasticClient, config *pb.LocalConfig, fieldName, value string) error {
	loraConfig := config.GetLora()
	if loraConfig == nil {
		loraConfig = &pb.Config_LoRaConfig{}
	}

	switch fieldName {
	case "region":
		regionValue, err := parseRegion(value)
		if err != nil {
			return err
		}
		loraConfig.Region = regionValue
	case "modem_preset":
		presetValue, err := parseModemPreset(value)
		if err != nil {
			return err
		}
		loraConfig.ModemPreset = presetValue
	case "hop_limit":
		intValue, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return errors.Wrap(err, "invalid integer value")
		}
		loraConfig.HopLimit = uint32(intValue)
	case "tx_enabled":
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Wrap(err, "invalid boolean value")
		}
		loraConfig.TxEnabled = boolValue
	case "tx_power":
		intValue, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return errors.Wrap(err, "invalid integer value")
		}
		loraConfig.TxPower = int32(intValue)
	default:
		return errors.Errorf("unsupported lora field: %s", fieldName)
	}

	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_SetConfig{
			SetConfig: &pb.Config{
				PayloadVariant: &pb.Config_Lora{Lora: loraConfig},
			},
		},
	}

	_, err := client.SendAdminMessage(adminMsg)
	return err
}

func setSimpleBluetoothConfig(client *client.RobustMeshtasticClient, config *pb.LocalConfig, fieldName, value string) error {
	bluetoothConfig := config.GetBluetooth()
	if bluetoothConfig == nil {
		bluetoothConfig = &pb.Config_BluetoothConfig{}
	}

	switch fieldName {
	case "enabled":
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Wrap(err, "invalid boolean value")
		}
		bluetoothConfig.Enabled = boolValue
	case "mode":
		modeValue, err := parseBluetoothMode(value)
		if err != nil {
			return err
		}
		bluetoothConfig.Mode = modeValue
	case "fixed_pin":
		intValue, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return errors.Wrap(err, "invalid integer value")
		}
		bluetoothConfig.FixedPin = uint32(intValue)
	default:
		return errors.Errorf("unsupported bluetooth field: %s", fieldName)
	}

	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_SetConfig{
			SetConfig: &pb.Config{
				PayloadVariant: &pb.Config_Bluetooth{Bluetooth: bluetoothConfig},
			},
		},
	}

	_, err := client.SendAdminMessage(adminMsg)
	return err
}

func buildSimpleExportData(config *pb.LocalConfig) map[string]interface{} {
	result := make(map[string]interface{})

	if config.GetDevice() != nil {
		result["device"] = map[string]interface{}{
			"role":                   config.GetDevice().GetRole().String(),
			"serial_enabled":         config.GetDevice().GetSerialEnabled(),
			"button_gpio":            config.GetDevice().GetButtonGpio(),
			"buzzer_gpio":            config.GetDevice().GetBuzzerGpio(),
			"disable_triple_click":   config.GetDevice().GetDisableTripleClick(),
			"tzdef":                  config.GetDevice().GetTzdef(),
			"led_heartbeat_disabled": config.GetDevice().GetLedHeartbeatDisabled(),
		}
	}

	if config.GetLora() != nil {
		result["lora"] = map[string]interface{}{
			"region":       config.GetLora().GetRegion().String(),
			"modem_preset": config.GetLora().GetModemPreset().String(),
			"hop_limit":    config.GetLora().GetHopLimit(),
			"tx_enabled":   config.GetLora().GetTxEnabled(),
			"tx_power":     config.GetLora().GetTxPower(),
		}
	}

	if config.GetBluetooth() != nil {
		result["bluetooth"] = map[string]interface{}{
			"enabled":   config.GetBluetooth().GetEnabled(),
			"mode":      config.GetBluetooth().GetMode().String(),
			"fixed_pin": config.GetBluetooth().GetFixedPin(),
		}
	}

	return result
}

// Helper parsing functions
func parseRole(value string) (pb.Config_DeviceConfig_Role, error) {
	switch strings.ToUpper(value) {
	case "CLIENT":
		return pb.Config_DeviceConfig_CLIENT, nil
	case "CLIENT_MUTE":
		return pb.Config_DeviceConfig_CLIENT_MUTE, nil
	case "ROUTER":
		return pb.Config_DeviceConfig_ROUTER, nil
	case "ROUTER_CLIENT":
		return pb.Config_DeviceConfig_ROUTER_CLIENT, nil
	case "REPEATER":
		return pb.Config_DeviceConfig_REPEATER, nil
	case "TRACKER":
		return pb.Config_DeviceConfig_TRACKER, nil
	case "SENSOR":
		return pb.Config_DeviceConfig_SENSOR, nil
	case "TAK":
		return pb.Config_DeviceConfig_TAK, nil
	case "CLIENT_HIDDEN":
		return pb.Config_DeviceConfig_CLIENT_HIDDEN, nil
	case "LOST_AND_FOUND":
		return pb.Config_DeviceConfig_LOST_AND_FOUND, nil
	case "TAK_TRACKER":
		return pb.Config_DeviceConfig_TAK_TRACKER, nil
	default:
		return 0, errors.Errorf("invalid role: %s", value)
	}
}

func parseRegion(value string) (pb.Config_LoRaConfig_RegionCode, error) {
	switch strings.ToUpper(value) {
	case "US":
		return pb.Config_LoRaConfig_US, nil
	case "EU_433":
		return pb.Config_LoRaConfig_EU_433, nil
	case "EU_868":
		return pb.Config_LoRaConfig_EU_868, nil
	case "CN":
		return pb.Config_LoRaConfig_CN, nil
	case "JP":
		return pb.Config_LoRaConfig_JP, nil
	case "ANZ":
		return pb.Config_LoRaConfig_ANZ, nil
	case "KR":
		return pb.Config_LoRaConfig_KR, nil
	case "TW":
		return pb.Config_LoRaConfig_TW, nil
	case "RU":
		return pb.Config_LoRaConfig_RU, nil
	case "IN":
		return pb.Config_LoRaConfig_IN, nil
	case "NZ_865":
		return pb.Config_LoRaConfig_NZ_865, nil
	case "TH":
		return pb.Config_LoRaConfig_TH, nil
	case "LORA_24":
		return pb.Config_LoRaConfig_LORA_24, nil
	case "UA_433":
		return pb.Config_LoRaConfig_UA_433, nil
	case "UA_868":
		return pb.Config_LoRaConfig_UA_868, nil
	case "MY_433":
		return pb.Config_LoRaConfig_MY_433, nil
	case "MY_919":
		return pb.Config_LoRaConfig_MY_919, nil
	case "SG_923":
		return pb.Config_LoRaConfig_SG_923, nil
	default:
		return 0, errors.Errorf("invalid region: %s", value)
	}
}

func parseModemPreset(value string) (pb.Config_LoRaConfig_ModemPreset, error) {
	switch strings.ToUpper(value) {
	case "LONG_FAST":
		return pb.Config_LoRaConfig_LONG_FAST, nil
	case "LONG_SLOW":
		return pb.Config_LoRaConfig_LONG_SLOW, nil
	case "VERY_LONG_SLOW":
		return pb.Config_LoRaConfig_VERY_LONG_SLOW, nil
	case "MEDIUM_SLOW":
		return pb.Config_LoRaConfig_MEDIUM_SLOW, nil
	case "MEDIUM_FAST":
		return pb.Config_LoRaConfig_MEDIUM_FAST, nil
	case "SHORT_SLOW":
		return pb.Config_LoRaConfig_SHORT_SLOW, nil
	case "SHORT_FAST":
		return pb.Config_LoRaConfig_SHORT_FAST, nil
	case "LONG_MODERATE":
		return pb.Config_LoRaConfig_LONG_MODERATE, nil
	case "SHORT_TURBO":
		return pb.Config_LoRaConfig_SHORT_TURBO, nil
	default:
		return 0, errors.Errorf("invalid modem preset: %s", value)
	}
}

func parseBluetoothMode(value string) (pb.Config_BluetoothConfig_PairingMode, error) {
	switch strings.ToUpper(value) {
	case "RANDOM_PIN":
		return pb.Config_BluetoothConfig_RANDOM_PIN, nil
	case "FIXED_PIN":
		return pb.Config_BluetoothConfig_FIXED_PIN, nil
	case "NO_PIN":
		return pb.Config_BluetoothConfig_NO_PIN, nil
	default:
		return 0, errors.Errorf("invalid bluetooth mode: %s", value)
	}
}
