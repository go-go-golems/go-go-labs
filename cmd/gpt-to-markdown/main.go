package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// TODO add flag for only exporting the assistant responses
// TODO add flag for exporting the source blocks
// TODO add flag for adding the messages as comments in the source blocks (if we can detect their type, for example)

func main() {
	var htmlContent []byte
	var err error

	if len(os.Args) < 2 {
		fmt.Println("Usage: gpt-to-markdown <html-file>")
		os.Exit(1)
	}

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

	renameRoles := map[string]string{
		"user":      "john",
		"assistant": "claire",
		"system":    "george",
	}

	renderer := &Renderer{
		RenameRoles:  renameRoles,
		Concise:      true,
		WithMetadata: false,
	}

	// Print the parsed data
	linearConversation := data.Props.PageProps.ServerResponse.LinearConversation
	fmt.Printf("# %s\n\n", data.Props.PageProps.ServerResponse.Title)
	createTime := time.Unix(int64(data.Props.PageProps.ServerResponse.CreateTime), 0)
	fmt.Printf("Created at: %s\n", createTime.Format(time.RFC3339))
	//fmt.Printf("Shared Conversation ID: %s\n", data.Props.PageProps.SharedConversationId)
	fmt.Printf("URL: %s\n\n", os.Args[1])

	renderer.PrintConversation(linearConversation)
}
