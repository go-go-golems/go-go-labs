package commands

import (
	"crypto/tls"
	"fmt"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// IMAPSettings represents the settings for connecting to an IMAP server
type IMAPSettings struct {
	Server   string `glazed.parameter:"server"`
	Port     int    `glazed.parameter:"port"`
	Username string `glazed.parameter:"username"`
	Password string `glazed.parameter:"password"`
	Mailbox  string `glazed.parameter:"mailbox"`
	Insecure bool   `glazed.parameter:"insecure"`
}

// NewIMAPParameterLayer creates a new parameter layer for IMAP server settings
func NewIMAPParameterLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"imap",
		"IMAP Server Connection Settings",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"server",
				parameters.ParameterTypeString,
				parameters.WithHelp("IMAP server address"),
			),
			parameters.NewParameterDefinition(
				"port",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("IMAP server port"),
				parameters.WithDefault(993),
			),
			parameters.NewParameterDefinition(
				"username",
				parameters.ParameterTypeString,
				parameters.WithHelp("IMAP username"),
			),
			parameters.NewParameterDefinition(
				"password",
				parameters.ParameterTypeString,
				parameters.WithHelp("IMAP password"),
			),
			parameters.NewParameterDefinition(
				"mailbox",
				parameters.ParameterTypeString,
				parameters.WithHelp("Mailbox to search in"),
				parameters.WithDefault("INBOX"),
			),
			parameters.NewParameterDefinition(
				"insecure",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Skip TLS verification"),
				parameters.WithDefault(false),
			),
		),
	)
}

func (s *IMAPSettings) ConnectToIMAPServer() (*imapclient.Client, error) {
	serverAddr := fmt.Sprintf("%s:%d", s.Server, s.Port)

	options := &imapclient.Options{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: s.Insecure,
		},
	}

	client, err := imapclient.DialTLS(serverAddr, options)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %w", err)
	}

	if err := client.Login(s.Username, s.Password).Wait(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return client, nil
}
