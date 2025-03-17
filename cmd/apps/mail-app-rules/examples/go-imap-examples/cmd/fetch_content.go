package cmd

import (
	"fmt"
	"io"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// FetchContentCmd demonstrates fetching and parsing message content
var FetchContentCmd = &cobra.Command{
	Use:   "fetch-content",
	Short: "Fetch message content",
	Long: `Demonstrates how to fetch the full content of a message
and parse it using the go-message library.`,
	Run: func(cmd *cobra.Command, args []string) {
		if uid == 0 {
			log.Fatal().Msg("UID is required for this command. Use --uid flag.")
		}

		log.Debug().Uint32("uid", uid).Msg("Starting fetch-content command")

		client, err := ConnectToIMAP()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to connect to IMAP server")
		}
		defer client.Close()

		// Create a sequence set for the message to fetch
		seqSet := imap.UIDSetNum(imap.UID(uid))
		log.Debug().Str("seqSet", seqSet.String()).Msg("Created sequence set")

		// Define the fetch options to get the message body
		bodySection := &imap.FetchItemBodySection{} // Empty means fetch the entire message
		fetchOptions := &imap.FetchOptions{
			BodySection: []*imap.FetchItemBodySection{bodySection},
		}
		log.Debug().Msg("Fetching message body")

		// Fetch the message
		fetchCmd := client.Fetch(seqSet, fetchOptions)
		defer fetchCmd.Close()

		// Get the first message
		log.Debug().Msg("Waiting for message data")
		msg := fetchCmd.Next()
		if msg == nil {
			log.Fatal().Uint32("uid", uid).Msg("Message not found")
		}
		log.Debug().Msg("Received message data")

		// Find the body section in the response
		var bodySectionData imapclient.FetchItemDataBodySection
		var found bool

		log.Debug().Msg("Processing message items")
		for {
			item := msg.Next()
			if item == nil {
				break
			}

			if data, ok := item.(imapclient.FetchItemDataBodySection); ok {
				bodySectionData = data
				found = true
				log.Debug().Msg("Found body section data")
				break
			}
		}

		if !found {
			log.Fatal().Msg("Body section not found in response")
		}

		log.Info().Uint32("uid", uid).Msg("Fetched message")

		// Parse the message using go-message
		log.Debug().Msg("Creating mail reader")
		mr, err := mail.CreateReader(bodySectionData.Literal)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create mail reader")
		}

		// Process the message header
		header := mr.Header
		log.Debug().Msg("Processing message headers")

		if date, err := header.Date(); err == nil {
			log.Debug().Time("date", date).Msg("Message date")
			fmt.Printf("Date: %v\n", date)
		}

		if from, err := header.AddressList("From"); err == nil {
			log.Debug().Interface("from", from).Msg("From addresses")
			fmt.Printf("From: ")
			for i, addr := range from {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s <%s>", addr.Name, addr.Address)
			}
			fmt.Println()
		}

		if to, err := header.AddressList("To"); err == nil {
			log.Debug().Interface("to", to).Msg("To addresses")
			fmt.Printf("To: ")
			for i, addr := range to {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s <%s>", addr.Name, addr.Address)
			}
			fmt.Println()
		}

		if subject, err := header.Subject(); err == nil {
			log.Debug().Str("subject", subject).Msg("Message subject")
			fmt.Printf("Subject: %v\n", subject)
		}

		fmt.Println("\nMessage Parts:")
		fmt.Println("-------------")

		// Process each part of the message
		partNum := 1
		log.Debug().Msg("Starting to process message parts")
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				log.Debug().Msg("Reached end of message parts")
				break
			} else if err != nil {
				log.Fatal().Err(err).Msg("Failed to read message part")
			}

			log.Debug().Int("partNum", partNum).Msg("Processing message part")
			fmt.Printf("\nPart %d:\n", partNum)

			switch header := part.Header.(type) {
			case *mail.InlineHeader:
				contentType, params, _ := header.ContentType()
				log.Debug().
					Str("type", contentType).
					Interface("params", params).
					Msg("Processing inline part")

				fmt.Printf("  Type: %s\n", contentType)
				if charset, ok := params["charset"]; ok {
					fmt.Printf("  Charset: %s\n", charset)
				}

				// Read up to 500 characters of the content
				content := make([]byte, 500)
				n, _ := part.Body.Read(content)
				if n > 0 {
					log.Debug().Int("bytes_read", n).Msg("Read content preview")
					fmt.Printf("  Content preview: %s\n", content[:n])
					if n == 500 {
						fmt.Println("  ... (content truncated)")
					}
				}

			case *mail.AttachmentHeader:
				filename, _ := header.Filename()
				contentType, _, _ := header.ContentType()
				log.Debug().
					Str("filename", filename).
					Str("type", contentType).
					Msg("Processing attachment")

				fmt.Printf("  Attachment: %s\n", filename)
				fmt.Printf("  Type: %s\n", contentType)

				// Get the size of the attachment
				data, err := io.ReadAll(part.Body)
				if err != nil {
					log.Error().Err(err).Str("filename", filename).Msg("Error reading attachment")
					fmt.Printf("  Error reading attachment: %v\n", err)
				} else {
					log.Debug().Int("size", len(data)).Str("filename", filename).Msg("Read attachment")
					fmt.Printf("  Size: %d bytes\n", len(data))
				}
			}

			partNum++
		}

		fetchCmd.Close()

		// Logout when done
		log.Debug().Msg("Logging out")
		if err := client.Logout().Wait(); err != nil {
			log.Fatal().Err(err).Msg("Failed to logout")
		}
		log.Info().Msg("Successfully completed fetch-content command")
	},
}

func init() {
	AddCommonFlags(FetchContentCmd)
}
