package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/go-go-golems/go-go-labs/cmd/apps/mail-app-rules/dsl"
)

type FetchMailCommand struct {
	*cmds.CommandDescription
}

type FetchMailSettings struct {
	// Search criteria settings
	Since            string   `glazed.parameter:"since"`
	Before           string   `glazed.parameter:"before"`
	WithinDays       int      `glazed.parameter:"within-days"`
	From             string   `glazed.parameter:"from"`
	To               string   `glazed.parameter:"to"`
	Subject          string   `glazed.parameter:"subject"`
	SubjectContains  string   `glazed.parameter:"subject-contains"`
	BodyContains     string   `glazed.parameter:"body-contains"`
	HasFlags         []string `glazed.parameter:"has-flags"`
	DoesNotHaveFlags []string `glazed.parameter:"not-has-flags"`
	LargerThan       string   `glazed.parameter:"larger-than"`
	SmallerThan      string   `glazed.parameter:"smaller-than"`

	// Output settings
	Limit                int    `glazed.parameter:"limit"`
	Format               string `glazed.parameter:"format"`
	IncludeContent       bool   `glazed.parameter:"include-content"`
	ConcatenateMimeParts bool   `glazed.parameter:"concatenate-mime-parts"`
	ContentMaxLength     int    `glazed.parameter:"content-max-length"`
	ContentType          string `glazed.parameter:"content-type"`
	PrintRule            bool   `glazed.parameter:"print-rule"`

	// IMAP settings
	IMAPSettings
}

func NewFetchMailCommand() (*FetchMailCommand, error) {
	layer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP layer: %w", err)
	}

	imapLayer, err := NewIMAPParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP layer: %w", err)
	}

	return &FetchMailCommand{
		CommandDescription: cmds.NewCommandDescription(
			"fetch-mail",
			cmds.WithShort("Fetch emails from an IMAP server using CLI arguments"),
			cmds.WithLong("This command connects to an IMAP server and fetches emails based on search criteria provided as command line arguments"),
			cmds.WithFlags(
				// Search criteria flags
				parameters.NewParameterDefinition(
					"since",
					parameters.ParameterTypeString,
					parameters.WithHelp("Fetch emails since date (YYYY-MM-DD)"),
				),
				parameters.NewParameterDefinition(
					"before",
					parameters.ParameterTypeString,
					parameters.WithHelp("Fetch emails before date (YYYY-MM-DD)"),
				),
				parameters.NewParameterDefinition(
					"within-days",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Fetch emails within the last N days"),
					parameters.WithDefault(0),
				),
				parameters.NewParameterDefinition(
					"from",
					parameters.ParameterTypeString,
					parameters.WithHelp("Fetch emails from a specific sender"),
				),
				parameters.NewParameterDefinition(
					"to",
					parameters.ParameterTypeString,
					parameters.WithHelp("Fetch emails sent to a specific recipient"),
				),
				parameters.NewParameterDefinition(
					"subject",
					parameters.ParameterTypeString,
					parameters.WithHelp("Fetch emails with an exact subject match"),
				),
				parameters.NewParameterDefinition(
					"subject-contains",
					parameters.ParameterTypeString,
					parameters.WithHelp("Fetch emails with subject containing a string"),
				),
				parameters.NewParameterDefinition(
					"body-contains",
					parameters.ParameterTypeString,
					parameters.WithHelp("Fetch emails with body containing a string"),
				),
				parameters.NewParameterDefinition(
					"has-flags",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Fetch emails with specific flags (comma-separated)"),
				),
				parameters.NewParameterDefinition(
					"not-has-flags",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Fetch emails without specific flags (comma-separated)"),
				),
				parameters.NewParameterDefinition(
					"larger-than",
					parameters.ParameterTypeString,
					parameters.WithHelp("Fetch emails larger than size (e.g., '1M', '500K')"),
				),
				parameters.NewParameterDefinition(
					"smaller-than",
					parameters.ParameterTypeString,
					parameters.WithHelp("Fetch emails smaller than size (e.g., '1M', '500K')"),
				),

				// Output flags
				parameters.NewParameterDefinition(
					"limit",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Maximum number of emails to fetch"),
					parameters.WithDefault(10),
				),
				parameters.NewParameterDefinition(
					"format",
					parameters.ParameterTypeString,
					parameters.WithHelp("Output format (json, text, table)"),
					parameters.WithDefault("text"),
				),
				parameters.NewParameterDefinition(
					"include-content",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Include email content in output"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"concatenate-mime-parts",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Concatenate all MIME parts into a single content string instead of showing structured output"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"content-max-length",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Maximum length of content to display"),
					parameters.WithDefault(1000),
				),
				parameters.NewParameterDefinition(
					"content-type",
					parameters.ParameterTypeString,
					parameters.WithHelp("MIME type to filter content (e.g., 'text/plain', 'text/*')"),
					parameters.WithDefault("text/plain"),
				),
				parameters.NewParameterDefinition(
					"print-rule",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Print the equivalent YAML rule instead of executing it"),
					parameters.WithDefault(false),
				),
			),
			cmds.WithLayersList(imapLayer, layer),
		),
	}, nil
}

func (c *FetchMailCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &FetchMailSettings{}
	if err := parsedLayers.InitializeStruct("default", settings); err != nil {
		return err
	}
	if err := parsedLayers.InitializeStruct("imap", &settings.IMAPSettings); err != nil {
		return err
	}

	// Build rule from command line arguments
	rule, err := c.buildRuleFromSettings(settings)
	if err != nil {
		return fmt.Errorf("error building rule from settings: %w", err)
	}

	// If print-rule is set, output the rule as YAML and return
	if settings.PrintRule {
		yamlData, err := yaml.Marshal(rule)
		if err != nil {
			return fmt.Errorf("error marshaling rule to YAML: %w", err)
		}

		// Create a row with the YAML data
		row := types.NewRow()
		row.Set("rule", string(yamlData))
		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("error adding rule to output: %w", err)
		}
		return nil
	}

	// Check if password is provided
	if settings.Password == "" {
		return fmt.Errorf("password is required (provide via --password flag or IMAP_PASSWORD environment variable)")
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

	// Fetch messages
	msgs, err := rule.FetchMessages(client)
	if err != nil {
		return fmt.Errorf("error fetching messages: %w", err)
	}

	// Process messages
	for _, msg := range msgs {
		// Create a new row for each message
		row := types.NewRow()

		// Always include UID
		row.Set("uid", msg.UID)

		// Always include basic email fields
		if msg.Envelope != nil {
			row.Set("subject", msg.Envelope.Subject)

			if len(msg.Envelope.From) > 0 {
				from := msg.Envelope.From[0]
				row.Set("from", fmt.Sprintf("%s <%s>", from.Name, from.Address))
			}

			if len(msg.Envelope.To) > 0 {
				var toAddresses []string
				for _, to := range msg.Envelope.To {
					toAddresses = append(toAddresses, fmt.Sprintf("%s <%s>", to.Name, to.Address))
				}
				row.Set("to", strings.Join(toAddresses, ", "))
			}

			row.Set("date", msg.Envelope.Date.Format(time.RFC3339))
		}

		// Always include flags and size
		row.Set("flags", strings.Join(msg.Flags, ", "))
		row.Set("size", msg.Size)

		// Handle content if requested
		if settings.IncludeContent && len(msg.MimeParts) > 0 {
			if settings.ConcatenateMimeParts {
				// Concatenate all matching MIME parts into a single content string
				var contents []string
				for _, part := range msg.MimeParts {
					// Fix: Only add slash if Subtype is not empty
					mimeType := part.Type
					if part.Subtype != "" {
						mimeType = part.Type + "/" + part.Subtype
					}

					if c.shouldIncludeMimeType(mimeType, settings.ContentType) {
						contents = append(contents, part.Content)
						log.Debug().
							Str("mime_type", mimeType).
							Str("filter", settings.ContentType).
							Bool("included", true).
							Int("content_length", len(part.Content)).
							Msg("Added MIME part content")
					} else {
						log.Debug().
							Str("mime_type", mimeType).
							Str("filter", settings.ContentType).
							Bool("included", false).
							Msg("Excluded MIME part content")
					}
				}
				content := strings.Join(contents, "\n\n")
				if settings.ContentMaxLength > 0 && len(content) > settings.ContentMaxLength {
					content = content[:settings.ContentMaxLength] + "..."
				}
				row.Set("content", content)
				log.Debug().
					Int("total_parts", len(msg.MimeParts)).
					Int("matched_parts", len(contents)).
					Int("final_content_length", len(content)).
					Msg("Finished processing MIME parts")
			} else {
				// Structured MIME parts output
				var parts []map[string]interface{}
				for _, part := range msg.MimeParts {
					// Fix: Only add slash if Subtype is not empty
					mimeType := part.Type
					if part.Subtype != "" {
						mimeType = part.Type + "/" + part.Subtype
					}

					if c.shouldIncludeMimeType(mimeType, settings.ContentType) {
						partMap := map[string]interface{}{
							"type":    mimeType,
							"size":    part.Size,
							"charset": part.Charset,
						}
						if part.Filename != "" {
							partMap["filename"] = part.Filename
						}

						content := part.Content
						if settings.ContentMaxLength > 0 && len(content) > settings.ContentMaxLength {
							content = content[:settings.ContentMaxLength] + "..."
						}
						partMap["content"] = content

						parts = append(parts, partMap)
						log.Debug().
							Str("mime_type", mimeType).
							Str("filter", settings.ContentType).
							Bool("included", true).
							Int("content_length", len(content)).
							Msg("Added structured MIME part")
					} else {
						log.Debug().
							Str("mime_type", mimeType).
							Str("filter", settings.ContentType).
							Bool("included", false).
							Msg("Excluded structured MIME part")
					}
				}
				row.Set("mime_parts", parts)
				log.Debug().
					Int("total_parts", len(msg.MimeParts)).
					Int("matched_parts", len(parts)).
					Msg("Finished processing structured MIME parts")
			}
		}

		// Add the row to the processor
		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("error adding row to processor: %w", err)
		}
	}

	return nil
}

// Build a Rule struct from command line settings
func (c *FetchMailCommand) buildRuleFromSettings(settings *FetchMailSettings) (*dsl.Rule, error) {
	// Start building the search config
	searchConfig := dsl.SearchConfig{
		Since:           settings.Since,
		Before:          settings.Before,
		WithinDays:      settings.WithinDays,
		From:            settings.From,
		To:              settings.To,
		Subject:         settings.Subject,
		SubjectContains: settings.SubjectContains,
		BodyContains:    settings.BodyContains,
	}

	// Add flag criteria if specified
	if len(settings.HasFlags) > 0 || len(settings.DoesNotHaveFlags) > 0 {
		searchConfig.Flags = &dsl.FlagCriteria{
			Has:    settings.HasFlags,
			NotHas: settings.DoesNotHaveFlags,
		}
	}

	// Add size criteria if specified
	if settings.LargerThan != "" || settings.SmallerThan != "" {
		searchConfig.Size = &dsl.SizeCriteria{
			LargerThan:  settings.LargerThan,
			SmallerThan: settings.SmallerThan,
		}
	}

	// Build fields for output config
	var fields []interface{}

	// Always include basic email fields
	fields = append(fields,
		dsl.Field{Name: "uid"},
		dsl.Field{Name: "subject"},
		dsl.Field{Name: "from"},
		dsl.Field{Name: "to"},
		dsl.Field{Name: "date"},
		dsl.Field{Name: "flags"},
		dsl.Field{Name: "size"},
	)

	// Add content field if needed
	if settings.IncludeContent {
		contentField := &dsl.ContentField{
			ShowContent: true,
			MaxLength:   settings.ContentMaxLength,
		}

		// Set types for filtering
		if settings.ContentType != "" {
			contentField.Mode = "filter"
			contentField.Types = []string{settings.ContentType}
		}

		fields = append(fields, dsl.Field{
			Name:    "mime_parts",
			Content: contentField,
		})
	}

	// Create output config
	outputConfig := dsl.OutputConfig{
		Format: settings.Format,
		Limit:  settings.Limit,
		Fields: fields,
	}

	// Create the rule
	rule := &dsl.Rule{
		Name:        "cli-rule",
		Description: "Rule generated from command line arguments",
		Search:      searchConfig,
		Output:      outputConfig,
	}

	// Validate the rule
	if err := rule.Validate(); err != nil {
		return nil, fmt.Errorf("invalid rule: %w", err)
	}

	return rule, nil
}

func (c *FetchMailCommand) selectMailbox(client *imapclient.Client, mailbox string) error {
	if _, err := client.Select(mailbox, nil).Wait(); err != nil {
		return fmt.Errorf("failed to select mailbox %q: %w", mailbox, err)
	}
	return nil
}

func (c *FetchMailCommand) shouldIncludeMimeType(mimeType string, filter string) bool {
	// If no filter, include all
	if filter == "" {
		return true
	}

	log.Debug().
		Str("mime_type", mimeType).
		Str("filter", filter).
		Msg("Checking MIME type match")

	// Exact match
	if mimeType == filter {
		log.Debug().Msg("Exact match")
		return true
	}

	// Wildcard match (e.g., text/*)
	if strings.HasSuffix(filter, "/*") {
		prefix := strings.TrimSuffix(filter, "/*")
		result := strings.HasPrefix(mimeType, prefix+"/")
		log.Debug().
			Str("prefix", prefix).
			Bool("wildcard_match", result).
			Msg("Checking wildcard match")
		return result
	}

	log.Debug().Msg("No match found")
	return false
}
