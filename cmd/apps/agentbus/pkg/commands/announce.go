package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
)

type AnnounceCommand struct {
	*cmds.CommandDescription
}

type AnnounceSettings struct {
	Flag  string `glazed.parameter:"flag"`
	Force bool   `glazed.parameter:"force"`
}

var _ cmds.GlazeCommand = (*AnnounceCommand)(nil)

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

Example usage in agent tool calling:
  agentbus announce building
  agentbus announce --flag "database-migration" --force
  agentbus announce deployment-lock
  agentbus announce integration-test-suite`),
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
	s := &AnnounceSettings{}
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

	flagKey := client.FlagKey(s.Flag)
	now := time.Now()
	flagValue := fmt.Sprintf("%s @ %s", agentID, now.Format(time.RFC3339))

	if s.Force {
		// Force set the flag
		err = client.Set(ctx, flagKey, flagValue, 0).Err()
		if err != nil {
			return errors.Wrap(err, "failed to set flag")
		}
	} else {
		// Set only if not exists (SETNX)
		success, err := client.SetNX(ctx, flagKey, flagValue, 0).Result()
		if err != nil {
			return errors.Wrap(err, "failed to set flag")
		}
		if !success {
			// Flag already exists, get current value for error message
			currentValue, _ := client.Get(ctx, flagKey).Result()
			return errors.Errorf("flag '%s' already exists (current: %s). Use --force to override", s.Flag, currentValue)
		}
	}

	// Publish to communication channel (non-blocking)
	message := fmt.Sprintf("🚩 Announced working on '%s'", s.Flag)
	_ = publishToChannel(ctx, client, agentID, message, "coordination")

	// Output the result
	row := types.NewRow(
		types.MRP("flag", s.Flag),
		types.MRP("agent_id", agentID),
		types.MRP("timestamp", now.Format(time.RFC3339)),
		types.MRP("forced", s.Force),
	)

	return gp.AddRow(ctx, row)
}
