package main

import (
	"fmt"
	"mime"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/spf13/cobra"
)

var (
	server   string
	port     int
	username string
	password string
)

var listEmailsCmd = &cobra.Command{
	Use:   "list-emails",
	Short: "List the first 10 emails in the inbox",
	Long:  `Connect to an IMAP server and list the first 10 emails in the inbox, showing subject, from, and date.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listEmails()
	},
}

func init() {
	listEmailsCmd.Flags().StringVarP(&server, "server", "s", "", "IMAP server address (required)")
	listEmailsCmd.Flags().IntVarP(&port, "port", "p", 993, "IMAP server port")
	listEmailsCmd.Flags().StringVarP(&username, "username", "u", "", "IMAP username (required)")
	listEmailsCmd.Flags().StringVarP(&password, "password", "w", "", "IMAP password (required)")

	listEmailsCmd.MarkFlagRequired("server")
	listEmailsCmd.MarkFlagRequired("username")
	listEmailsCmd.MarkFlagRequired("password")

	RootCmd.AddCommand(listEmailsCmd)
}

func listEmails() error {
	// Connect to the server
	options := &imapclient.Options{
		WordDecoder: &mime.WordDecoder{}, // Using standard mime.WordDecoder
	}

	client, err := imapclient.DialTLS(fmt.Sprintf("%s:%d", server, port), options)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer client.Close()

	// Login
	if err := client.Login(username, password).Wait(); err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}

	// Select INBOX
	mbox, err := client.Select("INBOX", nil).Wait()
	if err != nil {
		return fmt.Errorf("failed to select inbox: %v", err)
	}

	// Calculate the sequence set for the last 10 messages
	var seqSet imap.SeqSet
	start := uint32(1)
	if mbox.NumMessages > 10 {
		start = mbox.NumMessages - 9
	}
	seqSet.AddRange(start, mbox.NumMessages)

	// Set up fetch options
	fetchOptions := &imap.FetchOptions{
		Envelope:     true,
		Flags:        true,
		InternalDate: true,
	}

	// Fetch messages
	fetchCmd := client.Fetch(seqSet, fetchOptions)
	defer fetchCmd.Close()

	// Print message info
	count := 0
	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		count++
		fmt.Printf("\nMessage %d:\n", count)

		msgData, err := msg.Collect()
		if err != nil {
			return fmt.Errorf("failed to collect message data: %v", err)
		}

		if msgData.Envelope != nil {
			fmt.Printf("Subject: %s\n", msgData.Envelope.Subject)
			fmt.Printf("From: %s\n", formatAddresses(msgData.Envelope.From))
			fmt.Printf("Date: %s\n", msgData.Envelope.Date)
		}
		if msgData.Flags != nil {
			fmt.Printf("Flags: %v\n", msgData.Flags)
		}
		fmt.Println("----------------------------------------")
	}

	if err := fetchCmd.Close(); err != nil {
		return fmt.Errorf("failed to close fetch command: %v", err)
	}

	return nil
}

func formatAddresses(addresses []imap.Address) string {
	if len(addresses) == 0 {
		return ""
	}
	addr := addresses[0]
	if addr.Name != "" {
		return fmt.Sprintf("%s <%s@%s>", addr.Name, addr.Mailbox, addr.Host)
	}
	return fmt.Sprintf("%s@%s", addr.Mailbox, addr.Host)
}
