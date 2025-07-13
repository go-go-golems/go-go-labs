package commands

import (
	"context"
	"fmt"
	"io"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	agentredis "github.com/go-go-golems/go-go-labs/cmd/apps/agentbus/pkg/redis"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type ClearCommand struct {
	*cmds.CommandDescription
}

type ClearSettings struct {
	Force bool `glazed.parameter:"force"`
}

var _ cmds.GlazeCommand = (*ClearCommand)(nil)
var _ cmds.WriterCommand = (*ClearCommand)(nil)

func NewClearCommand() (*ClearCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	return &ClearCommand{
		CommandDescription: cmds.NewCommandDescription(
			"clear",
			cmds.WithShort("Clear all AgentBus data from Redis"),
			cmds.WithLong(`Clear all AgentBus data from Redis including messages, jots, flags, and agent positions.

This command will delete ALL AgentBus data for the current project stored in Redis:
- Chat stream messages (agentbus:{PROJECT_PREFIX}:ch:*)
- Knowledge snippets (agentbus:{PROJECT_PREFIX}:jot:*)
- Tag indices (agentbus:{PROJECT_PREFIX}:jots_by_tag:*)
- Coordination flags (agentbus:{PROJECT_PREFIX}:flag:*)
- Agent read positions (agentbus:{PROJECT_PREFIX}:last:*)

⚠️  WARNING: This operation is destructive and cannot be undone!

Use --force to skip the confirmation prompt, useful in automation scenarios.

This command is useful for:
- Cleaning up test data
- Resetting development environments
- Starting fresh coordination sessions
- Troubleshooting Redis state issues

Example usage:
  agentbus clear                 # Interactive confirmation
  agentbus clear --force         # Skip confirmation
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"force",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Skip confirmation prompt"),
					parameters.WithDefault(false),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

// ClearStats holds statistics about what was cleared
type ClearStats struct {
	Messages  int
	Jots      int
	TagIndex  int
	Flags     int
	Positions int
	Total     int
}

func (c *ClearCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ClearSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	client, err := getRedisClient()
	if err != nil {
		return err
	}
	defer client.Close()

	// Get confirmation unless force flag is set
	if !s.Force {
		return errors.New("clearing all AgentBus data requires confirmation - use --force flag to proceed")
	}

	stats, err := clearAllData(ctx, client)
	if err != nil {
		return err
	}

	// Output results as structured data
	row := types.NewRow(
		types.MRP("messages_deleted", stats.Messages),
		types.MRP("jots_deleted", stats.Jots),
		types.MRP("tag_indices_deleted", stats.TagIndex),
		types.MRP("flags_deleted", stats.Flags),
		types.MRP("agent_positions_deleted", stats.Positions),
		types.MRP("total_keys_deleted", stats.Total),
	)

	return gp.AddRow(ctx, row)
}

func (c *ClearCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	s := &ClearSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	client, err := getRedisClient()
	if err != nil {
		return err
	}
	defer client.Close()

	// Show what will be deleted
	if !s.Force {
		fmt.Fprintf(w, "🗑️  AgentBus Clear Operation\n")
		fmt.Fprintf(w, "═══════════════════════════\n\n")

		// Show current data counts
		counts, err := getDataCounts(ctx, client)
		if err != nil {
			return err
		}

		fmt.Fprintf(w, "This will delete ALL AgentBus data:\n")
		fmt.Fprintf(w, "  📨 Messages: %d\n", counts.Messages)
		fmt.Fprintf(w, "  📝 Knowledge snippets (jots): %d\n", counts.Jots)
		fmt.Fprintf(w, "  🏷️  Tag indices: %d\n", counts.TagIndex)
		fmt.Fprintf(w, "  🚩 Coordination flags: %d\n", counts.Flags)
		fmt.Fprintf(w, "  📍 Agent read positions: %d\n", counts.Positions)
		fmt.Fprintf(w, "  ═══════════════════════════\n")
		fmt.Fprintf(w, "  Total keys: %d\n\n", counts.Total)

		if counts.Total == 0 {
			fmt.Fprintf(w, "✅ No AgentBus data found to clear.\n")
			return nil
		}

		fmt.Fprintf(w, "⚠️  WARNING: This operation cannot be undone!\n")
		fmt.Fprintf(w, "To proceed, run: agentbus clear --force\n")
		return nil
	}

	// Perform the clear operation
	fmt.Fprintf(w, "🗑️  Clearing all AgentBus data...\n")

	stats, err := clearAllData(ctx, client)
	if err != nil {
		return err
	}

	// Show results
	fmt.Fprintf(w, "\n✅ Clear operation completed successfully!\n")
	fmt.Fprintf(w, "═══════════════════════════════════════\n")
	fmt.Fprintf(w, "  📨 Messages deleted: %d\n", stats.Messages)
	fmt.Fprintf(w, "  📝 Jots deleted: %d\n", stats.Jots)
	fmt.Fprintf(w, "  🏷️  Tag indices deleted: %d\n", stats.TagIndex)
	fmt.Fprintf(w, "  🚩 Flags deleted: %d\n", stats.Flags)
	fmt.Fprintf(w, "  📍 Agent positions deleted: %d\n", stats.Positions)
	fmt.Fprintf(w, "  ═══════════════════════════\n")
	fmt.Fprintf(w, "  Total keys deleted: %d\n", stats.Total)

	return nil
}

// getDataCounts returns counts of each data type
func getDataCounts(ctx context.Context, client *agentredis.Client) (*ClearStats, error) {
	stats := &ClearStats{}

	// Count message streams
	messageKeys, err := client.Keys(ctx, client.Key("ch:*")).Result()
	if err != nil {
		log.Error().Err(err).Msg("Failed to scan message keys")
	} else {
		stats.Messages = len(messageKeys)
	}

	// Count jots
	jotKeys, err := client.Keys(ctx, client.Key("jot:*")).Result()
	if err != nil {
		log.Error().Err(err).Msg("Failed to scan jot keys")
	} else {
		stats.Jots = len(jotKeys)
	}

	// Count tag indices
	tagKeys, err := client.Keys(ctx, client.Key("jots_by_tag:*")).Result()
	if err != nil {
		log.Error().Err(err).Msg("Failed to scan tag index keys")
	} else {
		stats.TagIndex = len(tagKeys)
	}

	// Count flags
	flagKeys, err := client.Keys(ctx, client.Key("flag:*")).Result()
	if err != nil {
		log.Error().Err(err).Msg("Failed to scan flag keys")
	} else {
		stats.Flags = len(flagKeys)
	}

	// Count agent positions
	positionKeys, err := client.Keys(ctx, client.Key("last:*")).Result()
	if err != nil {
		log.Error().Err(err).Msg("Failed to scan position keys")
	} else {
		stats.Positions = len(positionKeys)
	}

	stats.Total = stats.Messages + stats.Jots + stats.TagIndex + stats.Flags + stats.Positions
	return stats, nil
}

// clearAllData performs the actual deletion and returns statistics
func clearAllData(ctx context.Context, client *agentredis.Client) (*ClearStats, error) {
	stats := &ClearStats{}

	// Clear message streams
	messageKeys, err := client.Keys(ctx, client.Key("ch:*")).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan message keys")
	}
	if len(messageKeys) > 0 {
		deleted, err := client.Del(ctx, messageKeys...).Result()
		if err != nil {
			return nil, errors.Wrap(err, "failed to delete message keys")
		}
		stats.Messages = int(deleted)
		log.Info().Int("count", int(deleted)).Msg("Deleted message streams")
	}

	// Clear jots
	jotKeys, err := client.Keys(ctx, client.Key("jot:*")).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan jot keys")
	}
	if len(jotKeys) > 0 {
		deleted, err := client.Del(ctx, jotKeys...).Result()
		if err != nil {
			return nil, errors.Wrap(err, "failed to delete jot keys")
		}
		stats.Jots = int(deleted)
		log.Info().Int("count", int(deleted)).Msg("Deleted jots")
	}

	// Clear tag indices
	tagKeys, err := client.Keys(ctx, client.Key("jots_by_tag:*")).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan tag index keys")
	}
	if len(tagKeys) > 0 {
		deleted, err := client.Del(ctx, tagKeys...).Result()
		if err != nil {
			return nil, errors.Wrap(err, "failed to delete tag index keys")
		}
		stats.TagIndex = int(deleted)
		log.Info().Int("count", int(deleted)).Msg("Deleted tag indices")
	}

	// Clear flags
	flagKeys, err := client.Keys(ctx, client.Key("flag:*")).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan flag keys")
	}
	if len(flagKeys) > 0 {
		deleted, err := client.Del(ctx, flagKeys...).Result()
		if err != nil {
			return nil, errors.Wrap(err, "failed to delete flag keys")
		}
		stats.Flags = int(deleted)
		log.Info().Int("count", int(deleted)).Msg("Deleted flags")
	}

	// Clear agent positions
	positionKeys, err := client.Keys(ctx, client.Key("last:*")).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan position keys")
	}
	if len(positionKeys) > 0 {
		deleted, err := client.Del(ctx, positionKeys...).Result()
		if err != nil {
			return nil, errors.Wrap(err, "failed to delete position keys")
		}
		stats.Positions = int(deleted)
		log.Info().Int("count", int(deleted)).Msg("Deleted agent positions")
	}

	stats.Total = stats.Messages + stats.Jots + stats.TagIndex + stats.Flags + stats.Positions

	log.Info().
		Int("messages", stats.Messages).
		Int("jots", stats.Jots).
		Int("tag_indices", stats.TagIndex).
		Int("flags", stats.Flags).
		Int("positions", stats.Positions).
		Int("total", stats.Total).
		Msg("AgentBus clear operation completed")

	return stats, nil
}
