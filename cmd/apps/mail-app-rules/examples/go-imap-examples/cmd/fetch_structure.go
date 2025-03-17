package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/emersion/go-imap/v2"
	"github.com/spf13/cobra"
)

// FetchStructureCmd demonstrates fetching message structure
var FetchStructureCmd = &cobra.Command{
	Use:   "fetch-structure",
	Short: "Fetch message structure",
	Long: `Demonstrates how to fetch the MIME structure of a message
to understand its parts before fetching the content.`,
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

		// Define the fetch options to get the extended body structure
		fetchOptions := &imap.FetchOptions{
			BodyStructure: &imap.FetchItemBodyStructure{
				Extended: true, // Get detailed structure
			},
		}

		// Fetch the message
		messages, err := client.Fetch(seqSet, fetchOptions).Collect()
		if err != nil {
			log.Fatalf("Failed to fetch message structure: %v", err)
		}

		if len(messages) == 0 {
			log.Fatalf("Message with UID %d not found", uid)
		}

		msg := messages[0]
		fmt.Printf("Structure for message with UID %d:\n\n", uid)

		// Process the structure
		if msg.BodyStructure == nil {
			fmt.Println("No body structure returned")
		} else {
			// Walk through the structure
			msg.BodyStructure.Walk(func(path []int, part imap.BodyStructure) bool {
				indent := strings.Repeat("  ", len(path))
				pathStr := formatPath(path)

				fmt.Printf("%s%s: %s\n", indent, pathStr, part.MediaType())

				// For single parts, we can get more details
				if singlePart, ok := part.(*imap.BodyStructureSinglePart); ok {
					fmt.Printf("%s  Type: %s/%s\n", indent, singlePart.Type, singlePart.Subtype)
					fmt.Printf("%s  Encoding: %s\n", indent, singlePart.Encoding)
					fmt.Printf("%s  Size: %d bytes\n", indent, singlePart.Size)

					// Check for filename in Content-Disposition
					if singlePart.Extended != nil && singlePart.Extended.Disposition != nil {
						fmt.Printf("%s  Disposition: %s\n", indent, singlePart.Extended.Disposition.Value)

						if filename, ok := singlePart.Extended.Disposition.Params["filename"]; ok {
							fmt.Printf("%s  Filename: %s\n", indent, filename)
						}
					}
				}

				// For multipart, show the subtype
				if multiPart, ok := part.(*imap.BodyStructureMultiPart); ok {
					fmt.Printf("%s  Multipart subtype: %s\n", indent, multiPart.Subtype)
				}

				return true // Continue walking
			})
		}

		// Logout when done
		if err := client.Logout().Wait(); err != nil {
			log.Fatalf("Failed to logout: %v", err)
		}
	},
}

// formatPath formats a path array as a string like "1.2.3"
func formatPath(path []int) string {
	if len(path) == 0 {
		return "ROOT"
	}

	// Convert path to 1-based for display (IMAP uses 1-based indexing)
	parts := make([]string, len(path))
	for i, p := range path {
		parts[i] = fmt.Sprintf("%d", p+1)
	}

	return "PART " + strings.Join(parts, ".")
}

func init() {
	AddCommonFlags(FetchStructureCmd)
}
