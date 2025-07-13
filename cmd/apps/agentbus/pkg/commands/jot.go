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
	"github.com/rs/zerolog/log"
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
		log.Error().Err(err).Msg("Failed to initialize jot settings")
		return err
	}

	log.Info().
		Str("key", s.Key).
		Str("tag", s.Tag).
		Bool("if_absent", s.IfAbsent).
		Int("value_length", len(s.Value)).
		Msg("Starting jot operation")

	agentID, err := getAgentID()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get agent ID")
		return err
	}

	log.Debug().Str("agent_id", agentID).Msg("Retrieved agent ID")

	client, err := getRedisClient()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Redis client")
		return err
	}
	defer client.Close()

	jotKey := client.JotKey(s.Key)
	now := time.Now()

	// Check if key exists when if-absent is specified
	if s.IfAbsent {
		log.Debug().Str("key", s.Key).Msg("Checking if key exists for if-absent condition")
		exists, err := client.Exists(ctx, jotKey).Result()
		if err != nil {
			log.Error().Err(err).Str("key", s.Key).Msg("Failed to check if key exists")
			return errors.Wrap(err, "failed to check if key exists")
		}
		if exists > 0 {
			log.Warn().Str("key", s.Key).Msg("Key already exists and --if-absent was specified")
			return errors.New("key already exists and --if-absent was specified")
		}
		log.Debug().Str("key", s.Key).Msg("Key does not exist, proceeding with storage")
	}

	// Parse tags
	var tags []string
	if s.Tag != "" {
		log.Debug().Str("raw_tags", s.Tag).Msg("Parsing tags")
		tags = strings.Split(s.Tag, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
		log.Debug().Strs("parsed_tags", tags).Msg("Parsed tags successfully")
	}

	// Store the jot as a hash
	jotData := map[string]interface{}{
		"value":     s.Value,
		"author":    agentID,
		"timestamp": now.Unix(),
		"tags":      strings.Join(tags, ","),
	}

	log.Debug().
		Str("key", s.Key).
		Str("agent_id", agentID).
		Int("value_length", len(s.Value)).
		Strs("tags", tags).
		Msg("Storing jot data to Redis")

	err = client.HMSet(ctx, jotKey, jotData).Err()
	if err != nil {
		log.Error().Err(err).Str("key", s.Key).Msg("Failed to store jot")
		return errors.Wrap(err, "failed to store jot")
	}

	log.Debug().Str("key", s.Key).Msg("Successfully stored jot data")

	// Add to tag indices
	for _, tag := range tags {
		if tag != "" {
			log.Debug().Str("tag", tag).Str("key", s.Key).Msg("Adding to tag index")
			tagKey := client.JotsByTagKey(tag)
			err = client.ZAdd(ctx, tagKey, redis.Z{
				Score:  float64(now.Unix()),
				Member: s.Key,
			}).Err()
			if err != nil {
				log.Error().Err(err).Str("tag", tag).Str("key", s.Key).Msg("Failed to update tag index")
				return errors.Wrap(err, "failed to update tag index")
			}
			log.Debug().Str("tag", tag).Str("key", s.Key).Msg("Successfully added to tag index")
		}
	}

	// Auto-publish to communication channel
	message := fmt.Sprintf("📝 Stored knowledge snippet '%s'", s.Key)
	if len(tags) > 0 {
		message += fmt.Sprintf(" (tags: %s)", strings.Join(tags, ", "))
	}
	log.Debug().Str("message", message).Msg("Publishing to communication channel")
	err = publishToChannel(ctx, client, agentID, message, "knowledge")
	if err != nil {
		log.Warn().Err(err).Str("key", s.Key).Msg("Failed to publish to communication channel")
	}

	log.Info().
		Str("key", s.Key).
		Str("agent_id", agentID).
		Strs("tags", tags).
		Int("value_length", len(s.Value)).
		Msg("Successfully stored knowledge snippet")

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
