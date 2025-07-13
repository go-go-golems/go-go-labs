package commands

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"

	"github.com/pkg/errors"
)

type ListCommand struct {
	*cmds.CommandDescription
}

type ListSettings struct {
	Tag    string `glazed.parameter:"tag"`
	Latest int    `glazed.parameter:"latest"`
}

var _ cmds.GlazeCommand = (*ListCommand)(nil)

func NewListCommand() (*ListCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	return &ListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List available knowledge snippet keys"),
			cmds.WithLong(`List available knowledge snippet keys, optionally filtered by tag.

This command shows all stored knowledge snippets (jots) with their metadata
including keys, authors, tags, and timestamps. Useful for discovering what
documentation and notes are available.

Filter by tag to find specific types of content, or use --latest to limit
results to the most recently created entries.

This is ideal for:
- Discovering available documentation
- Finding relevant code examples by tag
- Browsing shared knowledge base
- Checking what information is already stored

Example usage in agent tool calling:
  agentbus list
  agentbus list --tag "docker" --latest 10
  agentbus list --tag "api,auth"
  agentbus list --latest 20`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"tag",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by tag (comma-separated for multiple tags)"),
				),
				parameters.NewParameterDefinition(
					"latest",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Limit to N most recent entries"),
					parameters.WithDefault(50),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *ListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ListSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	client, err := getRedisClient()
	if err != nil {
		return err
	}
	defer client.Close()

	var keys []string

	if s.Tag != "" {
		// Get keys by tag
		tags := strings.Split(s.Tag, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			
			tagKey := client.JotsByTagKey(tag)
			tagKeys, err := client.ZRevRange(ctx, tagKey, 0, int64(s.Latest-1)).Result()
			if err != nil {
				return errors.Wrapf(err, "failed to get keys for tag '%s'", tag)
			}
			keys = append(keys, tagKeys...)
		}

		// Remove duplicates while preserving order
		seen := make(map[string]bool)
		uniqueKeys := make([]string, 0, len(keys))
		for _, key := range keys {
			if !seen[key] {
				seen[key] = true
				uniqueKeys = append(uniqueKeys, key)
			}
		}
		keys = uniqueKeys
	} else {
		// Get all jot keys by scanning the keyspace
		pattern := client.JotKey("*")
		scanKeys, err := client.Keys(ctx, pattern).Result()
		if err != nil {
			return errors.Wrap(err, "failed to scan jot keys")
		}

		// Extract just the key part (remove prefix)
		prefix := client.JotKey("")
		for _, fullKey := range scanKeys {
			if strings.HasPrefix(fullKey, prefix) {
				key := strings.TrimPrefix(fullKey, prefix)
				keys = append(keys, key)
			}
		}
	}

	// Limit results
	if len(keys) > s.Latest {
		keys = keys[:s.Latest]
	}

	// Get metadata for each key
	for _, key := range keys {
		jotKey := client.JotKey(key)
		
		result, err := client.HGetAll(ctx, jotKey).Result()
		if err != nil || len(result) == 0 {
			continue // Skip missing jots
		}

		// Parse timestamp
		timestamp, _ := strconv.ParseInt(result["timestamp"], 10, 64)
		
		// Parse tags
		var jotTags []string
		if result["tags"] != "" {
			jotTags = strings.Split(result["tags"], ",")
		}

		row := types.NewRow(
			types.MRP("key", key),
			types.MRP("author", result["author"]),
			types.MRP("tags", jotTags),
			types.MRP("timestamp", time.Unix(timestamp, 0).Format(time.RFC3339)),
			types.MRP("value_length", len(result["value"])),
		)

		err = gp.AddRow(ctx, row)
		if err != nil {
			return err
		}
	}

	return nil
}
