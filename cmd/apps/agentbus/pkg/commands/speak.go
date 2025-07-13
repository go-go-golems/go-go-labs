package commands

import (
	"context"
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

type SpeakCommand struct {
	*cmds.CommandDescription
}

type SpeakSettings struct {
	Message string `glazed.parameter:"msg"`
	Topic   string `glazed.parameter:"topic"`
}

var _ cmds.GlazeCommand = (*SpeakCommand)(nil)

func NewSpeakCommand() (*SpeakCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	return &SpeakCommand{
		CommandDescription: cmds.NewCommandDescription(
			"speak",
			cmds.WithShort("Send a message to the shared agent communication channel"),
			cmds.WithLong(`Send a message to the shared Redis-backed communication stream for agent coordination.

All agents share a single communication channel. Messages can optionally include
a topic slug for categorization and filtering.

The message is added to a Redis Stream with the sender's agent ID, timestamp,
and optional topic. Other agents can receive these messages using the 'overhear' command.

This is ideal for:
- Broadcasting status updates ("Build completed successfully") 
- Sharing progress information ("Processing 50% complete")
- Coordinating with other agents ("Starting deployment phase")
- Announcing completion of tasks ("Unit tests passed ✅")

Example usage in agent tool calling:
  agentbus speak --msg "Compilation finished, running tests" --topic "build"
  agentbus speak --msg "Production deployment started" --topic "deploy"
  agentbus speak --msg "All services healthy" --topic "status"`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"msg",
					parameters.ParameterTypeString,
					parameters.WithHelp("Message content to send"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"topic",
					parameters.ParameterTypeString,
					parameters.WithHelp("Optional topic slug for message categorization"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *SpeakCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &SpeakSettings{}
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

	// Add message to Redis Stream
	streamKey := client.ChannelKey()
	values := map[string]interface{}{
		"agent_id": agentID,
		"message":  s.Message,
	}
	if s.Topic != "" {
		values["topic"] = s.Topic
	}
	
	result, err := client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: values,
	}).Result()
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	// Output the result
	row := types.NewRow(
		types.MRP("stream_id", result),
		types.MRP("agent_id", agentID),
		types.MRP("topic", s.Topic),
		types.MRP("message", s.Message),
		types.MRP("timestamp", time.Now().Format(time.RFC3339)),
	)

	return gp.AddRow(ctx, row)
}
