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

type AwaitCommand struct {
	*cmds.CommandDescription
}

type AwaitSettings struct {
	Flag    string `glazed.parameter:"flag"`
	Timeout int    `glazed.parameter:"timeout"`
	Delete  bool   `glazed.parameter:"delete"`
}

var _ cmds.GlazeCommand = (*AwaitCommand)(nil)

func NewAwaitCommand() (*AwaitCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	return &AwaitCommand{
		CommandDescription: cmds.NewCommandDescription(
			"await",
			cmds.WithShort("Wait for a coordination flag to be satisfied"),
			cmds.WithLong(`Wait for a coordination flag to exist, indicating another agent has completed a task.

This command polls the Redis flag key and blocks until the flag is set by
another agent using 'announce'. It's perfect for creating dependencies
between agent tasks and ensuring proper ordering.

The command returns information about the flag when it becomes available,
including which agent set it and when.

Use this for:
- Waiting for builds to complete before deploying
- Ensuring tests pass before merging
- Coordinating sequential deployment steps
- Waiting for resource locks to be released
- Synchronizing multi-agent workflows

Example usage in agent tool calling:
  agentbus await building --timeout 900
  agentbus await integration-tests --timeout 1800 --delete
  agentbus await deployment-ready
  agentbus await database-migration-complete --timeout 3600`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"flag",
					parameters.ParameterTypeString,
					parameters.WithHelp("Name of the coordination flag to wait for"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"timeout",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Maximum time to wait in seconds (0 = no timeout)"),
					parameters.WithDefault(0),
				),
				parameters.NewParameterDefinition(
					"delete",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Delete the flag after detecting it"),
					parameters.WithDefault(false),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *AwaitCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &AwaitSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize await settings")
		return err
	}

	log.Info().
		Str("flag", s.Flag).
		Int("timeout", s.Timeout).
		Bool("delete", s.Delete).
		Msg("Starting await operation")

	client, err := getRedisClient()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Redis client")
		return err
	}
	defer client.Close()

	flagKey := client.FlagKey(s.Flag)

	// Set up timeout context if specified
	waitCtx := ctx
	if s.Timeout > 0 {
		log.Debug().
			Str("flag", s.Flag).
			Int("timeout_seconds", s.Timeout).
			Msg("Setting up timeout context")
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, time.Duration(s.Timeout)*time.Second)
		defer cancel()
	}

	// Poll for the flag
	start := time.Now()
	log.Debug().Str("flag", s.Flag).Msg("Starting to poll for flag")
	ticker := time.NewTicker(100 * time.Millisecond) // Poll every 100ms
	defer ticker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			if s.Timeout > 0 && waitCtx.Err() == context.DeadlineExceeded {
				log.Warn().
					Str("flag", s.Flag).
					Int("timeout_seconds", s.Timeout).
					Msg("Timeout waiting for flag")
				return errors.Errorf("timeout waiting for flag '%s' after %d seconds", s.Flag, s.Timeout)
			}
			log.Debug().Err(waitCtx.Err()).Str("flag", s.Flag).Msg("Context cancelled while waiting")
			return waitCtx.Err()
		case <-ticker.C:
			// Check if flag exists
			flagValue, err := client.Get(waitCtx, flagKey).Result()
			if err == redis.Nil {
				// Flag doesn't exist yet, continue waiting
				log.Debug().Str("flag", s.Flag).Msg("Flag not found, continuing to wait")
				continue
			}
			if err != nil {
				log.Error().Err(err).Str("flag", s.Flag).Msg("Failed to check flag")
				return errors.Wrap(err, "failed to check flag")
			}

			// Flag exists! Parse the value
			log.Debug().Str("flag", s.Flag).Str("flag_value", flagValue).Msg("Flag found, parsing value")
			parts := strings.SplitN(flagValue, " @ ", 2)
			var agentID, timestamp string
			if len(parts) == 2 {
				agentID = parts[0]
				timestamp = parts[1]
			} else {
				agentID = flagValue
				timestamp = "unknown"
			}

			// Delete flag if requested
			if s.Delete {
				log.Debug().Str("flag", s.Flag).Msg("Deleting flag as requested")
				err = client.Del(waitCtx, flagKey).Err()
				if err != nil {
					log.Error().Err(err).Str("flag", s.Flag).Msg("Failed to delete flag")
					return errors.Wrap(err, "failed to delete flag")
				}
				log.Debug().Str("flag", s.Flag).Msg("Successfully deleted flag")
			}

			waitDuration := time.Since(start)

			// Publish to communication channel (non-blocking)
			currentAgentID, _ := getAgentID()
			message := fmt.Sprintf("⏳ Waiting for '%s' completed", s.Flag)
			log.Debug().Str("message", message).Str("flag", s.Flag).Msg("Publishing completion to communication channel")
			err = publishToChannel(waitCtx, client, currentAgentID, message, "coordination")
			if err != nil {
				log.Warn().Err(err).Str("flag", s.Flag).Msg("Failed to publish to communication channel")
			}

			log.Info().
				Str("flag", s.Flag).
				Str("satisfied_by", agentID).
				Int64("wait_duration_ms", waitDuration.Milliseconds()).
				Bool("deleted", s.Delete).
				Msg("Successfully awaited flag")

			// Output the result
			row := types.NewRow(
				types.MRP("flag", s.Flag),
				types.MRP("satisfied_by", agentID),
				types.MRP("satisfied_at", timestamp),
				types.MRP("wait_duration_ms", waitDuration.Milliseconds()),
				types.MRP("deleted", s.Delete),
			)

			return gp.AddRow(waitCtx, row)
		}
	}
}
