package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sort"
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
	showFields string
	sortBy     string
	liveMode   bool
)

// nodesCmd represents the nodes command
var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Display information about nodes in the mesh network",
	Long: `Display information about nodes in the mesh network with customizable fields and sorting.

Available fields:
  id          - Node ID
  user        - User name
  hardware    - Hardware model
  snr         - Signal-to-noise ratio
  distance    - Distance from current node
  last_heard  - Last heard timestamp
  battery     - Battery level
  position    - GPS coordinates
  role        - Node role (CLIENT, ROUTER, REPEATER)

Examples:
  meshtastic nodes
  meshtastic nodes --show-fields id,user,snr,distance
  meshtastic nodes --sort-by snr --show-fields id,user,snr
  meshtastic nodes --live`,
	RunE: runNodes,
}

type NodeInfo struct {
	ID        string    `json:"id"`
	User      string    `json:"user"`
	Hardware  string    `json:"hardware"`
	SNR       float32   `json:"snr"`
	Distance  float32   `json:"distance"`
	LastHeard time.Time `json:"last_heard"`
	Battery   int32     `json:"battery"`
	Position  string    `json:"position"`
	Role      string    `json:"role"`
	IsSelf    bool      `json:"is_self"`
}

func runNodes(cmd *cobra.Command, args []string) error {
	if liveMode {
		return runLiveNodes(cmd, args)
	}

	// Create client and connect
	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	// Get nodes
	nodes, err := gatherNodesInfo(client)
	if err != nil {
		return errors.Wrap(err, "failed to gather nodes information")
	}

	// Sort nodes
	sortNodes(nodes, sortBy)

	// Output in requested format
	if outputJSON {
		data, err := json.MarshalIndent(nodes, "", "  ")
		if err != nil {
			return errors.Wrap(err, "failed to marshal JSON")
		}
		fmt.Println(string(data))
		return nil
	}

	// Default table output
	displayNodesTable(nodes, showFields)
	return nil
}

func runLiveNodes(cmd *cobra.Command, args []string) error {
	// Create client and connect
	client, err := createAndConnectClient()
	if err != nil {
		return err
	}
	defer client.Close()
	defer client.Disconnect()

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Update display every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	fmt.Printf("Live node monitoring (Press Ctrl+C to stop)\n\n")

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-sigChan:
			fmt.Println("\nShutting down...")
			return nil
		case <-ticker.C:
			// Clear screen and update display
			fmt.Print("\033[H\033[2J")
			fmt.Printf("Live node monitoring (Press Ctrl+C to stop)\n")
			fmt.Printf("Last updated: %s\n\n", time.Now().Format("15:04:05"))

			nodes, err := gatherNodesInfo(client)
			if err != nil {
				log.Error().Err(err).Msg("Failed to gather nodes info")
				continue
			}

			sortNodes(nodes, sortBy)
			displayNodesTable(nodes, showFields)
		}
	}
}

func gatherNodesInfo(client *client.RobustMeshtasticClient) ([]NodeInfo, error) {
	myInfo := client.GetMyInfo()
	if myInfo == nil {
		return nil, errors.New("failed to get device info")
	}

	nodes := client.GetNodes()
	nodeInfos := make([]NodeInfo, 0, len(nodes))

	for nodeNum, node := range nodes {
		info := NodeInfo{
			ID:     fmt.Sprintf("!%08x", nodeNum),
			IsSelf: nodeNum == myInfo.MyNodeNum,
		}

		// User information
		user := node.GetUser()
		if user != nil {
			if user.LongName != "" {
				info.User = user.LongName
			} else if user.ShortName != "" {
				info.User = user.ShortName
			} else {
				info.User = "Unknown"
			}
			info.Role = user.Role.String()

			// Hardware information
			if user.HwModel != pb.HardwareModel_UNSET {
				info.Hardware = user.HwModel.String()
			}
		}

		// Device metrics
		deviceMetrics := node.GetDeviceMetrics()
		if deviceMetrics != nil {
			if deviceMetrics.BatteryLevel != nil {
				info.Battery = int32(*deviceMetrics.BatteryLevel)
			}
			if deviceMetrics.ChannelUtilization != nil {
				info.SNR = *deviceMetrics.ChannelUtilization
			}
		}

		// Position information
		position := node.GetPosition()
		if position != nil {
			if position.LatitudeI != nil && position.LongitudeI != nil {
				lat := float64(*position.LatitudeI) / 10000000.0
				lon := float64(*position.LongitudeI) / 10000000.0
				info.Position = fmt.Sprintf("%.6f,%.6f", lat, lon)
			}
		}

		// Last heard
		if node.LastHeard != 0 {
			info.LastHeard = time.Unix(int64(node.LastHeard), 0)
		}

		nodeInfos = append(nodeInfos, info)
	}

	return nodeInfos, nil
}

func sortNodes(nodes []NodeInfo, sortBy string) {
	switch sortBy {
	case "id":
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].ID < nodes[j].ID
		})
	case "user":
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].User < nodes[j].User
		})
	case "snr":
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].SNR > nodes[j].SNR
		})
	case "distance":
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Distance < nodes[j].Distance
		})
	case "last_heard":
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].LastHeard.After(nodes[j].LastHeard)
		})
	default:
		// Default sort by self first, then by ID
		sort.Slice(nodes, func(i, j int) bool {
			if nodes[i].IsSelf && !nodes[j].IsSelf {
				return true
			}
			if !nodes[i].IsSelf && nodes[j].IsSelf {
				return false
			}
			return nodes[i].ID < nodes[j].ID
		})
	}
}

func displayNodesTable(nodes []NodeInfo, fields string) {
	if len(nodes) == 0 {
		fmt.Println("No nodes found")
		return
	}

	// Parse fields
	fieldList := parseFields(fields)

	// Print header
	fmt.Printf("Mesh Network Nodes (%d nodes):\n", len(nodes))
	printTableHeader(fieldList)

	// Print nodes
	for _, node := range nodes {
		printNodeRow(node, fieldList)
	}

	printTableFooter(fieldList)
}

func parseFields(fields string) []string {
	if fields == "" {
		return []string{"id", "user", "hardware", "snr", "distance", "last_heard"}
	}
	return strings.Split(strings.ReplaceAll(fields, " ", ""), ",")
}

func printTableHeader(fields []string) {
	fmt.Printf("┌")
	for i, field := range fields {
		width := getFieldWidth(field)
		fmt.Print(strings.Repeat("─", width))
		if i < len(fields)-1 {
			fmt.Printf("┬")
		}
	}
	fmt.Printf("┐\n")

	fmt.Printf("│")
	for i, field := range fields {
		width := getFieldWidth(field)
		header := getFieldHeader(field)
		fmt.Printf(" %-*s", width-1, header)
		if i < len(fields)-1 {
			fmt.Printf("│")
		}
	}
	fmt.Printf("│\n")

	fmt.Printf("├")
	for i, field := range fields {
		width := getFieldWidth(field)
		fmt.Print(strings.Repeat("─", width))
		if i < len(fields)-1 {
			fmt.Printf("┼")
		}
	}
	fmt.Printf("┤\n")
}

func printNodeRow(node NodeInfo, fields []string) {
	fmt.Printf("│")
	for i, field := range fields {
		width := getFieldWidth(field)
		value := getFieldValue(node, field)
		fmt.Printf(" %-*s", width-1, value)
		if i < len(fields)-1 {
			fmt.Printf("│")
		}
	}
	fmt.Printf("│\n")
}

func printTableFooter(fields []string) {
	fmt.Printf("└")
	for i, field := range fields {
		width := getFieldWidth(field)
		fmt.Print(strings.Repeat("─", width))
		if i < len(fields)-1 {
			fmt.Printf("┴")
		}
	}
	fmt.Printf("┘\n")
}

func getFieldWidth(field string) int {
	switch field {
	case "id":
		return 13
	case "user":
		return 12
	case "hardware":
		return 15
	case "snr":
		return 7
	case "distance":
		return 10
	case "last_heard":
		return 13
	case "battery":
		return 9
	case "position":
		return 20
	case "role":
		return 10
	default:
		return 12
	}
}

func getFieldHeader(field string) string {
	switch field {
	case "id":
		return "Node ID"
	case "user":
		return "User"
	case "hardware":
		return "Hardware"
	case "snr":
		return "SNR"
	case "distance":
		return "Distance"
	case "last_heard":
		return "Last Heard"
	case "battery":
		return "Battery"
	case "position":
		return "Position"
	case "role":
		return "Role"
	default:
		return strings.Title(field)
	}
}

func getFieldValue(node NodeInfo, field string) string {
	switch field {
	case "id":
		return node.ID
	case "user":
		return node.User
	case "hardware":
		return node.Hardware
	case "snr":
		if node.SNR > 0 {
			return fmt.Sprintf("%.1f", node.SNR)
		}
		if node.IsSelf {
			return "N/A"
		}
		return "-"
	case "distance":
		if node.Distance > 0 {
			return fmt.Sprintf("%.1f km", node.Distance)
		}
		if node.IsSelf {
			return "0.0 km"
		}
		return "-"
	case "last_heard":
		if node.IsSelf {
			return "Self"
		}
		if node.LastHeard.IsZero() {
			return "Never"
		}
		return formatTimeSince(node.LastHeard)
	case "battery":
		if node.Battery > 0 {
			return fmt.Sprintf("%d%%", node.Battery)
		}
		return "-"
	case "position":
		if node.Position != "" {
			return node.Position
		}
		return "-"
	case "role":
		if node.Role != "" {
			return node.Role
		}
		return "CLIENT"
	default:
		return "-"
	}
}

func formatTimeSince(t time.Time) string {
	since := time.Since(t)
	if since < time.Minute {
		return "Now"
	}
	if since < time.Hour {
		return fmt.Sprintf("%dm ago", int(since.Minutes()))
	}
	if since < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(since.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(since.Hours()/24))
}

func init() {
	nodesCmd.Flags().StringVar(&showFields, "show-fields", "", "Comma-separated list of fields to display")
	nodesCmd.Flags().StringVar(&sortBy, "sort-by", "", "Sort nodes by field (id, user, snr, distance, last_heard)")
	nodesCmd.Flags().BoolVar(&outputJSON, "json", false, "Output in JSON format")
	nodesCmd.Flags().BoolVar(&liveMode, "live", false, "Live updating display")
}
