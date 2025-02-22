package places

import (
	"context"
	"fmt"
	"strconv"
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

type SearchSettings struct {
	Query    string `glazed.parameter:"query"`
	Location string `glazed.parameter:"location"`
	Radius   int    `glazed.parameter:"radius"`
	Type     string `glazed.parameter:"type"`
}

type SearchCommand struct {
	*cmds.CommandDescription
	settings SearchSettings
}

func NewSearchCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	cmd := &SearchCommand{
		CommandDescription: cmds.NewCommandDescription(
			"search",
			cmds.WithShort("Search for places"),
			cmds.WithLong("Search for places using text queries, location, and other filters"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"query",
					parameters.ParameterTypeString,
					parameters.WithHelp("Text to search for"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"location",
					parameters.ParameterTypeString,
					parameters.WithHelp("Location in lat,lng format (e.g. 40.7128,-74.0060)"),
				),
				parameters.NewParameterDefinition(
					"radius",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Search radius in meters"),
					parameters.WithDefault(1500),
				),
				parameters.NewParameterDefinition(
					"type",
					parameters.ParameterTypeChoice,
					parameters.WithHelp("Type of place (e.g. restaurant, museum)"),
					parameters.WithChoices(
						"accounting", "airport", "amusement_park", "aquarium", "art_gallery",
						"atm", "bakery", "bank", "bar", "beauty_salon", "bicycle_store",
						"book_store", "bowling_alley", "bus_station", "cafe", "campground",
						"car_dealer", "car_rental", "car_repair", "car_wash", "casino",
						"cemetery", "church", "city_hall", "clothing_store", "convenience_store",
						"courthouse", "dentist", "department_store", "doctor", "drugstore",
						"electrician", "electronics_store", "embassy", "fire_station", "florist",
						"funeral_home", "furniture_store", "gas_station", "gym", "hair_care",
						"hardware_store", "hindu_temple", "home_goods_store", "hospital",
						"insurance_agency", "jewelry_store", "laundry", "lawyer", "library",
						"light_rail_station", "liquor_store", "local_government_office",
						"locksmith", "lodging", "meal_delivery", "meal_takeaway", "mosque",
						"movie_rental", "movie_theater", "moving_company", "museum", "night_club",
						"painter", "park", "parking", "pet_store", "pharmacy", "physiotherapist",
						"plumber", "police", "post_office", "primary_school", "real_estate_agency",
						"restaurant", "roofing_contractor", "rv_park", "school", "secondary_school",
						"shoe_store", "shopping_mall", "spa", "stadium", "storage", "store",
						"subway_station", "supermarket", "synagogue", "taxi_stand",
						"tourist_attraction", "train_station", "transit_station", "travel_agency",
						"university", "veterinary_care", "zoo",
					),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return cli.BuildCobraCommandFromGlazeCommand(cmd)
}

func (c *SearchCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, &c.settings); err != nil {
		return err
	}

	// Log query parameters
	log.Debug().
		Str("query", c.settings.Query).
		Str("location", c.settings.Location).
		Int("radius", c.settings.Radius).
		Str("type", c.settings.Type).
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
		Query:  c.settings.Query,
		Radius: uint(c.settings.Radius),
	}

	// Add type if provided
	if c.settings.Type != "" {
		req.Type = maps.PlaceType(c.settings.Type)
	}

	// Add location if provided
	if c.settings.Location != "" {
		parts := strings.Split(c.settings.Location, ",")
		if len(parts) != 2 {
			return fmt.Errorf("invalid location format, expected lat,lng but got: %s", c.settings.Location)
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
		Str("query", c.settings.Query).
		Msg("Place search completed")

	return nil
}

type DetailsSettings struct {
	PlaceID string `glazed.parameter:"place-id"`
}

type DetailsCommand struct {
	*cmds.CommandDescription
	settings DetailsSettings
}

func NewDetailsCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	cmd := &DetailsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"details",
			cmds.WithShort("Get place details"),
			cmds.WithLong("Get detailed information about a specific place using its Place ID"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"place-id",
					parameters.ParameterTypeString,
					parameters.WithHelp("Place ID to get details for"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return cli.BuildCobraCommandFromGlazeCommand(cmd)
}

func (c *DetailsCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, &c.settings); err != nil {
		return err
	}

	// Log query parameters
	log.Debug().
		Str("placeID", c.settings.PlaceID).
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
		PlaceID: c.settings.PlaceID,
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
		Str("placeID", c.settings.PlaceID).
		Str("name", resp.Name).
		Msg("Place details retrieved")

	return nil
}

type NearbySettings struct {
	Location string `glazed.parameter:"location"`
	Radius   int    `glazed.parameter:"radius"`
	Type     string `glazed.parameter:"type"`
	Keyword  string `glazed.parameter:"keyword"`
}

type NearbyCommand struct {
	*cmds.CommandDescription
	settings NearbySettings
}

func NewNearbyCommand() (*cobra.Command, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	cmd := &NearbyCommand{
		CommandDescription: cmds.NewCommandDescription(
			"nearby",
			cmds.WithShort("Search for nearby places"),
			cmds.WithLong("Search for places near a specific location using various filters"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"location",
					parameters.ParameterTypeString,
					parameters.WithHelp("Location in lat,lng format (e.g. 40.7128,-74.0060)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"radius",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Search radius in meters"),
					parameters.WithDefault(1500),
				),
				parameters.NewParameterDefinition(
					"type",
					parameters.ParameterTypeString,
					parameters.WithHelp("Type of place (e.g. restaurant, museum)"),
				),
				parameters.NewParameterDefinition(
					"keyword",
					parameters.ParameterTypeString,
					parameters.WithHelp("Keyword to search for"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}

	return cli.BuildCobraCommandFromGlazeCommand(cmd)
}

func (c *NearbyCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, &c.settings); err != nil {
		return err
	}

	// Log query parameters
	log.Debug().
		Str("location", c.settings.Location).
		Int("radius", c.settings.Radius).
		Str("type", c.settings.Type).
		Str("keyword", c.settings.Keyword).
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
	parts := strings.Split(c.settings.Location, ",")
	if len(parts) != 2 {
		return fmt.Errorf("invalid location format, expected lat,lng but got: %s", c.settings.Location)
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
		Radius:  uint(c.settings.Radius),
		Keyword: c.settings.Keyword,
	}

	// Add type if provided
	if c.settings.Type != "" {
		req.Type = maps.PlaceType(c.settings.Type)
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
		Str("location", c.settings.Location).
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
