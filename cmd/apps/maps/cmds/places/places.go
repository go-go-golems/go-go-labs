package places

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"googlemaps.github.io/maps"
)

type RootCommandInterface interface {
	GetMapsClient(ctx context.Context) (*maps.Client, error)
}

type SearchCommand struct {
	*cmds.CommandDescription
	query    string
	location string
	radius   int
	typeOf   string
}

func NewSearchCommand() (*cobra.Command, error) {
	cmd := &SearchCommand{
		CommandDescription: cmds.NewCommandDescription(
			"search",
			cmds.WithShort("Search for places"),
			cmds.WithLong("Search for places using text queries, location, and other filters"),
		),
	}

	return cli.BuildCobraCommandFromGlazeCommand(cmd)
}

func (c *SearchCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	// Log query parameters
	log.Debug().
		Str("query", c.query).
		Str("location", c.location).
		Int("radius", c.radius).
		Str("type", c.typeOf).
		Msg("Executing place search")

	// Get the root command to access the Maps client
	rootCmd, ok := ctx.Value("rootCmd").(RootCommandInterface)
	if !ok {
		return fmt.Errorf("root command not found in context")
	}
	client, err := rootCmd.GetMapsClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Maps client: %w", err)
	}

	// Prepare the search request
	req := &maps.TextSearchRequest{
		Query:  c.query,
		Radius: uint(c.radius),
	}

	// Add type if provided
	if c.typeOf != "" {
		req.Type = maps.PlaceType(c.typeOf)
	}

	// Add location if provided
	if c.location != "" {
		parts := strings.Split(c.location, ",")
		if len(parts) != 2 {
			return fmt.Errorf("invalid location format, expected lat,lng but got: %s", c.location)
		}
		lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return fmt.Errorf("invalid latitude: %w", err)
		}
		lng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return fmt.Errorf("invalid longitude: %w", err)
		}
		req.Location = &maps.LatLng{
			Lat: lat,
			Lng: lng,
		}
	}

	// Execute the search
	resp, err := client.TextSearch(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute text search: %w", err)
	}

	// Output results as structured data
	for _, place := range resp.Results {
		row := types.NewRow(
			types.MRP("name", place.Name),
			types.MRP("address", place.FormattedAddress),
			types.MRP("place_id", place.PlaceID),
			types.MRP("rating", place.Rating),
			types.MRP("user_ratings_total", place.UserRatingsTotal),
			types.MRP("types", strings.Join(place.Types, ", ")),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	// Log results
	log.Debug().
		Int("results", len(resp.Results)).
		Str("query", c.query).
		Msg("Place search completed")

	return nil
}

type DetailsCommand struct {
	*cmds.CommandDescription
	placeID string
}

func NewDetailsCommand() (*cobra.Command, error) {
	cmd := &DetailsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"details",
			cmds.WithShort("Get place details"),
			cmds.WithLong("Get detailed information about a specific place using its Place ID"),
		),
	}

	return cli.BuildCobraCommandFromGlazeCommand(cmd)
}

func (c *DetailsCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	// Log query parameters
	log.Debug().
		Str("placeID", c.placeID).
		Msg("Fetching place details")

	// Get the root command to access the Maps client
	rootCmd, ok := ctx.Value("rootCmd").(RootCommandInterface)
	if !ok {
		return fmt.Errorf("root command not found in context")
	}
	client, err := rootCmd.GetMapsClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Maps client: %w", err)
	}

	// Execute the details request
	resp, err := client.PlaceDetails(ctx, &maps.PlaceDetailsRequest{
		PlaceID: c.placeID,
	})
	if err != nil {
		return fmt.Errorf("failed to get place details: %w", err)
	}

	// Create opening hours string if available
	var openingHours []string
	if resp.OpeningHours != nil {
		for _, period := range resp.OpeningHours.Periods {
			if period.Open.Time != "" && period.Close.Time != "" {
				openingHours = append(openingHours,
					fmt.Sprintf("%s: %s - %s",
						period.Open.Day,
						period.Open.Time,
						period.Close.Time))
			}
		}
	}

	// Output results as structured data
	row := types.NewRow(
		types.MRP("name", resp.Name),
		types.MRP("address", resp.FormattedAddress),
		types.MRP("place_id", resp.PlaceID),
		types.MRP("rating", resp.Rating),
		types.MRP("user_ratings_total", resp.UserRatingsTotal),
		types.MRP("types", strings.Join(resp.Types, ", ")),
		types.MRP("phone", resp.InternationalPhoneNumber),
		types.MRP("website", resp.Website),
		types.MRP("opening_hours", strings.Join(openingHours, "\n")),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return err
	}

	// Log results
	log.Debug().
		Str("placeID", c.placeID).
		Str("name", resp.Name).
		Msg("Place details retrieved")

	return nil
}

type NearbyCommand struct {
	*cmds.CommandDescription
	location string
	radius   int
	typeOf   string
	keyword  string
}

func NewNearbyCommand() (*cobra.Command, error) {
	cmd := &NearbyCommand{
		CommandDescription: cmds.NewCommandDescription(
			"nearby",
			cmds.WithShort("Search for nearby places"),
			cmds.WithLong("Search for places near a specific location using various filters"),
		),
	}

	return cli.BuildCobraCommandFromGlazeCommand(cmd)
}

func (c *NearbyCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	// Log query parameters
	log.Debug().
		Str("location", c.location).
		Int("radius", c.radius).
		Str("type", c.typeOf).
		Str("keyword", c.keyword).
		Msg("Searching for nearby places")

	// Get the root command to access the Maps client
	rootCmd, ok := ctx.Value("rootCmd").(RootCommandInterface)
	if !ok {
		return fmt.Errorf("root command not found in context")
	}
	client, err := rootCmd.GetMapsClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Maps client: %w", err)
	}

	// Parse location
	parts := strings.Split(c.location, ",")
	if len(parts) != 2 {
		return fmt.Errorf("invalid location format, expected lat,lng but got: %s", c.location)
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return fmt.Errorf("invalid latitude: %w", err)
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return fmt.Errorf("invalid longitude: %w", err)
	}

	// Prepare the nearby search request
	req := &maps.NearbySearchRequest{
		Location: &maps.LatLng{
			Lat: lat,
			Lng: lng,
		},
		Radius:  uint(c.radius),
		Keyword: c.keyword,
	}

	// Add type if provided
	if c.typeOf != "" {
		req.Type = maps.PlaceType(c.typeOf)
	}

	// Execute the nearby search
	resp, err := client.NearbySearch(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute nearby search: %w", err)
	}

	// Output results as structured data
	for _, place := range resp.Results {
		row := types.NewRow(
			types.MRP("name", place.Name),
			types.MRP("address", place.Vicinity),
			types.MRP("place_id", place.PlaceID),
			types.MRP("rating", place.Rating),
			types.MRP("user_ratings_total", place.UserRatingsTotal),
			types.MRP("types", strings.Join(place.Types, ", ")),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	// Log results
	log.Debug().
		Int("results", len(resp.Results)).
		Str("location", c.location).
		Msg("Nearby search completed")

	return nil
}

func NewPlacesCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "places",
		Short: "Interact with Google Places API",
		Long:  "Search for places, get place details, and find nearby locations using the Google Places API",
	}

	searchCmd, err := NewSearchCommand()
	if err != nil {
		return nil, err
	}

	detailsCmd, err := NewDetailsCommand()
	if err != nil {
		return nil, err
	}

	nearbyCmd, err := NewNearbyCommand()
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(searchCmd, detailsCmd, nearbyCmd)
	return cmd, nil
}
