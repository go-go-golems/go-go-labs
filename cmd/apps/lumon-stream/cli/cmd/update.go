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

var (
	title       string
	description string
	language    string
	githubRepo  string
	viewerCount int
	resetTimer  bool
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update stream information",
	Long:  `Update the stream information on the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		// First get current info to preserve fields not being updated
		url := fmt.Sprintf("%s/api/stream-info", serverURL)
		
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error connecting to server: %v\n", err)
			return
		}
		
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
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
		
		// Extract stream info from response
		dataJSON, err := json.Marshal(response.Data)
		if err != nil {
			fmt.Printf("Error processing response: %v\n", err)
			return
		}
		
		var streamInfoResponse struct {
			StreamInfo StreamInfo `json:"StreamInfo"`
		}
		
		if err := json.Unmarshal(dataJSON, &streamInfoResponse); err != nil {
			fmt.Printf("Error extracting stream info: %v\n", err)
			return
		}
		
		info := streamInfoResponse.StreamInfo
		
		// Update fields that were specified
		if cmd.Flags().Changed("title") {
			info.Title = title
		}
		if cmd.Flags().Changed("description") {
			info.Description = description
		}
		if cmd.Flags().Changed("language") {
			info.Language = language
		}
		if cmd.Flags().Changed("github-repo") {
			info.GithubRepo = githubRepo
		}
		if cmd.Flags().Changed("viewer-count") {
			info.ViewerCount = viewerCount
		}
		if resetTimer {
			info.StartTime = time.Now()
		}
		
		// Send update request
		jsonData, err := json.Marshal(info)
		if err != nil {
			fmt.Printf("Error preparing request: %v\n", err)
			return
		}
		
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			return
		}
		defer resp.Body.Close()
		
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			return
		}
		
		var updateResponse Response
		if err := json.Unmarshal(body, &updateResponse); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return
		}
		
		if !updateResponse.Success {
			fmt.Printf("Error: %s\n", updateResponse.Message)
			return
		}
		
		fmt.Println("Stream information updated successfully")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	
	// Add flags for all updatable fields
	updateCmd.Flags().StringVar(&title, "title", "", "Stream title")
	updateCmd.Flags().StringVar(&description, "description", "", "Stream description")
	updateCmd.Flags().StringVar(&language, "language", "", "Programming language/framework")
	updateCmd.Flags().StringVar(&githubRepo, "github-repo", "", "GitHub repository URL")
	updateCmd.Flags().IntVar(&viewerCount, "viewer-count", 0, "Current viewer count")
	updateCmd.Flags().BoolVar(&resetTimer, "reset-timer", false, "Reset the stream timer")
}
