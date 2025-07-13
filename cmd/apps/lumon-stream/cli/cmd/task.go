package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	content string
	status  string
	stepID  int
)

// taskCmd represents the task command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage stream tasks",
	Long:  `Add, update, or manage tasks for the stream.`,
}

// addTaskCmd represents the add task command
var addTaskCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new task",
	Long:  `Add a new task to the stream.`,
	Run: func(cmd *cobra.Command, args []string) {
		if content == "" {
			fmt.Println("Error: Task content is required")
			return
		}

		// Validate status
		if status != "completed" && status != "active" && status != "upcoming" {
			status = "upcoming" // Default to upcoming if invalid
		}

		url := fmt.Sprintf("%s/api/steps", serverURL)

		// Prepare request data
		requestData := map[string]string{
			"content": content,
			"status":  status,
		}

		jsonData, err := json.Marshal(requestData)
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
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
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

		fmt.Println("Task added successfully")
	},
}

// updateTaskCmd represents the update task status command
var updateTaskCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a task's status",
	Long:  `Update the status of an existing task.`,
	Run: func(cmd *cobra.Command, args []string) {
		if stepID <= 0 {
			fmt.Println("Error: Valid step ID is required")
			return
		}

		// Validate status
		if status != "completed" && status != "active" && status != "upcoming" {
			fmt.Println("Error: Valid status is required (completed, active, or upcoming)")
			return
		}

		url := fmt.Sprintf("%s/api/steps/status", serverURL)

		// Prepare request data
		requestData := map[string]interface{}{
			"id":     stepID,
			"status": status,
		}

		jsonData, err := json.Marshal(requestData)
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
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
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

		fmt.Println("Task status updated successfully")
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(addTaskCmd)
	taskCmd.AddCommand(updateTaskCmd)

	// Add flags for task commands
	addTaskCmd.Flags().StringVar(&content, "content", "", "Task content")
	addTaskCmd.Flags().StringVar(&status, "status", "upcoming", "Task status (completed, active, or upcoming)")
	addTaskCmd.MarkFlagRequired("content")

	updateTaskCmd.Flags().IntVar(&stepID, "id", 0, "Task ID")
	updateTaskCmd.Flags().StringVar(&status, "status", "", "Task status (completed, active, or upcoming)")
	updateTaskCmd.MarkFlagRequired("id")
	updateTaskCmd.MarkFlagRequired("status")
}
