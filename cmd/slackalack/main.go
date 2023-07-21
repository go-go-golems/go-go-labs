package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"time"
)

type File struct {
	UrlPrivate string `json:"url_private"`
}

type Message struct {
	User      string `json:"user"`
	Text      string `json:"text"`
	Timestamp string `json:"ts"`
	Files     []File `json:"files"`
}

type SlackExport struct {
	Messages []Message `json:"messages"`
}

var slackConversationCmd = &cobra.Command{
	Use:   "slack-conversation",
	Short: "Converts slack conversation JSON file to markdown",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		fileBytes, err := os.ReadFile(filename)
		if err != nil {
			cmd.PrintErrf("Unable to read file '%s': %v\n", filename, err)
			return
		}

		export := SlackExport{}
		err = json.Unmarshal(fileBytes, &export)
		if err != nil {
			return
		}

		firstTimePrinted := false
		prevUser := ""
		for _, message := range export.Messages {
			if prevUser != message.User && prevUser != "" {
				fmt.Println("\n---\n")
				fmt.Println("- User:", message.User)
			}

			if !firstTimePrinted {

				ts, _ := strconv.ParseFloat(message.Timestamp, 64)
				timestamp := time.Unix(int64(ts), 0).Format("2006-01-02 03:04:05")
				firstTimePrinted = true
				fmt.Println("- Date:", timestamp)
			}
			fmt.Println(message.Text)
			if len(message.Files) > 0 {
				fmt.Println("\n")

				for _, file := range message.Files {
					fmt.Println(
						fmt.Sprintf("![](%s)", file.UrlPrivate),
					)
				}
			}

			prevUser = message.User
		}
	},
}

var rootCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert JSON data",
}

func main() {
	rootCmd.AddCommand(slackConversationCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
