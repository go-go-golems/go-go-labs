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

type SatisfyCommand struct {
	*cmds.CommandDescription
}

type SatisfySettings struct {
	Flag string `glazed.parameter:"flag"`
}

var _ cmds.GlazeCommand = (*SatisfyCommand)(nil)

func NewSatisfyCommand() (*SatisfyCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	return &SatisfyCommand{
		CommandDescription: cmds.NewCommandDescription(
			"satisfy",
			cmds.WithShort("Mark a coordination flag as completed"),
			cmds.WithLong(`Mark a coordination flag as completed by deleting it from Redis.

This signals to any agents waiting with 'await' that the task or dependency
has been satisfied. It's the final step in the coordination cycle.

Use this after:
- Completing a build process
- Finishing test suites  
- Deploying applications
- Releasing resource locks
- Completing any announced task

Good practice is to always call satisfy after completing work that was
announced, to prevent other agents from waiting indefinitely.

The command returns information about the flag that was satisfied,
including details about who originally announced it.

Example usage in agent tool calling:
  agentbus satisfy building
  agentbus satisfy integration-tests
  agentbus satisfy deployment-lock
  agentbus satisfy database-migration`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"flag",
					parameters.ParameterTypeString,
					parameters.WithHelp("Name of the coordination flag to satisfy"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *SatisfyCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &SatisfySettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	client, err := getRedisClient()
	if err != nil {
		return err
	}
	defer client.Close()

	flagKey := client.FlagKey(s.Flag)

	// Get flag value before deleting (for output)
	flagValue, err := client.Get(ctx, flagKey).Result()
	if err == redis.Nil {
		return errors.Errorf("flag '%s' does not exist", s.Flag)
	}
	if err != nil {
		return errors.Wrap(err, "failed to get flag value")
	}

	// Delete the flag
	deleted, err := client.Del(ctx, flagKey).Result()
	if err != nil {
		return errors.Wrap(err, "failed to delete flag")
	}

	if deleted == 0 {
		return errors.Errorf("flag '%s' was not deleted (may have been removed by another agent)", s.Flag)
	}

	// Parse flag value for output
	parts := strings.SplitN(flagValue, " @ ", 2)
	var announcedBy, announcedAt string
	if len(parts) == 2 {
		announcedBy = parts[0]
		announcedAt = parts[1]
	} else {
		announcedBy = flagValue
		announcedAt = "unknown"
	}

	// Publish to communication channel (non-blocking)
	agentID, _ := getAgentID()
	message := fmt.Sprintf("✅ Satisfied '%s'", s.Flag)
	_ = publishToChannel(ctx, client, agentID, message, "coordination")

	// Output the result
	row := types.NewRow(
		types.MRP("flag", s.Flag),
		types.MRP("announced_by", announcedBy),
		types.MRP("announced_at", announcedAt),
		types.MRP("satisfied_at", time.Now().Format(time.RFC3339)),
	)

	return gp.AddRow(ctx, row)
}
