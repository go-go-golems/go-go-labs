package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-go-golems/go-go-labs/cmd/apps/mail-app-rules/dsl"
)

func main() {
	// Parse command line flags
	var (
		ruleFile string
		server   string
		port     int
		username string
		password string
		mailbox  string
		insecure bool
	)

	flag.StringVar(&ruleFile, "rule", "", "Path to YAML rule file")
	flag.StringVar(&server, "server", "", "IMAP server address")
	flag.IntVar(&port, "port", 993, "IMAP server port")
	flag.StringVar(&username, "username", "", "IMAP username")
	flag.StringVar(&password, "password", "", "IMAP password")
	flag.StringVar(&mailbox, "mailbox", "INBOX", "Mailbox to search in")
	flag.BoolVar(&insecure, "insecure", false, "Skip TLS verification")
	flag.Parse()

	// Validate required flags
	if ruleFile == "" {
		fmt.Println("Error: rule file is required")
		flag.Usage()
		os.Exit(1)
	}

	if server == "" {
		fmt.Println("Error: server is required")
		flag.Usage()
		os.Exit(1)
	}

	if username == "" {
		fmt.Println("Error: username is required")
		flag.Usage()
		os.Exit(1)
	}

	// Read password from environment if not provided
	if password == "" {
		password = os.Getenv("IMAP_PASSWORD")
		if password == "" {
			fmt.Println("Error: password is required (provide via -password flag or IMAP_PASSWORD environment variable)")
			os.Exit(1)
		}
	}

	// Parse rule file
	rule, err := parseRuleFile(ruleFile)
	if err != nil {
		fmt.Printf("Error parsing rule file: %v\n", err)
		os.Exit(1)
	}

	// Connect to IMAP server
	client, err := connectToIMAPServer(server, port, username, password, insecure)
	if err != nil {
		fmt.Printf("Error connecting to IMAP server: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Select mailbox
	if err := selectMailbox(client, mailbox); err != nil {
		fmt.Printf("Error selecting mailbox: %v\n", err)
		os.Exit(1)
	}

	// Process rule
	if err := dsl.ProcessRule(client, rule); err != nil {
		fmt.Printf("Error processing rule: %v\n", err)
		os.Exit(1)
	}
}

func parseRuleFile(path string) (*dsl.Rule, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("rule file does not exist: %s", path)
	}

	// Parse rule file
	rule, err := dsl.ParseRuleFile(path)
	if err != nil {
		return nil, err
	}

	return rule, nil
}

func connectToIMAPServer(server string, port int, username, password string, insecure bool) (*imapclient.Client, error) {
	// Build server address
	serverAddr := fmt.Sprintf("%s:%d", server, port)

	// Connect to server
	options := &imapclient.Options{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	client, err := imapclient.DialTLS(serverAddr, options)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %w", err)
	}

	// Login
	if err := client.Login(username, password).Wait(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return client, nil
}

func selectMailbox(client *imapclient.Client, mailbox string) error {
	// Select mailbox
	if _, err := client.Select(mailbox, nil).Wait(); err != nil {
		return fmt.Errorf("failed to select mailbox %q: %w", mailbox, err)
	}
	return nil
}
