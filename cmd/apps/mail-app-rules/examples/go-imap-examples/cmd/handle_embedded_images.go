package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
	"github.com/spf13/cobra"
)

var (
	imagesDir   string
	processHTML bool
)

// HandleEmbeddedImagesCmd demonstrates handling embedded images in HTML emails
var HandleEmbeddedImagesCmd = &cobra.Command{
	Use:   "handle-embedded-images",
	Short: "Handle embedded images in HTML emails",
	Long: `Demonstrates how to handle embedded images (inline attachments) in HTML emails,
extracting them and optionally modifying the HTML to reference local files.`,
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

		// Create images directory if it doesn't exist
		if imagesDir == "" {
			imagesDir = fmt.Sprintf("images_%d", uid)
		}

		if err := os.MkdirAll(imagesDir, 0755); err != nil {
			log.Fatalf("Failed to create images directory: %v", err)
		}

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

		// Map to store Content-ID -> filename for embedded images
		embeddedImages := make(map[string]string)
		var htmlContent string

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
				// Check if this is the HTML part
				contentType, _, _ := header.ContentType()
				if strings.HasPrefix(contentType, "text/html") {
					content, err := io.ReadAll(part.Body)
					if err != nil {
						log.Printf("Failed to read HTML content: %v", err)
						continue
					}

					htmlContent = string(content)
					fmt.Println("\nFound HTML content")
				} else {
					// Skip other inline parts
					io.Copy(io.Discard, part.Body)
				}

			case *mail.AttachmentHeader:
				// Check if this is an inline attachment with Content-ID
				contentID := header.Get("Content-ID")
				disposition, _, _ := header.ContentDisposition()

				if contentID != "" || disposition == "inline" {
					// Clean up the Content-ID (remove < and >)
					contentID = strings.Trim(contentID, "<>")

					// Get the filename or generate one
					filename, err := header.Filename()
					if err != nil || filename == "" {
						if contentID != "" {
							// Try to extract a filename from the Content-ID
							filename = sanitizeFilename(contentID)
						} else {
							// Generate a generic filename
							contentType, _, _ := header.ContentType()
							ext := getExtensionFromContentType(contentType)
							filename = fmt.Sprintf("inline_%d_%d%s", uid, partNum, ext)
						}
					}

					// Ensure the filename is safe and unique
					safeName := sanitizeFilename(filename)
					fullPath := filepath.Join(imagesDir, safeName)

					// Save the image
					file, err := os.Create(fullPath)
					if err != nil {
						log.Printf("Failed to create file: %v", err)
						continue
					}

					n, err := io.Copy(file, part.Body)
					file.Close()
					if err != nil {
						log.Printf("Failed to save image: %v", err)
						continue
					}

					// Store the mapping
					if contentID != "" {
						embeddedImages[contentID] = safeName
						fmt.Printf("\nSaved embedded image with Content-ID %s to %s (%d bytes)\n",
							contentID, safeName, n)
					} else {
						fmt.Printf("\nSaved inline attachment to %s (%d bytes)\n",
							safeName, n)
					}
				} else {
					// Skip regular attachments
					io.Copy(io.Discard, part.Body)
				}
			}

			partNum++
		}

		// Process the HTML content if requested
		if processHTML && htmlContent != "" && len(embeddedImages) > 0 {
			fmt.Println("\nProcessing HTML content to replace embedded image references")

			// Replace cid: references with file paths
			for cid, filename := range embeddedImages {
				// Look for cid: references in the HTML
				cidPattern := fmt.Sprintf("cid:%s", regexp.QuoteMeta(cid))
				localPath := filepath.Join(imagesDir, filename)

				// Use file:// URLs for local files
				replacement := fmt.Sprintf("file://%s", localPath)

				// Replace in the HTML
				htmlContent = strings.ReplaceAll(htmlContent, cidPattern, replacement)

				fmt.Printf("Replaced references to cid:%s with %s\n", cid, replacement)
			}

			// Save the processed HTML
			htmlFilename := fmt.Sprintf("message_%d.html", uid)
			if err := os.WriteFile(htmlFilename, []byte(htmlContent), 0644); err != nil {
				log.Fatalf("Failed to write HTML file: %v", err)
			}

			fmt.Printf("\nSaved processed HTML to: %s\n", htmlFilename)
		}

		// Logout when done
		if err := client.Logout().Wait(); err != nil {
			log.Fatalf("Failed to logout: %v", err)
		}
	},
}

// getExtensionFromContentType returns a file extension based on the content type
func getExtensionFromContentType(contentType string) string {
	switch {
	case strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg"):
		return ".jpg"
	case strings.Contains(contentType, "png"):
		return ".png"
	case strings.Contains(contentType, "gif"):
		return ".gif"
	case strings.Contains(contentType, "bmp"):
		return ".bmp"
	case strings.Contains(contentType, "tiff"):
		return ".tiff"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	case strings.Contains(contentType, "svg"):
		return ".svg"
	default:
		return ".bin"
	}
}

func init() {
	AddCommonFlags(HandleEmbeddedImagesCmd)
	HandleEmbeddedImagesCmd.Flags().StringVar(&imagesDir, "images-dir", "", "Directory to save embedded images (default: images_<uid>)")
	HandleEmbeddedImagesCmd.Flags().BoolVar(&processHTML, "process-html", true, "Process HTML to replace embedded image references")
}
