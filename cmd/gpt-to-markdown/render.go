package main

import (
	"fmt"
	"strings"
	"time"
)

type Renderer struct {
	WithMetadata bool
	Concise      bool
	RenameRoles  map[string]string
}

func (r *Renderer) PrintConversation(url string, response ServerResponseData, linearConversation []Conversation) {
	fmt.Printf("# %s\n\n", response.Title)
	createTime := time.Unix(int64(response.CreateTime), 0)
	fmt.Printf("Created at: %s\n", createTime.Format(time.RFC3339))
	fmt.Printf("URL: %s\n\n", url)
	for _, conversation := range linearConversation {
		parts := conversation.Message.Content.Parts
		if len(parts) == 0 {
			continue
		}

		role := conversation.Message.Author.Role
		if r.RenameRoles != nil {
			if newRole, ok := r.RenameRoles[role]; ok {
				role = newRole
			}
		}

		content := strings.Join(parts, "\n")
		if content == "" {
			continue
		}

		if !r.Concise {
			fmt.Println("### Message Details:")
			fmt.Printf("- **ID**: %s\n", conversation.Message.ID)

			fmt.Printf("- **Author Role**: %s\n", role)
		} else {
			fmt.Printf("**%s**: ", role)
		}

		if r.WithMetadata && !r.Concise {
			if len(conversation.Message.Author.Metadata) > 0 {
				fmt.Println("- **Author Metadata**:")
				for key, value := range conversation.Message.Author.Metadata {
					fmt.Printf("  - %s: %v\n", key, value)
				}
			}
		}

		if !r.Concise {
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

		if !r.Concise {
			fmt.Println("- **Parts**: ")
			for _, part := range parts {
				fmt.Printf("  - %s\n", part)
			}
		} else {
			for _, part := range parts {
				fmt.Printf("%s\n", part)
			}
		}

		if r.WithMetadata && !r.Concise {
			if len(conversation.Message.Metadata) > 0 {
				fmt.Println("- **Message Metadata**:")
				for key, value := range conversation.Message.Metadata {
					fmt.Printf("  - %s: %v\n", key, value)
				}
			}
		}

		fmt.Printf("\n---\n\n")
	}
}
