package cmds

import (
	"context"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/mastoid/pkg"
	"github.com/mattn/go-mastodon"
	"github.com/spf13/cobra"
)

var RegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register the app with the Mastodon instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientName, _ := cmd.Flags().GetString("client-name")
		redirectURIs, _ := cmd.Flags().GetString("redirect-uris")
		scopes, _ := cmd.Flags().GetString("scopes")
		website, _ := cmd.Flags().GetString("website")
		server, _ := cmd.Flags().GetString("server")

		ctx := context.Background()
		appConfig := &mastodon.AppConfig{
			ClientName:   clientName,
			RedirectURIs: redirectURIs,
			Scopes:       scopes,
			Website:      website,
			Server:       server,
		}

		app, err := mastodon.RegisterApp(ctx, appConfig)
		if err != nil {
			return fmt.Errorf("Error registering app: %w", err)
		}

		credentials := &pkg.Credentials{
			Server:      server,
			Application: app,
		}

		fmt.Printf("App registration successful!\n")
		fmt.Printf("Client ID: %s\n", app.ClientID)
		fmt.Printf("Client Secret: %s\n", app.ClientSecret)
		fmt.Printf("Auth URI: %s\n", app.AuthURI)
		fmt.Printf("Redirect URI: %s\n", app.RedirectURI)

		err = pkg.Authorize(ctx, credentials)
		if err != nil {
			return fmt.Errorf("Error authorizing app: %w", err)
		}

		err = pkg.StoreCredentials(credentials)
		if err != nil {
			return fmt.Errorf("Error storing credentials: %w", err)
		}

		fmt.Printf("Grant Token: %s\n", credentials.GrantToken)
		fmt.Printf("Access Token: %s\n", credentials.AccessToken)

		return nil
	},
}
