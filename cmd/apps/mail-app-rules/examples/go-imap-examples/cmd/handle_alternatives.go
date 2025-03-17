package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
	"github.com/spf13/cobra"
)

var (
	preferHTML bool
	saveOutput bool
	outputFile string
)

// HandleAlternativesCmd demonstrates handling HTML and plain text alternatives
var HandleAlternativesCmd = &cobra.Command{
	Use:   "handle-alternatives",
	Short: "Handle HTML and plain text alternatives",
	Long: `Demonstrates how to handle multipart/alternative messages
that contain both HTML and plain text versions of the content.`,
	Run: func(cmd *cobra.Command, args []string) {
		if uid == 0 {
			log.Fatalf("UID is required for this command. Use --uid flag.")
		}

		client, err := ConnectToIMAP()
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer client.Close()

		// Create a sequence set for the message to fetch
		seqSet := imap.UIDSetNum(imap.UID(uid))

		// Define the fetch options to get the message body
		bodySection := &imap.FetchItemBodySection{} // Empty means fetch the entire message
		fetchOptions := &imap.FetchOptions{
			BodySection: []*imap.FetchItemBodySection{bodySection},
		}

		// Fetch the message
		fetchCmd := client.Fetch(seqSet, fetchOptions)
		defer fetchCmd.Close()

		// Get the first message
		msg := fetchCmd.Next()
		if msg == nil {
			log.Fatalf("Message with UID %d not found", uid)
		}

		// Find the body section in the response
		var bodySectionData imapclient.FetchItemDataBodySection
		var found bool

		for {
			item := msg.Next()
			if item == nil {
				break
			}

			if data, ok := item.(imapclient.FetchItemDataBodySection); ok {
				bodySectionData = data
				found = true
				break
			}
		}

		if !found {
			log.Fatalf("Body section not found in response")
		}

		fmt.Printf("Processing message with UID %d\n", uid)

		// Parse the message using go-message
		mr, err := mail.CreateReader(bodySectionData.Literal)
		if err != nil {
			log.Fatalf("Failed to create mail reader: %v", err)
		}

		// Process the message header
		header := mr.Header
		var subject string
		if s, err := header.Subject(); err == nil {
			subject = s
		} else {
			subject = "Unknown Subject"
		}

		fmt.Printf("Subject: %s\n", subject)

		if date, err := header.Date(); err == nil {
			fmt.Printf("Date: %v\n", date)
		}

		if from, err := header.AddressList("From"); err == nil && len(from) > 0 {
			fmt.Printf("From: %s <%s>\n", from[0].Name, from[0].Address)
		}

		// Variables to store the different content versions
		var plainText, htmlText string

		// Process each part
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatalf("Failed to read part: %v", err)
			}

			if header, ok := part.Header.(*mail.InlineHeader); ok {
				contentType, _, _ := header.ContentType()
				content, err := io.ReadAll(part.Body)
				if err != nil {
					log.Printf("Failed to read content: %v", err)
					continue
				}

				if strings.HasPrefix(contentType, "text/plain") {
					plainText = string(content)
					fmt.Println("\nFound plain text version")
				} else if strings.HasPrefix(contentType, "text/html") {
					htmlText = string(content)
					fmt.Println("\nFound HTML version")
				} else {
					fmt.Printf("\nFound other text part: %s\n", contentType)
				}
			}
		}

		// Determine which version to use
		var selectedContent, contentType string
		if preferHTML && htmlText != "" {
			selectedContent = htmlText
			contentType = "text/html"
			fmt.Println("\nUsing HTML version")
		} else if plainText != "" {
			selectedContent = plainText
			contentType = "text/plain"
			fmt.Println("\nUsing plain text version")
		} else if htmlText != "" {
			// Fallback to HTML if no plain text is available
			selectedContent = htmlText
			contentType = "text/html"
			fmt.Println("\nFalling back to HTML version (no plain text available)")
		} else {
			fmt.Println("\nNo text content found in the message")
			return
		}

		// Display a preview of the selected content
		previewLength := 200
		if len(selectedContent) > previewLength {
			fmt.Printf("\nPreview (%s):\n%s...\n", contentType, selectedContent[:previewLength])
		} else {
			fmt.Printf("\nContent (%s):\n%s\n", contentType, selectedContent)
		}

		// Save the content to a file if requested
		if saveOutput {
			if outputFile == "" {
				// Generate a filename based on the UID and content type
				ext := "txt"
				if contentType == "text/html" {
					ext = "html"
				}
				outputFile = fmt.Sprintf("message_%d.%s", uid, ext)
			}

			// Ensure the directory exists
			dir := filepath.Dir(outputFile)
			if dir != "." {
				if err := os.MkdirAll(dir, 0755); err != nil {
					log.Fatalf("Failed to create directory: %v", err)
				}
			}

			// Write the content to the file
			if err := os.WriteFile(outputFile, []byte(selectedContent), 0644); err != nil {
				log.Fatalf("Failed to write to file: %v", err)
			}

			fmt.Printf("\nSaved %s content to: %s\n", contentType, outputFile)
		}

		// Logout when done
		if err := client.Logout().Wait(); err != nil {
			log.Fatalf("Failed to logout: %v", err)
		}
	},
}

func init() {
	AddCommonFlags(HandleAlternativesCmd)
	HandleAlternativesCmd.Flags().BoolVar(&preferHTML, "prefer-html", false, "Prefer HTML over plain text when both are available")
	HandleAlternativesCmd.Flags().BoolVar(&saveOutput, "save", false, "Save the selected content to a file")
	HandleAlternativesCmd.Flags().StringVar(&outputFile, "output", "", "Output file (default: message_<uid>.<ext>)")
}
