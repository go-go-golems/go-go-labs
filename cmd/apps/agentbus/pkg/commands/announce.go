package commands

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type AnnounceCommand struct {
	*cmds.CommandDescription
}

type AnnounceSettings struct {
	Flag    string `glazed.parameter:"flag"`
	Force   bool   `glazed.parameter:"force"`
	Timeout int    `glazed.parameter:"timeout"`
}

var _ cmds.GlazeCommand = (*AnnounceCommand)(nil)
var _ cmds.WriterCommand = (*AnnounceCommand)(nil)

func NewAnnounceCommand() (*AnnounceCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	return &AnnounceCommand{
		CommandDescription: cmds.NewCommandDescription(
			"announce",
			cmds.WithShort("Declare that an agent is working on a task"),
			cmds.WithLong(`Declare that this agent is working on a specific task or owns a resource.

This sets a coordination flag that other agents can wait for using 'await'.
The flag includes the announcing agent's ID and timestamp, providing
clear ownership and timing information.

Use this for:
- Declaring work in progress ("building", "testing", "deploying")
- Claiming exclusive access to resources ("database-migration", "config-update")
- Signaling start of long-running tasks ("integration-test", "backup")
- Coordinating sequential operations ("step1", "step2", "step3")

By default, announce fails if the flag already exists to prevent conflicts.
Use --force to override existing flags.

Timeouts prevent stale flags from blocking other agents indefinitely:
- Set --timeout to automatically expire flags after specified seconds
- Use timeouts for temporary locks or time-bounded operations
- Choose timeout values longer than expected task duration
- When flags expire, they are automatically removed from Redis

Example usage in agent tool calling:
  agentbus announce building --timeout 1800
  agentbus announce --flag "database-migration" --force --timeout 3600
  agentbus announce deployment-lock --timeout 900
  agentbus announce integration-test-suite --timeout 2700`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"flag",
					parameters.ParameterTypeString,
					parameters.WithHelp("Name of the coordination flag to announce"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"force",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Override existing flag if present"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"timeout",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Automatically expire flag after this many seconds (0 = no timeout)"),
					parameters.WithDefault(0),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *AnnounceCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	startTime := time.Now()
	log.Debug().Msg("ANNOUNCE: Starting RunIntoGlazeProcessor")

	// Add timeout to context to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s := &AnnounceSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		log.Error().Err(err).Msg("ANNOUNCE: Failed to initialize announce settings")
		return err
	}

	log.Info().
		Str("flag", s.Flag).
		Bool("force", s.Force).
		Int("timeout", s.Timeout).
		Msg("ANNOUNCE: Starting announce operation")

	agentID, err := getAgentID()
	if err != nil {
		log.Error().Err(err).Msg("ANNOUNCE: Failed to get agent ID")
		return err
	}
	log.Debug().Str("agent_id", agentID).Str("flag", s.Flag).Msg("ANNOUNCE: Retrieved agent ID")

	log.Debug().Msg("ANNOUNCE: Creating Redis client")
	client, err := getRedisClient()
	if err != nil {
		log.Error().Err(err).Msg("ANNOUNCE: Failed to get Redis client")
		return err
	}
	defer func() {
		log.Debug().Msg("ANNOUNCE: Closing Redis client")
		client.Close()
	}()

	flagKey := client.FlagKey(s.Flag)
	now := time.Now()
	flagValue := fmt.Sprintf("%s @ %s", agentID, now.Format(time.RFC3339))
	log.Debug().Str("flag_key", flagKey).Str("flag_value", flagValue).Msg("ANNOUNCE: Prepared flag data")

	// Calculate expiration based on timeout
	var expiration time.Duration
	if s.Timeout > 0 {
		expiration = time.Duration(s.Timeout) * time.Second
		log.Debug().
			Int("timeout_seconds", s.Timeout).
			Str("flag", s.Flag).
			Dur("expiration", expiration).
			Msg("ANNOUNCE: Setting flag with timeout")
	} else {
		expiration = 0 // No expiration
		log.Debug().Str("flag", s.Flag).Msg("ANNOUNCE: Setting flag without timeout")
	}

	redisOpStart := time.Now()
	if s.Force {
		log.Debug().Str("flag", s.Flag).Str("agent_id", agentID).Msg("ANNOUNCE: Force setting flag (Redis SET)")
		// Force set the flag with optional expiration
		err = client.Set(ctx, flagKey, flagValue, expiration).Err()
		if err != nil {
			log.Error().Err(err).Str("flag", s.Flag).Dur("duration", time.Since(redisOpStart)).Msg("ANNOUNCE: Failed to force set flag")
			return errors.Wrap(err, "failed to set flag")
		}
		log.Debug().Str("flag", s.Flag).Str("agent_id", agentID).Dur("duration", time.Since(redisOpStart)).Msg("ANNOUNCE: Successfully force set flag")
	} else {
		log.Debug().Str("flag", s.Flag).Str("agent_id", agentID).Msg("ANNOUNCE: Setting flag if not exists (Redis SETNX)")
		// Set only if not exists (SETNX) with optional expiration
		success, err := client.SetNX(ctx, flagKey, flagValue, expiration).Result()
		if err != nil {
			log.Error().Err(err).Str("flag", s.Flag).Dur("duration", time.Since(redisOpStart)).Msg("ANNOUNCE: Failed to set flag")
			return errors.Wrap(err, "failed to set flag")
		}
		if !success {
			log.Debug().Str("flag", s.Flag).Msg("ANNOUNCE: Flag already exists, getting current value")
			// Flag already exists, get current value for error message
			currentValue, getErr := client.Get(ctx, flagKey).Result()
			if getErr != nil {
				log.Warn().Err(getErr).Str("flag", s.Flag).Msg("ANNOUNCE: Failed to get current flag value")
				currentValue = "<unknown>"
			}
			log.Warn().
				Str("flag", s.Flag).
				Str("current_value", currentValue).
				Dur("duration", time.Since(redisOpStart)).
				Msg("ANNOUNCE: Flag already exists and force not specified")
			return errors.Errorf("flag '%s' already exists (current: %s). Use --force to override", s.Flag, currentValue)
		}
		log.Debug().Str("flag", s.Flag).Str("agent_id", agentID).Dur("duration", time.Since(redisOpStart)).Msg("ANNOUNCE: Successfully set flag")
	}

	// Publish to communication channel (non-blocking)
	message := fmt.Sprintf("🚩 Announced working on '%s'", s.Flag)
	log.Debug().Str("message", message).Str("flag", s.Flag).Msg("ANNOUNCE: Publishing to communication channel")
	publishStart := time.Now()
	err = publishToChannel(ctx, client, agentID, message, "coordination")
	if err != nil {
		log.Warn().Err(err).Str("flag", s.Flag).Dur("duration", time.Since(publishStart)).Msg("ANNOUNCE: Failed to publish to communication channel")
	} else {
		log.Debug().Str("flag", s.Flag).Dur("duration", time.Since(publishStart)).Msg("ANNOUNCE: Successfully published to communication channel")
	}

	log.Info().
		Str("flag", s.Flag).
		Str("agent_id", agentID).
		Bool("forced", s.Force).
		Int("timeout", s.Timeout).
		Dur("total_duration", time.Since(startTime)).
		Msg("ANNOUNCE: Successfully announced flag")

	// Output the result
	row := types.NewRow(
		types.MRP("flag", s.Flag),
		types.MRP("agent_id", agentID),
		types.MRP("timestamp", now.Format(time.RFC3339)),
		types.MRP("forced", s.Force),
		types.MRP("timeout", s.Timeout),
	)

	log.Debug().Dur("total_duration", time.Since(startTime)).Msg("ANNOUNCE: Adding row to processor")
	return gp.AddRow(ctx, row)
}

func (c *AnnounceCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	startTime := time.Now()
	log.Debug().Msg("ANNOUNCE: Starting RunIntoWriter")

	// Add timeout to context to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s := &AnnounceSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		log.Error().Err(err).Msg("ANNOUNCE: Failed to initialize announce settings")
		return err
	}

	agentID, err := getAgentID()
	if err != nil {
		log.Error().Err(err).Msg("ANNOUNCE: Failed to get agent ID")
		return err
	}

	log.Info().
		Str("agent_id", agentID).
		Str("flag", s.Flag).
		Bool("force", s.Force).
		Int("timeout", s.Timeout).
		Msg("ANNOUNCE: Announcing coordination flag")

	log.Debug().Msg("ANNOUNCE: Creating Redis client")
	client, err := getRedisClient()
	if err != nil {
		log.Error().Err(err).Msg("ANNOUNCE: Failed to get Redis client")
		return err
	}
	defer func() {
		log.Debug().Msg("ANNOUNCE: Closing Redis client")
		client.Close()
	}()

	flagKey := client.FlagKey(s.Flag)
	now := time.Now()
	flagValue := fmt.Sprintf("%s @ %s", agentID, now.Format(time.RFC3339))
	log.Debug().Str("flag_key", flagKey).Str("flag_value", flagValue).Msg("ANNOUNCE: Prepared flag data")

	// Calculate expiration based on timeout
	var expiration time.Duration
	if s.Timeout > 0 {
		expiration = time.Duration(s.Timeout) * time.Second
		log.Debug().Int("timeout_seconds", s.Timeout).Dur("expiration", expiration).Msg("ANNOUNCE: Setting expiration")
	} else {
		log.Debug().Msg("ANNOUNCE: No expiration set")
	}

	// Check if flag already exists (non-force mode)
	if !s.Force {
		log.Debug().Str("flag", s.Flag).Msg("ANNOUNCE: Checking if flag already exists")
		checkStart := time.Now()
		existingValue, err := client.Get(ctx, flagKey).Result()
		if err == nil && existingValue != "" {
			log.Debug().Str("flag", s.Flag).Str("existing_value", existingValue).
				Dur("duration", time.Since(checkStart)).
				Msg("ANNOUNCE: Flag already exists, not overriding")
			fmt.Fprintf(w, "🚩 Flag '%s' already announced by %s (use --force to override)\n", s.Flag, existingValue)
			return nil
		}
		log.Debug().Str("flag", s.Flag).Dur("duration", time.Since(checkStart)).Msg("ANNOUNCE: Flag check completed")
	}

	// Set the flag with agent ID, timestamp, and optional expiration
	redisOpStart := time.Now()
	if s.Force {
		log.Debug().Str("flag", s.Flag).Msg("ANNOUNCE: Force setting flag (Redis SET)")
		err = client.Set(ctx, flagKey, flagValue, expiration).Err()
	} else {
		log.Debug().Str("flag", s.Flag).Msg("ANNOUNCE: Setting flag if not exists (Redis SETNX)")
		success, err := client.SetNX(ctx, flagKey, flagValue, expiration).Result()
		if err == nil && !success {
			log.Debug().Str("flag", s.Flag).Msg("ANNOUNCE: SETNX failed, getting existing value")
			existingValue, _ := client.Get(ctx, flagKey).Result()
			log.Warn().Str("flag", s.Flag).Str("existing_value", existingValue).
				Dur("duration", time.Since(redisOpStart)).
				Msg("ANNOUNCE: Flag was set by another process concurrently")
			fmt.Fprintf(w, "🚩 Flag '%s' already announced by %s (use --force to override)\n", s.Flag, existingValue)
			return nil
		}
	}
	if err != nil {
		log.Error().Err(err).Str("flag", s.Flag).Dur("duration", time.Since(redisOpStart)).Msg("ANNOUNCE: Failed to announce flag")
		return errors.Wrap(err, "failed to announce flag")
	}
	log.Debug().Str("flag", s.Flag).Dur("duration", time.Since(redisOpStart)).Msg("ANNOUNCE: Successfully set flag")

	// Output success message
	timestamp := now.Format("15:04:05")
	if s.Timeout > 0 {
		timeoutInfo := fmt.Sprintf(" (expires in %ds)", s.Timeout)
		if s.Force {
			fmt.Fprintf(w, "🚩 [%s] Flag '%s' forcefully announced by %s%s\n", timestamp, s.Flag, agentID, timeoutInfo)
		} else {
			fmt.Fprintf(w, "🚩 [%s] Flag '%s' announced by %s%s\n", timestamp, s.Flag, agentID, timeoutInfo)
		}
	} else {
		if s.Force {
			fmt.Fprintf(w, "🚩 [%s] Flag '%s' forcefully announced by %s\n", timestamp, s.Flag, agentID)
		} else {
			fmt.Fprintf(w, "🚩 [%s] Flag '%s' announced by %s\n", timestamp, s.Flag, agentID)
		}
	}

	// Show latest messages after announcing flag
	log.Debug().Msg("ANNOUNCE: Showing latest messages")
	messageStart := time.Now()
	err = showLatestMessages(ctx, client, w, agentID, 3)
	if err != nil {
		log.Warn().Err(err).Dur("duration", time.Since(messageStart)).Msg("ANNOUNCE: Failed to show latest messages")
	} else {
		log.Debug().Dur("duration", time.Since(messageStart)).Msg("ANNOUNCE: Successfully showed latest messages")
	}

	log.Debug().Dur("total_duration", time.Since(startTime)).Msg("ANNOUNCE: Completed RunIntoWriter")
	return nil
}
