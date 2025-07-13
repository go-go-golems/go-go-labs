package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	serverPort int
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Control the LumonStream server",
	Long:  `Start, stop, or check the status of the LumonStream server.`,
}

// statusCmd represents the server status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check server status",
	Long:  `Check if the LumonStream server is running.`,
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("%s/api/stream-info", serverURL)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Server is not running or not accessible at %s\n", serverURL)
			return
		}
		defer resp.Body.Close()

		fmt.Printf("Server is running at %s\n", serverURL)
	},
}

// startCmd represents the server start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long:  `Start the LumonStream server on the specified port.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("To start the server, run the following command in a separate terminal:\n\n")
		fmt.Printf("cd /path/to/LumonStream/backend && ./lumonstream --port %d\n\n", serverPort)
		fmt.Printf("Or if you're using the source code directly:\n\n")
		fmt.Printf("cd /path/to/LumonStream/backend && go run main.go --port %d\n", serverPort)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(statusCmd)
	serverCmd.AddCommand(startCmd)

	// Add flags for server commands
	startCmd.Flags().IntVar(&serverPort, "port", 8080, "Port to run the server on")
}
