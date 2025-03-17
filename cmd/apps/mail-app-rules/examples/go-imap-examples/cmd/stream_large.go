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
	outputDir       string
	saveAttachments bool
	saveTextParts   bool
)

// StreamLargeCmd demonstrates streaming large messages
var StreamLargeCmd = &cobra.Command{
	Use:   "stream-large",
	Short: "Stream large messages",
	Long: `Demonstrates how to stream large messages efficiently,
saving attachments directly to disk without loading them into memory.`,
	Run: func(cmd *cobra.Command, args []string) {
		if uid == 0 {
			log.Fatalf("UID is required for this command. Use --uid flag.")
		}

		client, err := ConnectToIMAP()
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer client.Close()

		// Create output directory if it doesn't exist
		if outputDir == "" {
			outputDir = "attachments"
		}

		if saveAttachments || saveTextParts {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				log.Fatalf("Failed to create output directory: %v", err)
			}
		}

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

		fmt.Printf("Streaming message with UID %d\n", uid)

		// Process items as they arrive
		for {
			item := msg.Next()
			if item == nil {
				break
			}

			if data, ok := item.(imapclient.FetchItemDataBodySection); ok {
				// Create a mail reader that will stream the message
				mr, err := mail.CreateReader(data.Literal)
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

				// Process each part
				partNum := 1
				for {
					part, err := mr.NextPart()
					if err == io.EOF {
						break
					} else if err != nil {
						log.Fatalf("Failed to read part: %v", err)
					}

					switch header := part.Header.(type) {
					case *mail.InlineHeader:
						// This is a text part (plain text or HTML)
						contentType, _, _ := header.ContentType()
						fmt.Printf("\nPart %d: %s\n", partNum, contentType)

						if saveTextParts {
							// Generate a filename based on content type
							ext := "txt"
							if strings.Contains(contentType, "html") {
								ext = "html"
							}

							filename := filepath.Join(outputDir, fmt.Sprintf("%d_part%d.%s", uid, partNum, ext))
							fmt.Printf("Saving text part to: %s\n", filename)

							file, err := os.Create(filename)
							if err != nil {
								log.Printf("Failed to create file: %v", err)
								continue
							}

							n, err := io.Copy(file, part.Body)
							file.Close()
							if err != nil {
								log.Printf("Failed to save text part: %v", err)
							} else {
								fmt.Printf("Saved text part (%d bytes)\n", n)
							}
						} else {
							// Just read a preview
							preview := make([]byte, 100)
							n, _ := part.Body.Read(preview)
							if n > 0 {
								fmt.Printf("Preview: %s\n", preview[:n])
								if n == 100 {
									fmt.Println("... (content truncated)")
								}
							}

							// Discard the rest
							io.Copy(io.Discard, part.Body)
						}

					case *mail.AttachmentHeader:
						// This is an attachment
						filename, _ := header.Filename()
						if filename == "" {
							filename = fmt.Sprintf("attachment_%d_%d", uid, partNum)
						}

						contentType, _, _ := header.ContentType()
						fmt.Printf("\nPart %d: Attachment %s (%s)\n", partNum, filename, contentType)

						if saveAttachments {
							// Save the attachment to disk
							safeName := filepath.Join(outputDir, sanitizeFilename(filename))
							fmt.Printf("Saving attachment to: %s\n", safeName)

							file, err := os.Create(safeName)
							if err != nil {
								log.Printf("Failed to create file: %v", err)
								continue
							}

							n, err := io.Copy(file, part.Body)
							file.Close()
							if err != nil {
								log.Printf("Failed to save attachment: %v", err)
							} else {
								fmt.Printf("Saved attachment %s (%d bytes)\n", filename, n)
							}
						} else {
							// Just count the bytes without storing them
							n, err := io.Copy(io.Discard, part.Body)
							if err != nil {
								log.Printf("Failed to read attachment: %v", err)
							} else {
								fmt.Printf("Attachment size: %d bytes (not saved)\n", n)
							}
						}
					}

					partNum++
				}
			}
		}

		// Logout when done
		if err := client.Logout().Wait(); err != nil {
			log.Fatalf("Failed to logout: %v", err)
		}
	},
}

// sanitizeFilename removes or replaces characters that are invalid in filenames
func sanitizeFilename(filename string) string {
	// Replace invalid characters with underscores
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename

	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}

	return result
}

func init() {
	AddCommonFlags(StreamLargeCmd)
	StreamLargeCmd.Flags().StringVar(&outputDir, "output", "attachments", "Directory to save attachments")
	StreamLargeCmd.Flags().BoolVar(&saveAttachments, "save-attachments", true, "Save attachments to disk")
	StreamLargeCmd.Flags().BoolVar(&saveTextParts, "save-text", false, "Save text parts to disk")
}
