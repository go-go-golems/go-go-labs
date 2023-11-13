package cmd

import (
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/pkg"
	"github.com/spf13/cobra"
	"os"
)

var AssistantCmd = &cobra.Command{
	Use:   "assistant",
	Short: "Manage OpenAI Assistants",
}

var createCmd = &cobra.Command{
	Use:   "create [name] [model] [instructions]",
	Short: "Create a new assistant",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		assistantData := pkg.Assistant{
			Name:         args[0],
			Model:        args[1],
			Instructions: args[2],
		}
		assistant, err := pkg.CreateAssistant(apiKey, assistantData)
		if err != nil {
			fmt.Println("Error creating assistant:", err)
			return
		}
		fmt.Printf("Assistant created: %+v\n", assistant)
	},
}

var retrieveCmd = &cobra.Command{
	Use:   "retrieve [assistantID]",
	Short: "Retrieve an assistant",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := os.Getenv("OPENAI_API_KEY") // get API key
		assistant, err := pkg.RetrieveAssistant(apiKey, args[0])
		if err != nil {
			fmt.Println("Error retrieving assistant:", err)
			return
		}
		fmt.Printf("Assistant retrieved: %+v\n", assistant)
	},
}

var modifyCmd = &cobra.Command{
	Use:   "modify [assistantID] [name] [model] [instructions]",
	Short: "Modify an assistant",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		updateData := pkg.Assistant{
			Name:         args[1],
			Model:        args[2],
			Instructions: args[3],
		}
		assistant, err := pkg.ModifyAssistant(apiKey, args[0], updateData)
		if err != nil {
			fmt.Println("Error modifying assistant:", err)
			return
		}
		fmt.Printf("Assistant modified: %+v\n", assistant)
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [assistantID]",
	Short: "Delete an assistant",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		err := pkg.DeleteAssistant(apiKey, args[0])
		if err != nil {
			fmt.Println("Error deleting assistant:", err)
			return
		}
		fmt.Println("Assistant deleted successfully")
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all assistants",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		after := "" // Initialize after to an empty string
		limit := 20 // Set default limit or get from flags

		for {
			assistants, hasMore, err := pkg.ListAssistants(apiKey, after, limit)
			if err != nil {
				fmt.Println("Error listing assistants:", err)
				return
			}

			for _, assistant := range assistants {
				fmt.Printf("Assistant: %+v\n", assistant)
			}

			if !hasMore {
				break
			}

			fmt.Print("Load more? (y/n): ")
			var input string
			_, _ = fmt.Scanln(&input)
			if input != "y" {
				break
			}

			if len(assistants) > 0 {
				after = assistants[len(assistants)-1].ID
			}
		}
	},
}

func init() {
	AssistantCmd.AddCommand(createCmd)
	AssistantCmd.AddCommand(retrieveCmd)
	AssistantCmd.AddCommand(modifyCmd)
	AssistantCmd.AddCommand(deleteCmd)
	AssistantCmd.AddCommand(listCmd)
}
