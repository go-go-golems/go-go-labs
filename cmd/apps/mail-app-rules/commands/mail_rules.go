package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"

	"github.com/go-go-golems/go-go-labs/cmd/apps/mail-app-rules/dsl"
)

type MailRulesCommand struct {
	*cmds.CommandDescription
}

type MailRulesSettings struct {
	RuleFile string `glazed.parameter:"rule"`
	IMAPSettings
}

func NewMailRulesCommand() (*MailRulesCommand, error) {
	layer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP layer: %w", err)
	}

	imapLayer, err := NewIMAPParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP layer: %w", err)
	}

	return &MailRulesCommand{
		CommandDescription: cmds.NewCommandDescription(
			"mail-rules",
			cmds.WithShort("Process mail rules on an IMAP server"),
			cmds.WithLong("This command connects to an IMAP server and processes mail rules defined in a YAML file"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"rule",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to YAML rule file"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(imapLayer, layer),
		),
	}, nil
}

func (c *MailRulesCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &MailRulesSettings{}
	if err := parsedLayers.InitializeStruct("default", settings); err != nil {
		return err
	}
	if err := parsedLayers.InitializeStruct("imap", &settings.IMAPSettings); err != nil {
		return err
	}

	// Read password from environment if not provided
	if settings.Password == "" {
		return fmt.Errorf("password is required (provide via --password flag or IMAP_PASSWORD environment variable)")
	}

	// Parse rule file
	rule, err := c.parseRuleFile(settings.RuleFile)
	if err != nil {
		return fmt.Errorf("error parsing rule file: %w", err)
	}

	// Connect to IMAP server
	client, err := settings.IMAPSettings.ConnectToIMAPServer()
	if err != nil {
		return fmt.Errorf("error connecting to IMAP server: %w", err)
	}
	defer client.Close()

	// Select mailbox
	if err := c.selectMailbox(client, settings.Mailbox); err != nil {
		return fmt.Errorf("error selecting mailbox: %w", err)
	}

	// Process rule
	if err := dsl.ProcessRule(client, rule); err != nil {
		return fmt.Errorf("error processing rule: %w", err)
	}

	return nil
}

func (c *MailRulesCommand) parseRuleFile(path string) (*dsl.Rule, error) {
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

func (c *MailRulesCommand) selectMailbox(client *imapclient.Client, mailbox string) error {
	// Select mailbox
	if _, err := client.Select(mailbox, nil).Wait(); err != nil {
		return fmt.Errorf("failed to select mailbox %q: %w", mailbox, err)
	}
	return nil
}
