package directions

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"googlemaps.github.io/maps"
)

type RootCommandInterface interface {
	GetMapsClient(ctx context.Context) (*maps.Client, error)
}

type DirectionsSettings struct {
	Origin      string   `glazed.parameter:"origin"`
	Destination string   `glazed.parameter:"destination"`
	Mode        string   `glazed.parameter:"mode"`
	Waypoints   []string `glazed.parameter:"waypoints"`
	Avoid       []string `glazed.parameter:"avoid"`
	Units       string   `glazed.parameter:"units"`
}

type DirectionsCommand struct {
	*cmds.CommandDescription
	settings DirectionsSettings
}

func NewDirectionsCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	cmd := &DirectionsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"directions",
			cmds.WithShort("Get directions between locations"),
			cmds.WithLong("Get detailed directions between two locations with optional waypoints and preferences"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"origin",
					parameters.ParameterTypeString,
					parameters.WithHelp("Starting location (address or lat,lng)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"destination",
					parameters.ParameterTypeString,
					parameters.WithHelp("Ending location (address or lat,lng)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"mode",
					parameters.ParameterTypeChoice,
					parameters.WithHelp("Travel mode"),
					parameters.WithDefault("driving"),
					parameters.WithChoices("driving", "walking", "bicycling", "transit"),
				),
				parameters.NewParameterDefinition(
					"waypoints",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of waypoints to include in the route"),
				),
				parameters.NewParameterDefinition(
					"avoid",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Features to avoid (tolls, highways, ferries)"),
				),
				parameters.NewParameterDefinition(
					"units",
					parameters.ParameterTypeChoice,
					parameters.WithHelp("Unit system for distances"),
					parameters.WithDefault("metric"),
					parameters.WithChoices("metric", "imperial"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return cli.BuildCobraCommandFromGlazeCommand(cmd)
}

func (c *DirectionsCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, &c.settings); err != nil {
		return err
	}

	// Log query parameters
	log.Debug().
		Str("origin", c.settings.Origin).
		Str("destination", c.settings.Destination).
		Str("mode", c.settings.Mode).
		Strs("waypoints", c.settings.Waypoints).
		Strs("avoid", c.settings.Avoid).
		Str("units", c.settings.Units).
		Msg("Getting directions")

	// Get the root command to access the Maps client
	rootCmd, ok := ctx.Value("rootCmd").(RootCommandInterface)
	if !ok {
		return fmt.Errorf("root command not found in context")
	}
	client, err := rootCmd.GetMapsClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Maps client: %w", err)
	}

	// Prepare the directions request
	req := &maps.DirectionsRequest{
		Origin:      c.settings.Origin,
		Destination: c.settings.Destination,
		Mode:        maps.Mode(strings.ToUpper(c.settings.Mode)),
		Units:       maps.Units(strings.ToUpper(c.settings.Units)),
	}

	// Add waypoints if provided
	if len(c.settings.Waypoints) > 0 {
		req.Waypoints = c.settings.Waypoints
	}

	// Add avoid preferences if provided
	for _, avoid := range c.settings.Avoid {
		switch strings.ToLower(avoid) {
		case "tolls":
			req.Avoid = append(req.Avoid, maps.AvoidTolls)
		case "highways":
			req.Avoid = append(req.Avoid, maps.AvoidHighways)
		case "ferries":
			req.Avoid = append(req.Avoid, maps.AvoidFerries)
		}
	}

	// Execute the directions request
	resp, _, err := client.Directions(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get directions: %w", err)
	}

	if len(resp) == 0 {
		return fmt.Errorf("no routes found")
	}

	// Output each route as structured data
	for routeIndex, route := range resp {
		// Output route summary
		summaryRow := types.NewRow(
			types.MRP("route_number", routeIndex+1),
			types.MRP("summary", route.Summary),
			types.MRP("distance", fmt.Sprintf("%.1f km", float64(route.Legs[0].Distance.Meters)/1000)),
			types.MRP("duration", route.Legs[0].Duration.String()),
		)
		if err := gp.AddRow(ctx, summaryRow); err != nil {
			return err
		}

		// Output each step in the route
		for stepIndex, step := range route.Legs[0].Steps {
			stepRow := types.NewRow(
				types.MRP("route_number", routeIndex+1),
				types.MRP("step_number", stepIndex+1),
				types.MRP("instruction", step.HTMLInstructions),
				types.MRP("distance", fmt.Sprintf("%.1f km", float64(step.Distance.Meters)/1000)),
				types.MRP("duration", step.Duration.String()),
				types.MRP("start_location", fmt.Sprintf("%f,%f", step.StartLocation.Lat, step.StartLocation.Lng)),
				types.MRP("end_location", fmt.Sprintf("%f,%f", step.EndLocation.Lat, step.EndLocation.Lng)),
				types.MRP("travel_mode", string(step.TravelMode)),
			)
			if err := gp.AddRow(ctx, stepRow); err != nil {
				return err
			}
		}
	}

	// Log results
	log.Debug().
		Int("routes", len(resp)).
		Str("origin", c.settings.Origin).
		Str("destination", c.settings.Destination).
		Msg("Directions retrieved")

	return nil
}
