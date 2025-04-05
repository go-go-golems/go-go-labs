package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

// StreamInfo represents the stream information data structure
type StreamInfo struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"startTime"`
	Language    string    `json:"language"`
	GithubRepo  string    `json:"githubRepo"`
	ViewerCount int       `json:"viewerCount"`
}

// Response is a generic API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get stream information",
	Long:  `Retrieve the current stream information from the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("%s/api/stream-info", serverURL)
		
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error connecting to server: %v\n", err)
			return
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			return
		}
		
		var response Response
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return
		}
		
		if !response.Success {
			fmt.Printf("Error: %s\n", response.Message)
			return
		}
		
		// Pretty print the data
		prettyJSON, err := json.MarshalIndent(response.Data, "", "  ")
		if err != nil {
			fmt.Printf("Error formatting response: %v\n", err)
			return
		}
		
		fmt.Println(string(prettyJSON))
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
