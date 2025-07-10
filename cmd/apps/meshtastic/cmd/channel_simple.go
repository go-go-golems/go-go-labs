package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
)

var simpleChannelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Channel management commands",
	Long:  `Manage device channels including list, add, delete, and modify operations.`,
}

var simpleChannelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured channels",
	Long:  `List all configured channels with their settings.`,
	RunE:  runSimpleChannelList,
}

var simpleChannelAddCmd = &cobra.Command{
	Use:   "add NAME",
	Short: "Add a new channel",
	Long:  `Add a new channel with the specified name.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSimpleChannelAdd,
}

var simpleChannelDeleteCmd = &cobra.Command{
	Use:   "delete INDEX",
	Short: "Delete a channel",
	Long:  `Delete a channel by index.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSimpleChannelDelete,
}

var simpleChannelSetCmd = &cobra.Command{
	Use:   "set INDEX FIELD VALUE",
	Short: "Set channel parameters",
	Long:  `Set channel parameters for the specified channel.`,
	Args:  cobra.ExactArgs(3),
	RunE:  runSimpleChannelSet,
}

var simpleChannelEnableCmd = &cobra.Command{
	Use:   "enable INDEX",
	Short: "Enable a channel",
	Long:  `Enable a channel by index.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSimpleChannelEnable,
}

var simpleChannelDisableCmd = &cobra.Command{
	Use:   "disable INDEX",
	Short: "Disable a channel",
	Long:  `Disable a channel by index.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSimpleChannelDisable,
}

// Simple channel command variables
var (
	simpleChannelShowKeys bool
	simpleChannelPSK      string
	simpleChannelRole     string
	simpleChannelIndex    int
)

func init() {
	simpleChannelCmd.AddCommand(simpleChannelListCmd)
	simpleChannelCmd.AddCommand(simpleChannelAddCmd)
	simpleChannelCmd.AddCommand(simpleChannelDeleteCmd)
	simpleChannelCmd.AddCommand(simpleChannelSetCmd)
	simpleChannelCmd.AddCommand(simpleChannelEnableCmd)
	simpleChannelCmd.AddCommand(simpleChannelDisableCmd)

	simpleChannelListCmd.Flags().BoolVar(&outputJSON, "json", false, "Output in JSON format")
	simpleChannelListCmd.Flags().BoolVar(&simpleChannelShowKeys, "show-keys", false, "Show encryption keys")

	simpleChannelAddCmd.Flags().StringVar(&simpleChannelPSK, "psk", "", "Pre-shared key (base64)")
	simpleChannelAddCmd.Flags().StringVar(&simpleChannelRole, "role", "SECONDARY", "Channel role")
	simpleChannelAddCmd.Flags().IntVar(&simpleChannelIndex, "index", -1, "Channel index")
}

func runSimpleChannelList(cmd *cobra.Command, args []string) error {
	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	channels := client.GetChannels()
	if channels == nil {
		return errors.New("failed to get channels")
	}

	type SimpleChannelInfo struct {
		Index    int32               `json:"index"`
		Name     string              `json:"name"`
		Role     string              `json:"role"`
		PSK      string              `json:"psk,omitempty"`
		Settings *pb.ChannelSettings `json:"settings,omitempty"`
	}

	var channelList []SimpleChannelInfo
	for index, channel := range channels {
		info := SimpleChannelInfo{
			Index:    int32(index),
			Name:     channel.GetSettings().GetName(),
			Role:     channel.GetRole().String(),
			Settings: channel.GetSettings(),
		}

		if info.Name == "" {
			info.Name = "Default"
		}

		if simpleChannelShowKeys && len(channel.GetSettings().GetPsk()) > 0 {
			info.PSK = base64.StdEncoding.EncodeToString(channel.GetSettings().GetPsk())
		} else if len(channel.GetSettings().GetPsk()) == 0 {
			info.PSK = "none"
		} else {
			info.PSK = "***encrypted***"
		}

		channelList = append(channelList, info)
	}

	if outputJSON {
		output, err := json.MarshalIndent(channelList, "", "  ")
		if err != nil {
			return errors.Wrap(err, "failed to marshal JSON")
		}
		fmt.Println(string(output))
		return nil
	}

	fmt.Printf("Channels:\n")
	fmt.Printf("┌─────┬─────────────┬─────────┬─────────────────┐\n")
	fmt.Printf("│ ID  │ Name        │ Role    │ PSK             │\n")
	fmt.Printf("├─────┼─────────────┼─────────┼─────────────────┤\n")

	for _, channel := range channelList {
		fmt.Printf("│ %-3d │ %-11s │ %-7s │ %-15s │\n",
			channel.Index,
			truncateString(channel.Name, 11),
			channel.Role,
			truncateString(channel.PSK, 15))
	}

	fmt.Printf("└─────┴─────────────┴─────────┴─────────────────┘\n")
	return nil
}

func runSimpleChannelAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	channels := client.GetChannels()
	if channels == nil {
		return errors.New("failed to get current channels")
	}

	// Find next available index
	index := simpleChannelIndex
	if index == -1 {
		index = findNextAvailableChannel(channels)
		if index == -1 {
			return errors.New("no available channel slots")
		}
	}

	if index < 0 || index > 7 {
		return errors.New("channel index must be between 0 and 7")
	}

	// Parse role
	role, err := parseChannelRole(simpleChannelRole)
	if err != nil {
		return err
	}

	// Parse PSK
	var psk []byte
	if simpleChannelPSK != "" {
		psk, err = base64.StdEncoding.DecodeString(simpleChannelPSK)
		if err != nil {
			return errors.Wrap(err, "invalid PSK format")
		}
	}

	// Create channel settings
	settings := &pb.ChannelSettings{
		Name: name,
		Psk:  psk,
		Id:   generateChannelID(),
	}

	// Create channel
	channel := &pb.Channel{
		Index:    int32(index),
		Settings: settings,
		Role:     role,
	}

	// Send admin message
	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_SetChannel{
			SetChannel: channel,
		},
	}

	_, err = client.SendAdminMessage(adminMsg)
	if err != nil {
		return errors.Wrap(err, "failed to add channel")
	}

	fmt.Printf("✓ Channel %d '%s' added successfully\n", index, name)
	return nil
}

func runSimpleChannelDelete(cmd *cobra.Command, args []string) error {
	index, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.Wrap(err, "invalid channel index")
	}

	if index < 0 || index > 7 {
		return errors.New("channel index must be between 0 and 7")
	}

	if index == 0 {
		return errors.New("cannot delete primary channel")
	}

	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	// Create empty channel to delete
	channel := &pb.Channel{
		Index:    int32(index),
		Settings: &pb.ChannelSettings{},
		Role:     pb.Channel_DISABLED,
	}

	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_SetChannel{
			SetChannel: channel,
		},
	}

	_, err = client.SendAdminMessage(adminMsg)
	if err != nil {
		return errors.Wrap(err, "failed to delete channel")
	}

	fmt.Printf("✓ Channel %d deleted successfully\n", index)
	return nil
}

func runSimpleChannelSet(cmd *cobra.Command, args []string) error {
	index, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.Wrap(err, "invalid channel index")
	}

	if index < 0 || index > 7 {
		return errors.New("channel index must be between 0 and 7")
	}

	field := args[1]
	value := args[2]

	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	channels := client.GetChannels()
	if channels == nil {
		return errors.New("failed to get current channels")
	}

	channel, exists := channels[uint32(index)]
	if !exists {
		return errors.Errorf("channel %d does not exist", index)
	}

	// Clone channel settings
	settings := &pb.ChannelSettings{
		Name:            channel.Settings.Name,
		Psk:             channel.Settings.Psk,
		Id:              channel.Settings.Id,
		UplinkEnabled:   channel.Settings.UplinkEnabled,
		DownlinkEnabled: channel.Settings.DownlinkEnabled,
	}

	// Update field
	switch field {
	case "name":
		settings.Name = value
	case "psk":
		psk, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return errors.Wrap(err, "invalid PSK format")
		}
		settings.Psk = psk
	case "role":
		role, err := parseChannelRole(value)
		if err != nil {
			return err
		}
		channel.Role = role
	case "uplink_enabled":
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Wrap(err, "invalid boolean value")
		}
		settings.UplinkEnabled = boolValue
	case "downlink_enabled":
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Wrap(err, "invalid boolean value")
		}
		settings.DownlinkEnabled = boolValue
	default:
		return errors.Errorf("unsupported field: %s", field)
	}

	// Update channel
	channel.Settings = settings

	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_SetChannel{
			SetChannel: channel,
		},
	}

	_, err = client.SendAdminMessage(adminMsg)
	if err != nil {
		return errors.Wrap(err, "failed to update channel")
	}

	fmt.Printf("✓ Channel %d %s updated successfully\n", index, field)
	return nil
}

func runSimpleChannelEnable(cmd *cobra.Command, args []string) error {
	index, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.Wrap(err, "invalid channel index")
	}

	role := pb.Channel_SECONDARY
	if index == 0 {
		role = pb.Channel_PRIMARY
	}

	return setSimpleChannelRole(index, role)
}

func runSimpleChannelDisable(cmd *cobra.Command, args []string) error {
	index, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.Wrap(err, "invalid channel index")
	}

	if index == 0 {
		return errors.New("cannot disable primary channel")
	}

	return setSimpleChannelRole(index, pb.Channel_DISABLED)
}

func setSimpleChannelRole(index int, role pb.Channel_Role) error {
	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	channels := client.GetChannels()
	if channels == nil {
		return errors.New("failed to get current channels")
	}

	channel, exists := channels[uint32(index)]
	if !exists {
		return errors.Errorf("channel %d does not exist", index)
	}

	channel.Role = role

	adminMsg := &pb.AdminMessage{
		PayloadVariant: &pb.AdminMessage_SetChannel{
			SetChannel: channel,
		},
	}

	_, err = client.SendAdminMessage(adminMsg)
	if err != nil {
		return errors.Wrap(err, "failed to update channel role")
	}

	action := "enabled"
	if role == pb.Channel_DISABLED {
		action = "disabled"
	}

	fmt.Printf("✓ Channel %d %s successfully\n", index, action)
	return nil
}

// Helper functions
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func findNextAvailableChannel(channels map[uint32]*pb.Channel) int {
	for i := 1; i <= 7; i++ {
		if _, exists := channels[uint32(i)]; !exists {
			return i
		}
	}
	return -1
}

func generateChannelID() uint32 {
	// Generate a random 32-bit ID
	var id uint32
	for {
		randBytes := make([]byte, 4)
		rand.Read(randBytes)
		id = uint32(randBytes[0])<<24 | uint32(randBytes[1])<<16 | uint32(randBytes[2])<<8 | uint32(randBytes[3])
		if id != 0 {
			break
		}
	}
	return id
}

func parseChannelRole(role string) (pb.Channel_Role, error) {
	switch strings.ToUpper(role) {
	case "DISABLED":
		return pb.Channel_DISABLED, nil
	case "PRIMARY":
		return pb.Channel_PRIMARY, nil
	case "SECONDARY":
		return pb.Channel_SECONDARY, nil
	default:
		return 0, errors.Errorf("invalid channel role: %s", role)
	}
}
