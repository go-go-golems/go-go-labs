package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"

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

	msgs, err := rule.FetchMessages(client)
	if err != nil {
		return fmt.Errorf("error fetching messages: %w", err)
	}

	for _, msg := range msgs {
		// Create a new row for each message
		row := types.NewRow()

		// Process each field according to the rule's output configuration
		for _, fieldInterface := range rule.Output.Fields {
			field, ok := fieldInterface.(dsl.Field)
			if !ok {
				continue
			}

			switch field.Name {
			case "uid":
				row.Set("uid", msg.UID)
			case "subject":
				if msg.Envelope != nil {
					row.Set("subject", msg.Envelope.Subject)
				}
			case "from":
				if msg.Envelope != nil && len(msg.Envelope.From) > 0 {
					from := msg.Envelope.From[0]
					row.Set("from", fmt.Sprintf("%s <%s>", from.Name, from.Address))
				}
			case "to":
				if msg.Envelope != nil && len(msg.Envelope.To) > 0 {
					var toAddresses []string
					for _, to := range msg.Envelope.To {
						toAddresses = append(toAddresses, fmt.Sprintf("%s <%s>", to.Name, to.Address))
					}
					row.Set("to", strings.Join(toAddresses, ", "))
				}
			case "date":
				if msg.Envelope != nil {
					row.Set("date", msg.Envelope.Date.Format(time.RFC3339))
				}
			case "flags":
				row.Set("flags", strings.Join(msg.Flags, ", "))
			case "size":
				row.Set("size", msg.Size)
			case "mime_parts":
				if field.Content != nil && len(msg.MimeParts) > 0 {
					var parts []map[string]interface{}
					for _, part := range msg.MimeParts {
						if field.Content.ShouldInclude(part.Type + "/" + part.Subtype) {
							partMap := map[string]interface{}{
								"type":    part.Type + "/" + part.Subtype,
								"size":    part.Size,
								"charset": part.Charset,
							}
							if part.Filename != "" {
								partMap["filename"] = part.Filename
							}
							if field.Content.ShowContent {
								content := part.Content
								if field.Content.MaxLength > 0 && len(content) > field.Content.MaxLength {
									content = content[:field.Content.MaxLength] + "..."
								}
								partMap["content"] = content
							}
							parts = append(parts, partMap)
						}
					}
					row.Set("mime_parts", parts)
				}
			}
		}

		// Add the row to the processor
		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("error adding row to processor: %w", err)
		}
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
