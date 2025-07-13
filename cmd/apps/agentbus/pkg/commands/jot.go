package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type JotCommand struct {
	*cmds.CommandDescription
}

type JotSettings struct {
	Key      string `glazed.parameter:"key"`
	Value    string `glazed.parameter:"value"`
	Tag      string `glazed.parameter:"tag"`
	IfAbsent bool   `glazed.parameter:"if-absent"`
}

var _ cmds.GlazeCommand = (*JotCommand)(nil)

func NewJotCommand() (*JotCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	return &JotCommand{
		CommandDescription: cmds.NewCommandDescription(
			"jot",
			cmds.WithShort("Store a knowledge snippet or documentation note"),
			cmds.WithLong(`Store a knowledge snippet with optional tags for later retrieval.

This creates a persistent key-value store entry that can be tagged and searched.
Perfect for storing documentation, lessons learned, configuration snippets,
or any information that needs to be shared across agents or sessions.

The value can be:
- Documentation snippets
- Configuration examples  
- Code templates
- Troubleshooting notes
- Best practices
- API endpoints and examples

Tags help organize and filter content for easy discovery.

Example usage in agent tool calling:
  agentbus jot --key "docker-build-pattern" --value "docker build -t app ." --tag "docker,build"
  agentbus jot --key "api-auth-header" --value "Authorization: Bearer \${TOKEN}" --tag "api,auth"
  agentbus jot --key "deploy-checklist" --value "$(cat checklist.md)" --tag "deploy,docs"`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"key",
					parameters.ParameterTypeString,
					parameters.WithHelp("Unique identifier for this knowledge snippet"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"value",
					parameters.ParameterTypeString,
					parameters.WithHelp("Content to store (use $(cat file) to read from file)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"tag",
					parameters.ParameterTypeString,
					parameters.WithHelp("Comma-separated tags for categorization"),
				),
				parameters.NewParameterDefinition(
					"if-absent",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Only store if key doesn't already exist"),
					parameters.WithDefault(false),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *JotCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &JotSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	agentID, err := getAgentID()
	if err != nil {
		return err
	}

	client, err := getRedisClient()
	if err != nil {
		return err
	}
	defer client.Close()

	jotKey := client.JotKey(s.Key)
	now := time.Now()

	// Check if key exists when if-absent is specified
	if s.IfAbsent {
		exists, err := client.Exists(ctx, jotKey).Result()
		if err != nil {
			return errors.Wrap(err, "failed to check if key exists")
		}
		if exists > 0 {
			return errors.New("key already exists and --if-absent was specified")
		}
	}

	// Parse tags
	var tags []string
	if s.Tag != "" {
		tags = strings.Split(s.Tag, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	// Store the jot as a hash
	jotData := map[string]interface{}{
		"value":     s.Value,
		"author":    agentID,
		"timestamp": now.Unix(),
		"tags":      strings.Join(tags, ","),
	}

	err = client.HMSet(ctx, jotKey, jotData).Err()
	if err != nil {
		return errors.Wrap(err, "failed to store jot")
	}

	// Add to tag indices
	for _, tag := range tags {
		if tag != "" {
			tagKey := client.JotsByTagKey(tag)
			err = client.ZAdd(ctx, tagKey, redis.Z{
				Score:  float64(now.Unix()),
				Member: s.Key,
			}).Err()
			if err != nil {
				return errors.Wrap(err, "failed to update tag index")
			}
		}
	}

	// Auto-publish to communication channel
	message := fmt.Sprintf("📝 Stored knowledge snippet '%s'", s.Key)
	if len(tags) > 0 {
		message += fmt.Sprintf(" (tags: %s)", strings.Join(tags, ", "))
	}
	err = publishToChannel(ctx, client, agentID, message, "knowledge")
	if err != nil {
		// Don't fail the command if publish fails, just log it
		// In a real implementation, you might want to use a logger here
	}

	// Output the result
	row := types.NewRow(
		types.MRP("key", s.Key),
		types.MRP("author", agentID),
		types.MRP("tags", tags),
		types.MRP("timestamp", now.Format(time.RFC3339)),
		types.MRP("value_length", len(s.Value)),
	)

	return gp.AddRow(ctx, row)
}
