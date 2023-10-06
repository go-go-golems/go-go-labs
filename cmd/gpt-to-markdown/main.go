package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type NextData struct {
	Props struct {
		PageProps struct {
			SharedConversationId string `json:"sharedConversationId"`
			ServerResponse       struct {
				ServerResponseData `json:"data"`
			} `json:"serverResponse"`
			Model           map[string]interface{} `json:"model"`
			ModerationState map[string]interface{} `json:"moderation_state"`
		} `json:"pageProps"`
	} `json:"props"`
}

type ServerResponseData struct {
	Title              string         `json:"title"`
	CreateTime         float64        `json:"create_time"`
	UpdateTime         float64        `json:"update_time"`
	LinearConversation []Conversation `json:"linear_conversation"`
}

type Author struct {
	Role     string                 `json:"role"`
	Metadata map[string]interface{} `json:"metadata"`
}

type Content struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

type Message struct {
	ID        string                 `json:"id"`
	Author    Author                 `json:"author"`
	Content   Content                `json:"content"`
	Status    string                 `json:"status"`
	EndTurn   bool                   `json:"end_turn"`
	Weight    float64                `json:"weight"`
	Metadata  map[string]interface{} `json:"metadata"`
	Recipient string                 `json:"recipient"`
}

type Conversation struct {
	ID       string                 `json:"id"`
	Message  Message                `json:"message"`
	Parent   string                 `json:"parent"`
	Children []string               `json:"children"`
	Metadata map[string]interface{} `json:"metadata"`
}

func printConversation(linearConversation []Conversation, withMetadata bool, concise bool) {
	for _, conversation := range linearConversation {
		parts := conversation.Message.Content.Parts
		if len(parts) == 0 {
			continue
		}

		if !concise {
			fmt.Println("### Message Details:")
			fmt.Printf("- **ID**: %s\n", conversation.Message.ID)
			fmt.Printf("- **Author Role**: %s\n", conversation.Message.Author.Role)
		} else {
			fmt.Printf("**%s**: ", conversation.Message.Author.Role)
		}

		if withMetadata && !concise {
			if len(conversation.Message.Author.Metadata) > 0 {
				fmt.Println("- **Author Metadata**:")
				for key, value := range conversation.Message.Author.Metadata {
					fmt.Printf("  - %s: %v\n", key, value)
				}
			}
		}

		if !concise {
			fmt.Printf("- **Content Type**: %s\n", conversation.Message.Content.ContentType)
			fmt.Printf("- **Status**: %s\n", conversation.Message.Status)
			fmt.Printf("- **End Turn**: %v\n", conversation.Message.EndTurn)
			fmt.Printf("- **Weight**: %f\n", conversation.Message.Weight)
			fmt.Printf("- **Recipient**: %s\n", conversation.Message.Recipient)
			if len(conversation.Children) > 0 {
				fmt.Println("- **Children IDs**:")
				for _, child := range conversation.Children {
					fmt.Printf("  - %s\n", child)
				}
			}
		}

		if !concise {
			fmt.Println("- **Parts**: ")
			for _, part := range parts {
				fmt.Printf("  - %s\n", part)
			}
		} else {
			for _, part := range parts {
				fmt.Printf("%s\n", part)
			}
		}

		if withMetadata && !concise {
			if len(conversation.Message.Metadata) > 0 {
				fmt.Println("- **Message Metadata**:")
				for key, value := range conversation.Message.Metadata {
					fmt.Printf("  - %s: %v\n", key, value)
				}
			}
		}

		if !concise {
			fmt.Println("\n---\n")
		} else {
			fmt.Println()
		}
	}
}

func main() {
	var htmlContent []byte
	var err error

	if strings.HasPrefix(os.Args[1], "http://") || strings.HasPrefix(os.Args[1], "https://") {
		// Download the content from the internet
		resp, err := http.Get(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		htmlContent, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Read the file content from local
		htmlContent, err = os.ReadFile(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlContent)))
	if err != nil {
		log.Fatal(err)
	}

	// Extract script content
	scriptContent := doc.Find("#__NEXT_DATA__").Text()

	var data NextData
	err = json.Unmarshal([]byte(scriptContent), &data)
	if err != nil {
		log.Fatal(err)
	}

	// Print the parsed data
	fmt.Printf("Shared Conversation ID: %s\n", data.Props.PageProps.SharedConversationId)
	linearConversation := data.Props.PageProps.ServerResponse.LinearConversation
	printConversation(linearConversation, false, true)
}
