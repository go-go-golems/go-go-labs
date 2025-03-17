package cmd

import (
	"fmt"
	"log"

	"github.com/emersion/go-imap/v2"
	"github.com/spf13/cobra"
)

var (
	numMessages int
)

// FetchMetadataCmd demonstrates fetching message metadata
var FetchMetadataCmd = &cobra.Command{
	Use:   "fetch-metadata",
	Short: "Fetch message metadata",
	Long: `Demonstrates how to fetch metadata for messages in a mailbox,
including envelope information, flags, and sizes.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := ConnectToIMAP()
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer client.Close()

		// Create a sequence set for the messages to fetch
		var seqSet imap.SeqSet
		if uid > 0 {
			// Fetch a specific message by UID
			seqSet = imap.SeqSetNum(uid)
			fmt.Printf("Fetching metadata for message with UID %d\n", uid)
		} else {
			// Fetch the last N messages
			if numMessages <= 0 {
				numMessages = 10 // Default to 10 messages
			}

			// Get mailbox status to determine message count
			status, err := client.Status(mailbox, &imap.StatusOptions{
				NumMessages: true,
			}).Wait()
			if err != nil {
				log.Fatalf("Failed to get mailbox status: %v", err)
			}

			if status.NumMessages == nil {
				log.Fatalf("Failed to get mailbox status: %v", err)
			}

			// Calculate the sequence range for the last N messages
			start := uint32(1)
			if *status.NumMessages > uint32(numMessages) {
				start = *status.NumMessages - uint32(numMessages) + 1
			}

			seqSet = imap.SeqSetNum(start)
			seqSet.AddNum(*status.NumMessages)
			fmt.Printf("Fetching metadata for the last %d messages\n", numMessages)
		}

		// Define the fetch options
		fetchOptions := &imap.FetchOptions{
			Envelope:     true,
			Flags:        true,
			InternalDate: true,
			RFC822Size:   true,
			UID:          true,
		}

		// Fetch the messages
		messages, err := client.Fetch(seqSet, fetchOptions).Collect()
		if err != nil {
			log.Fatalf("Failed to fetch messages: %v", err)
		}

		// Display the results
		fmt.Printf("Found %d messages\n", len(messages))
		for i, msg := range messages {
			fmt.Printf("\nMessage %d:\n", i+1)
			fmt.Printf("  UID: %d\n", msg.UID)
			fmt.Printf("  Date: %v\n", msg.InternalDate)
			fmt.Printf("  Size: %d bytes\n", msg.RFC822Size)
			fmt.Printf("  Flags: %v\n", msg.Flags)

			if msg.Envelope != nil {
				fmt.Printf("  Subject: %s\n", msg.Envelope.Subject)

				if len(msg.Envelope.From) > 0 {
					from := msg.Envelope.From[0]
					fmt.Printf("  From: %s <%s@%s>\n", from.Name, from.Mailbox, from.Host)
				}

				if len(msg.Envelope.To) > 0 {
					to := msg.Envelope.To[0]
					fmt.Printf("  To: %s <%s@%s>\n", to.Name, to.Mailbox, to.Host)
				}

				fmt.Printf("  Message-ID: %s\n", msg.Envelope.MessageID)
			}
		}

		// Logout when done
		if err := client.Logout().Wait(); err != nil {
			log.Fatalf("Failed to logout: %v", err)
		}
	},
}

func init() {
	AddCommonFlags(FetchMetadataCmd)
	FetchMetadataCmd.Flags().IntVar(&numMessages, "num", 10, "Number of messages to fetch")
}
