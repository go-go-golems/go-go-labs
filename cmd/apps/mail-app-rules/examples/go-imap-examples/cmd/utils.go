package cmd

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/spf13/cobra"
)

// Common flags for all commands
var (
	server   string
	port     int
	username string
	password string
	mailbox  string
	useSSL   bool
	uid      uint32
)

// AddCommonFlags adds common IMAP connection flags to a command
func AddCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&server, "server", "s", "", "IMAP server address (required)")
	cmd.Flags().IntVarP(&port, "port", "p", 993, "IMAP server port")
	cmd.Flags().StringVarP(&username, "username", "u", "", "IMAP username (required)")
	cmd.Flags().StringVarP(&password, "password", "w", "", "IMAP password (required)")
	cmd.Flags().StringVarP(&mailbox, "mailbox", "m", "INBOX", "Mailbox to select")
	cmd.Flags().BoolVar(&useSSL, "ssl", true, "Use SSL/TLS connection")
	cmd.Flags().Uint32Var(&uid, "uid", 0, "Message UID to fetch (when applicable)")

	cmd.MarkFlagRequired("server")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
}

// ConnectToIMAP establishes a connection to the IMAP server
func ConnectToIMAP() (*imapclient.Client, error) {
	var client *imapclient.Client
	var err error

	serverAddr := fmt.Sprintf("%s:%d", server, port)

	if useSSL {
		client, err = imapclient.DialTLS(serverAddr, nil)
	} else {
		// For non-TLS connections, we need to establish a TCP connection first
		conn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to IMAP server: %v", err)
		}

		client = imapclient.New(conn, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %v", err)
	}

	if err := client.Login(username, password).Wait(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to login: %v", err)
	}

	if _, err := client.Select(mailbox, nil).Wait(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to select mailbox %s: %v", mailbox, err)
	}

	return client, nil
}

// CheckError logs the error and exits if err is not nil
func CheckError(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
		os.Exit(1)
	}
}
