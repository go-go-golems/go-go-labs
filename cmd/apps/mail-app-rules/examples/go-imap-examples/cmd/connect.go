package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// ConnectCmd demonstrates connecting to an IMAP server
var ConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to an IMAP server",
	Long: `Demonstrates how to connect to an IMAP server, authenticate,
and select a mailbox.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := ConnectToIMAP()
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer client.Close()

		fmt.Println("Successfully connected to IMAP server!")
		fmt.Printf("Server: %s:%d\n", server, port)
		fmt.Printf("Username: %s\n", username)
		fmt.Printf("Selected mailbox: %s\n", mailbox)

		// Logout when done
		if err := client.Logout().Wait(); err != nil {
			log.Fatalf("Failed to logout: %v", err)
		}
		fmt.Println("Successfully logged out")
	},
}

func init() {
	AddCommonFlags(ConnectCmd)
}
