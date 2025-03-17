package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/emersion/go-imap/v2"
	"github.com/spf13/cobra"
)

var (
	partPath    string
	fetchHeader bool
	fetchMime   bool
	fetchText   bool
)

// FetchPartsCmd demonstrates fetching specific message parts
var FetchPartsCmd = &cobra.Command{
	Use:   "fetch-parts",
	Short: "Fetch specific message parts",
	Long: `Demonstrates how to fetch specific parts of a message
based on the message structure.`,
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

		// First, fetch the structure to understand the parts
		structureOptions := &imap.FetchOptions{
			BodyStructure: &imap.FetchItemBodyStructure{Extended: true},
		}

		structMsgs, err := client.Fetch(seqSet, structureOptions).Collect()
		if err != nil {
			log.Fatalf("Failed to fetch structure: %v", err)
		}

		if len(structMsgs) == 0 {
			log.Fatalf("Message with UID %d not found", uid)
		}

		// Display the structure
		msg := structMsgs[0]
		fmt.Printf("Structure for message with UID %d:\n\n", uid)

		if msg.BodyStructure == nil {
			log.Fatalf("No body structure returned")
		}

		// Print the structure
		msg.BodyStructure.Walk(func(path []int, part imap.BodyStructure) bool {
			indent := strings.Repeat("  ", len(path))
			pathStr := formatPath(path)

			fmt.Printf("%s%s: %s\n", indent, pathStr, part.MediaType())
			return true
		})

		// Parse the part path
		var partIndices []int
		if partPath != "" {
			parts := strings.Split(partPath, ".")
			partIndices = make([]int, len(parts))

			for i, p := range parts {
				idx, err := strconv.Atoi(p)
				if err != nil {
					log.Fatalf("Invalid part path: %s", partPath)
				}
				// Convert from 1-based (user input) to 0-based (internal)
				partIndices[i] = idx - 1
			}
		}

		// Create the body section items based on flags
		var bodySections []*imap.FetchItemBodySection

		// If no specifier is set, fetch the whole part
		if !fetchHeader && !fetchMime && !fetchText {
			bodySections = append(bodySections, &imap.FetchItemBodySection{
				Part: partIndices,
			})
		} else {
			if fetchHeader {
				bodySections = append(bodySections, &imap.FetchItemBodySection{
					Part:      partIndices,
					Specifier: imap.PartSpecifierHeader,
				})
			}

			if fetchMime {
				bodySections = append(bodySections, &imap.FetchItemBodySection{
					Part:      partIndices,
					Specifier: imap.PartSpecifierMIME,
				})
			}

			if fetchText {
				bodySections = append(bodySections, &imap.FetchItemBodySection{
					Part:      partIndices,
					Specifier: imap.PartSpecifierText,
				})
			}
		}

		// Fetch the specific parts
		fetchOptions := &imap.FetchOptions{
			BodySection: bodySections,
		}

		messages, err := client.Fetch(seqSet, fetchOptions).Collect()
		if err != nil {
			log.Fatalf("Failed to fetch parts: %v", err)
		}

		if len(messages) == 0 {
			log.Fatalf("Message with UID %d not found", uid)
		}

		// Display the fetched parts
		fetchedMsg := messages[0]
		fmt.Println("\nFetched Parts:")
		fmt.Println("--------------")

		for _, section := range bodySections {
			data := fetchedMsg.FindBodySection(section)
			if data == nil {
				fmt.Printf("Part not found: %v\n", formatPartSection(section))
				continue
			}

			fmt.Printf("\nPart: %s\n", formatPartSection(section))
			fmt.Printf("Size: %d bytes\n", len(data))

			// Print the content (limit to 1000 bytes for display)
			maxDisplay := 1000
			if len(data) > maxDisplay {
				fmt.Printf("Content (first %d bytes):\n%s\n... (truncated)\n", maxDisplay, data[:maxDisplay])
			} else {
				fmt.Printf("Content:\n%s\n", data)
			}
		}

		// Logout when done
		if err := client.Logout().Wait(); err != nil {
			log.Fatalf("Failed to logout: %v", err)
		}
	},
}

// formatPartSection formats a FetchItemBodySection for display
func formatPartSection(section *imap.FetchItemBodySection) string {
	var parts []string

	// Format the part path
	if len(section.Part) > 0 {
		pathParts := make([]string, len(section.Part))
		for i, p := range section.Part {
			pathParts[i] = fmt.Sprintf("%d", p+1) // Convert to 1-based for display
		}
		parts = append(parts, strings.Join(pathParts, "."))
	} else {
		parts = append(parts, "BODY")
	}

	// Add the specifier if any
	if section.Specifier != imap.PartSpecifierNone {
		parts = append(parts, string(section.Specifier))
	}

	return strings.Join(parts, ".")
}

func init() {
	AddCommonFlags(FetchPartsCmd)
	FetchPartsCmd.Flags().StringVar(&partPath, "part", "", "Part path to fetch (e.g., '1.2.3')")
	FetchPartsCmd.Flags().BoolVar(&fetchHeader, "header", false, "Fetch the header of the part")
	FetchPartsCmd.Flags().BoolVar(&fetchMime, "mime", false, "Fetch the MIME headers of the part")
	FetchPartsCmd.Flags().BoolVar(&fetchText, "text", false, "Fetch the text/body of the part")
}
