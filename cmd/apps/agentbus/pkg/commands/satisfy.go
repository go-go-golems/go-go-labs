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
		log.Error().Err(err).Msg("Failed to initialize satisfy settings")
		return err
	}

	log.Info().
		Str("flag", s.Flag).
		Msg("Starting satisfy operation")

	client, err := getRedisClient()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Redis client")
		return err
	}
	defer client.Close()

	flagKey := client.FlagKey(s.Flag)

	log.Debug().Str("flag", s.Flag).Msg("Getting flag value before deleting")

	// Get flag value before deleting (for output)
	flagValue, err := client.Get(ctx, flagKey).Result()
	if err == redis.Nil {
		log.Warn().Str("flag", s.Flag).Msg("Flag does not exist")
		return errors.Errorf("flag '%s' does not exist", s.Flag)
	}
	if err != nil {
		log.Error().Err(err).Str("flag", s.Flag).Msg("Failed to get flag value")
		return errors.Wrap(err, "failed to get flag value")
	}

	log.Debug().Str("flag", s.Flag).Str("flag_value", flagValue).Msg("Retrieved flag value, proceeding to delete")

	// Delete the flag
	deleted, err := client.Del(ctx, flagKey).Result()
	if err != nil {
		log.Error().Err(err).Str("flag", s.Flag).Msg("Failed to delete flag")
		return errors.Wrap(err, "failed to delete flag")
	}

	if deleted == 0 {
		log.Warn().Str("flag", s.Flag).Msg("Flag was not deleted (may have been removed by another agent)")
		return errors.Errorf("flag '%s' was not deleted (may have been removed by another agent)", s.Flag)
	}

	log.Debug().Str("flag", s.Flag).Msg("Successfully deleted flag")

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

	log.Debug().
		Str("flag", s.Flag).
		Str("announced_by", announcedBy).
		Str("announced_at", announcedAt).
		Msg("Parsed flag details")

	// Publish to communication channel (non-blocking)
	agentID, _ := getAgentID()
	message := fmt.Sprintf("✅ Satisfied '%s'", s.Flag)
	log.Debug().Str("message", message).Str("flag", s.Flag).Msg("Publishing to communication channel")
	err = publishToChannel(ctx, client, agentID, message, "coordination")
	if err != nil {
		log.Warn().Err(err).Str("flag", s.Flag).Msg("Failed to publish to communication channel")
	}

	log.Info().
		Str("flag", s.Flag).
		Str("announced_by", announcedBy).
		Str("announced_at", announcedAt).
		Msg("Successfully satisfied flag")

	// Output the result
	row := types.NewRow(
		types.MRP("flag", s.Flag),
		types.MRP("announced_by", announcedBy),
		types.MRP("announced_at", announcedAt),
		types.MRP("satisfied_at", time.Now().Format(time.RFC3339)),
	)

	return gp.AddRow(ctx, row)
}
