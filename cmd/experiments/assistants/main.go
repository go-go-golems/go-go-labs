package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/cmd/assistants"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/cmd/files"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/cmd/threads"
	assistants2 "github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/pkg/assistants"
	"github.com/spf13/cobra"
	"os"
	"time"
)

const (
	openAIURL          = "https://api.openai.com/v1/"
	assistantsEndpoint = openAIURL + "assistants"
	threadsEndpoint    = openAIURL + "threads"
	messagesEndpoint   = openAIURL + "threads/%s/messages"
	runsEndpoint       = openAIURL + "threads/%s/runs"
)

func runAssistant(apiKey, threadID string, run assistants2.Run) (string, error) {
	url := fmt.Sprintf(runsEndpoint, threadID)
	runParameters := map[string]interface{}{
		"assistant_id": run.AssistantID,
	}
	response, err := doPostRequest(apiKey, url, runParameters)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", err
	}

	// Extract the run ID from the response
	runID, ok := result["id"].(string)
	if !ok {
		// check for result["error"] which is a map[string]interface{}
		if err, ok := result["error"].(map[string]interface{}); ok {
			if message, ok := err["message"].(string); ok {
				return "", fmt.Errorf("could not extract run ID from response: %s", message)
			}
			return "", fmt.Errorf("could not extract run ID from response: %v", err)
		}
		return "", fmt.Errorf("could not extract run ID from response")
	}

	return runID, nil
}

func createAssistant(apiKey string, assistant assistants2.Assistant) (string, error) {
	response, err := doPostRequest(apiKey, assistantsEndpoint, assistant)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", err
	}

	// Extract the assistant ID from the response
	assistantID, ok := result["id"].(string)
	if !ok {
		// check for result["error"] which is a map[string]interface{}
		if err, ok := result["error"].(map[string]interface{}); ok {
			if message, ok := err["message"].(string); ok {
				return "", fmt.Errorf("could not extract run ID from response: %s", message)
			}
			return "", fmt.Errorf("could not extract run ID from response: %v", err)
		}
		return "", fmt.Errorf("could not extract run ID from response")
	}

	return assistantID, nil
}

func createThread(apiKey string) (string, error) {
	response, err := doPostRequest(apiKey, threadsEndpoint, nil)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", err
	}

	// Extract the thread ID from the response
	threadID, ok := result["id"].(string)
	if !ok {
		return "", fmt.Errorf("could not extract thread ID from response")
	}

	return threadID, nil
}

func addMessageToThread(apiKey, threadID string, message assistants2.Message) (string, error) {
	url := fmt.Sprintf(messagesEndpoint, threadID)
	response, err := doPostRequest(apiKey, url, message)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", err
	}

	// Extract the message ID from the response
	messageID, ok := result["id"].(string)
	if !ok {
		// check for result["error"] which is a map[string]interface{}
		if err, ok := result["error"].(map[string]interface{}); ok {
			if message, ok := err["message"].(string); ok {
				return "", fmt.Errorf("could not extract run ID from response: %s", message)
			}
			return "", fmt.Errorf("could not extract run ID from response: %v", err)
		}
		return "", fmt.Errorf("could not extract run ID from response")
	}
	return messageID, err
}

// Helper function to make an HTTP POST request
func pollRunCompletion(apiKey, threadID, runID string) error {
	url := fmt.Sprintf(runsEndpoint+"/%s", threadID, runID)

	for {
		response, err := doGetRequest(apiKey, url)
		if err != nil {
			return err
		}

		var run map[string]interface{}
		if err := json.Unmarshal(response, &run); err != nil {
			return err
		}

		status, ok := run["status"].(string)
		if !ok {
			return fmt.Errorf("could not extract run status from response")
		}

		if status == "completed" {
			break
		}

		// Sleep for a short duration before polling again
		time.Sleep(2 * time.Second)
	}

	return nil
}

func printThreadMessages(apiKey, threadID string) error {
	url := fmt.Sprintf(messagesEndpoint, threadID)
	response, err := doGetRequest(apiKey, url)
	if err != nil {
		return err
	}

	var messages struct {
		Data []struct {
			Role    string `json:"role"`
			Content []struct {
				Text struct {
					Value string `json:"value"`
				} `json:"text"`
			} `json:"content"`
		} `json:"data"`
	}

	if err := json.Unmarshal(response, &messages); err != nil {
		return err
	}

	for _, message := range messages.Data {
		for _, content := range message.Content {
			fmt.Printf("[%s] %s\n", message.Role, content.Text.Value)
		}
	}

	return nil
}

func main() {
	var rootCmd = &cobra.Command{Use: "assistant-cli"}
	rootCmd.AddCommand(assistants.AssistantCmd)
	rootCmd.AddCommand(files.FilesCmd)
	rootCmd.AddCommand(threads.ThreadCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func oldMain() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Set your OPENAI_API_KEY as an environment variable.")
		os.Exit(1)
	}

	// Step 1: Create an Assistant
	assistant := assistants2.Assistant{
		Name:         "Math Tutor",
		Instructions: "You are a personal math tutor. Answer questions briefly, in a sentence or less.",
		Model:        "gpt-4-1106-preview", // Replace with the model you want to use
	}
	assistantID, err := createAssistant(apiKey, assistant)
	if err != nil {
		fmt.Printf("Error creating assistant: %s\n", err)
		os.Exit(1)
	}

	// Step 2: Create a Thread
	threadID, err := createThread(apiKey)
	if err != nil {
		fmt.Printf("Error creating thread: %s\n", err)
		os.Exit(1)
	}

	// Step 3: Add a Message to the Thread
	message := assistants2.Message{
		ThreadID: threadID,
		Role:     "user",
		Content: []assistants2.ContentObject{
			{
				Type: "text",
				Text: &assistants2.TextContent{
					Value: "I need to solve the equation `3x + 11 = 14`. Can you help me?",
				},
			},
		},
	}
	messageID, err := addMessageToThread(apiKey, threadID, message)
	if err != nil {
		fmt.Printf("Error adding message to thread: %s\n", err)
		os.Exit(1)
	}
	_ = messageID

	// Step 4: Run the Assistant

	run := assistants2.Run{
		ThreadID:    threadID,
		AssistantID: assistantID,
	}
	runID, err := runAssistant(apiKey, threadID, run)
	if err != nil {
		fmt.Printf("Error running assistant: %s\n", err)
		os.Exit(1)
	}

	// Poll for Run completion
	if err := pollRunCompletion(apiKey, threadID, runID); err != nil {
		fmt.Printf("Error polling run completion: %s\n", err)
		os.Exit(1)
	}

	// Print resulting messages
	if err := printThreadMessages(apiKey, threadID); err != nil {
		fmt.Printf("Error printing thread messages: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Assistant run successfully.")
}
