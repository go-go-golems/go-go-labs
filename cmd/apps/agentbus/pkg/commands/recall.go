package commands

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	agentredis "github.com/go-go-golems/go-go-labs/cmd/apps/agentbus/pkg/redis"
	"github.com/pkg/errors"
)

type RecallCommand struct {
	*cmds.CommandDescription
}

type RecallSettings struct {
	Key    string `glazed.parameter:"key"`
	Tag    string `glazed.parameter:"tag"`
	Latest int    `glazed.parameter:"latest"`
}

var _ cmds.GlazeCommand = (*RecallCommand)(nil)
var _ cmds.WriterCommand = (*RecallCommand)(nil)

func NewRecallCommand() (*RecallCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	return &RecallCommand{
		CommandDescription: cmds.NewCommandDescription(
			"recall",
			cmds.WithShort("Retrieve knowledge snippets by key or tag"),
			cmds.WithLong(`Retrieve previously stored knowledge snippets (jots) by key or tag.

Use this to access documentation, configuration examples, code snippets,
or any other information previously stored by agents.

Retrieval modes:
- By key: Get a specific snippet by its exact key
- By tag: Get all snippets with a specific tag (reverse chronological order)
- Latest: Limit tag results to most recent N entries

This is ideal for:
- Looking up configuration templates
- Finding code examples and patterns
- Accessing shared documentation
- Retrieving troubleshooting guides
- Getting deployment procedures

Example usage in agent tool calling:
  agentbus recall --key "docker-build-pattern"
  agentbus recall --tag "deploy" --latest 5
  agentbus recall --tag "api,auth"
  agentbus recall --tag "troubleshooting"`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"key",
					parameters.ParameterTypeString,
					parameters.WithHelp("Specific key to retrieve"),
				),
				parameters.NewParameterDefinition(
					"tag",
					parameters.ParameterTypeString,
					parameters.WithHelp("Tag to filter by (comma-separated for multiple tags)"),
				),
				parameters.NewParameterDefinition(
					"latest",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Limit to N most recent entries (when using --tag)"),
					parameters.WithDefault(10),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *RecallCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &RecallSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	if s.Key == "" && s.Tag == "" {
		return errors.New("either --key or --tag must be specified")
	}

	if s.Key != "" && s.Tag != "" {
		return errors.New("--key and --tag are mutually exclusive")
	}

	client, err := getRedisClient()
	if err != nil {
		return err
	}
	defer client.Close()

	if s.Key != "" {
		// Retrieve specific jot by key
		return c.retrieveByKey(ctx, client, s.Key, gp)
	} else {
		// Retrieve jots by tag
		return c.retrieveByTag(ctx, client, s.Tag, s.Latest, gp)
	}
}

func (c *RecallCommand) retrieveByKey(
	ctx context.Context,
	client *agentredis.Client,
	key string,
	gp middlewares.Processor,
) error {
	jotKey := client.JotKey(key)

	result, err := client.HGetAll(ctx, jotKey).Result()
	if err != nil {
		return errors.Wrap(err, "failed to retrieve jot")
	}

	if len(result) == 0 {
		return errors.New("jot not found")
	}

	// Parse timestamp
	timestamp, _ := strconv.ParseInt(result["timestamp"], 10, 64)

	// Parse tags
	var tags []string
	if result["tags"] != "" {
		tags = strings.Split(result["tags"], ",")
	}

	row := types.NewRow(
		types.MRP("key", key),
		types.MRP("value", result["value"]),
		types.MRP("author", result["author"]),
		types.MRP("tags", tags),
		types.MRP("timestamp", time.Unix(timestamp, 0).Format(time.RFC3339)),
	)

	return gp.AddRow(ctx, row)
}

func (c *RecallCommand) retrieveByTag(
	ctx context.Context,
	client *agentredis.Client,
	tagStr string,
	latest int,
	gp middlewares.Processor,
) error {
	tags := strings.Split(tagStr, ",")
	var allKeys []string

	// Get keys from all specified tags
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		tagKey := client.JotsByTagKey(tag)

		// Get keys in reverse chronological order (highest scores first)
		keys, err := client.ZRevRange(ctx, tagKey, 0, int64(latest-1)).Result()
		if err != nil {
			return errors.Wrapf(err, "failed to get keys for tag '%s'", tag)
		}

		allKeys = append(allKeys, keys...)
	}

	// Remove duplicates while preserving order
	seen := make(map[string]bool)
	uniqueKeys := make([]string, 0, len(allKeys))
	for _, key := range allKeys {
		if !seen[key] {
			seen[key] = true
			uniqueKeys = append(uniqueKeys, key)
		}
	}

	// Limit to latest N
	if len(uniqueKeys) > latest {
		uniqueKeys = uniqueKeys[:latest]
	}

	// Retrieve each jot
	for _, key := range uniqueKeys {
		jotKey := client.JotKey(key)

		result, err := client.HGetAll(ctx, jotKey).Result()
		if err != nil {
			continue // Skip missing jots
		}

		if len(result) == 0 {
			continue
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
			types.MRP("value", result["value"]),
			types.MRP("author", result["author"]),
			types.MRP("tags", jotTags),
			types.MRP("timestamp", time.Unix(timestamp, 0).Format(time.RFC3339)),
		)

		err = gp.AddRow(ctx, row)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *RecallCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	s := &RecallSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	client, err := getRedisClient()
	if err != nil {
		return err
	}
	defer client.Close()

	if s.Key != "" {
		return c.retrieveByKeyWriter(ctx, client, s.Key, w)
	} else if s.Tag != "" {
		return c.retrieveByTagWriter(ctx, client, s.Tag, s.Latest, w)
	} else {
		return c.retrieveAllWriter(ctx, client, s.Latest, w)
	}
}

func (c *RecallCommand) retrieveByKeyWriter(
	ctx context.Context,
	client *agentredis.Client,
	key string,
	w io.Writer,
) error {
	result, err := client.HGetAll(ctx, key).Result()
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve knowledge snippet for key %s", key)
	}

	if len(result) == 0 {
		fmt.Fprintf(w, "🔍 No knowledge snippet found for key '%s'\n", key)
		return nil
	}

	// Output human-readable format
	fmt.Fprintf(w, "📚 Knowledge snippet: %s\n", key)

	if author, ok := result["author"]; ok {
		fmt.Fprintf(w, "   Author: %s\n", author)
	}

	if tags, ok := result["tags"]; ok && tags != "" {
		fmt.Fprintf(w, "   Tags: %s\n", tags)
	}

	if timestamp, ok := result["timestamp"]; ok {
		fmt.Fprintf(w, "   Created: %s\n", timestamp)
	}

	if value, ok := result["value"]; ok {
		fmt.Fprintf(w, "   Content:\n")
		// Indent the content
		lines := strings.Split(value, "\n")
		for _, line := range lines {
			fmt.Fprintf(w, "     %s\n", line)
		}
	}

	return nil
}

func (c *RecallCommand) retrieveByTagWriter(
	ctx context.Context,
	client *agentredis.Client,
	tagStr string,
	latest int,
	w io.Writer,
) error {
	tagKey := fmt.Sprintf("tag:%s", tagStr)
	keys, err := client.SMembers(ctx, tagKey).Result()
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve keys for tag %s", tagStr)
	}

	if len(keys) == 0 {
		fmt.Fprintf(w, "🏷️  No knowledge snippets found with tag '%s'\n", tagStr)
		return nil
	}

	fmt.Fprintf(w, "🏷️  Knowledge snippets with tag '%s':\n\n", tagStr)

	count := 0
	for _, key := range keys {
		if latest > 0 && count >= latest {
			break
		}

		result, err := client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		if len(result) == 0 {
			continue
		}

		fmt.Fprintf(w, "📚 %s\n", key)

		if author, ok := result["author"]; ok {
			fmt.Fprintf(w, "   Author: %s\n", author)
		}

		if timestamp, ok := result["timestamp"]; ok {
			fmt.Fprintf(w, "   Created: %s\n", timestamp)
		}

		if value, ok := result["value"]; ok {
			valuePreview := value
			if len(valuePreview) > 100 {
				valuePreview = valuePreview[:97] + "..."
			}
			valuePreview = strings.ReplaceAll(valuePreview, "\n", " ")
			fmt.Fprintf(w, "   Content: %s\n", valuePreview)
		}

		fmt.Fprintf(w, "\n")
		count++
	}

	return nil
}

func (c *RecallCommand) retrieveAllWriter(
	ctx context.Context,
	client *agentredis.Client,
	latest int,
	w io.Writer,
) error {
	fmt.Fprintf(w, "📚 All knowledge snippets:\n\n")

	// This is a simplified implementation
	// In a real scenario, you'd want to implement proper pagination
	fmt.Fprintf(w, "Use --tag to filter by specific tags or --key to retrieve specific snippets\n")

	return nil
}
