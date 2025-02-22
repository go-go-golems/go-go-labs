package places

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

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

	cobraCmd := &cobra.Command{
		Use:   "search",
		Short: cmd.Short,
		Long:  cmd.Long,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			ctx := cobraCmd.Context()
			parsedLayers := &layers.ParsedLayers{}
			return cmd.Run(ctx, parsedLayers)
		},
	}

	cobraCmd.Flags().StringVarP(&cmd.query, "query", "q", "", "Text to search for")
	cobraCmd.Flags().StringVarP(&cmd.location, "location", "l", "", "Location in lat,lng format (e.g. 40.7128,-74.0060)")
	cobraCmd.Flags().IntVarP(&cmd.radius, "radius", "r", 1500, "Search radius in meters")
	cobraCmd.Flags().StringVarP(&cmd.typeOf, "type", "t", "", "Type of place (e.g. restaurant, museum)")

	_ = cobraCmd.MarkFlagRequired("query")

	return cobraCmd, nil
}

func (c *SearchCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	// Log query parameters
	log.Debug().
		Str("query", c.query).
		Str("location", c.location).
		Int("radius", c.radius).
		Str("type", c.typeOf).
		Msg("Executing place search")

	// TODO: Implement the actual search functionality using the Google Maps client
	fmt.Printf("Searching for places with query: %s\n", c.query)
	if c.location != "" {
		fmt.Printf("Near location: %s\n", c.location)
	}
	if c.typeOf != "" {
		fmt.Printf("Of type: %s\n", c.typeOf)
	}
	fmt.Printf("Within radius: %d meters\n", c.radius)

	// Log mock results for now
	log.Debug().
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

	cobraCmd := &cobra.Command{
		Use:   "details",
		Short: cmd.Short,
		Long:  cmd.Long,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			ctx := cobraCmd.Context()
			parsedLayers := &layers.ParsedLayers{}
			return cmd.Run(ctx, parsedLayers)
		},
	}

	cobraCmd.Flags().StringVarP(&cmd.placeID, "place-id", "p", "", "Place ID to get details for")
	_ = cobraCmd.MarkFlagRequired("place-id")

	return cobraCmd, nil
}

func (c *DetailsCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	// Log query parameters
	log.Debug().
		Str("placeID", c.placeID).
		Msg("Fetching place details")

	// TODO: Implement the actual details retrieval using the Google Maps client
	fmt.Printf("Getting details for place ID: %s\n", c.placeID)

	// Log mock results for now
	log.Debug().
		Str("placeID", c.placeID).
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

	cobraCmd := &cobra.Command{
		Use:   "nearby",
		Short: cmd.Short,
		Long:  cmd.Long,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			ctx := cobraCmd.Context()
			parsedLayers := &layers.ParsedLayers{}
			return cmd.Run(ctx, parsedLayers)
		},
	}

	cobraCmd.Flags().StringVarP(&cmd.location, "location", "l", "", "Location in lat,lng format (e.g. 40.7128,-74.0060)")
	cobraCmd.Flags().IntVarP(&cmd.radius, "radius", "r", 1500, "Search radius in meters")
	cobraCmd.Flags().StringVarP(&cmd.typeOf, "type", "t", "", "Type of place (e.g. restaurant, museum)")
	cobraCmd.Flags().StringVarP(&cmd.keyword, "keyword", "k", "", "Keyword to search for")

	_ = cobraCmd.MarkFlagRequired("location")

	return cobraCmd, nil
}

func (c *NearbyCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	// Log query parameters
	log.Debug().
		Str("location", c.location).
		Int("radius", c.radius).
		Str("type", c.typeOf).
		Str("keyword", c.keyword).
		Msg("Searching for nearby places")

	// TODO: Implement the actual nearby search using the Google Maps client
	fmt.Printf("Searching for places near %s\n", c.location)
	fmt.Printf("Within radius: %d meters\n", c.radius)
	if c.typeOf != "" {
		fmt.Printf("Of type: %s\n", c.typeOf)
	}
	if c.keyword != "" {
		fmt.Printf("With keyword: %s\n", c.keyword)
	}

	// Log mock results for now
	log.Debug().
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
