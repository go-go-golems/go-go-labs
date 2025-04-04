package cmds

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/go-go-labs/cmd/apps/maps/cmds/auth"
	"github.com/go-go-golems/go-go-labs/cmd/apps/maps/cmds/directions"
	"github.com/go-go-golems/go-go-labs/cmd/apps/maps/cmds/places"
	"github.com/spf13/cobra"
	"googlemaps.github.io/maps"
)

type RootCommand struct {
	*cmds.CommandDescription
	apiKey string
}

func NewRootCommand() (*cobra.Command, error) {
	root := &RootCommand{}

	description := cmds.NewCommandDescription(
		"maps",
		cmds.WithShort("Google Maps CLI tool"),
		cmds.WithLong("A CLI tool for interacting with Google Maps API"),
	)
	root.CommandDescription = description

	rootCmd := &cobra.Command{
		Use:   "maps",
		Short: description.Short,
		Long:  description.Long,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Store the root command in the context for subcommands
			cmd.SetContext(context.WithValue(cmd.Context(), "rootCmd", root))
		},
	}

	// Add flag for API key
	rootCmd.PersistentFlags().StringVar(&root.apiKey, "api-key",
		os.Getenv("GOOGLE_MAPS_API_KEY"),
		"Google Maps API Key (can also be set via GOOGLE_MAPS_API_KEY environment variable)")

	// Add auth commands
	authCmd, err := auth.NewAuthCommand()
	if err != nil {
		return nil, fmt.Errorf("failed to create auth command: %w", err)
	}
	rootCmd.AddCommand(authCmd)

	// Add places commands
	placesCmd, err := places.NewPlacesCommand()
	if err != nil {
		return nil, fmt.Errorf("failed to create places command: %w", err)
	}
	rootCmd.AddCommand(placesCmd)

	// Add directions commands
	directionsCmd, err := directions.NewDirectionsCommand()
	if err != nil {
		return nil, fmt.Errorf("failed to create directions command: %w", err)
	}
	rootCmd.AddCommand(directionsCmd)

	return rootCmd, nil
}

// GetMapsClient creates and returns a Maps client
func (r *RootCommand) GetMapsClient(ctx context.Context) (*maps.Client, error) {
	if r.apiKey == "" {
		return nil, fmt.Errorf("Google Maps API key not provided. Set it via --api-key flag or GOOGLE_MAPS_API_KEY environment variable")
	}

	client, err := maps.NewClient(maps.WithAPIKey(r.apiKey))
	if err != nil {
		return nil, fmt.Errorf("unable to create Maps client: %v", err)
	}

	return client, nil
}
