package cmds

import (
	"context"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/mastoid/pkg"
	"github.com/spf13/cobra"
)

var AuthorizeCmd = &cobra.Command{
	Use:   "Authorize",
	Short: "Authorize the app with the Mastodon instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		credentials, err := pkg.LoadCredentials()
		if err != nil {
			return fmt.Errorf("Error loading credentials: %w", err)
		}

		err = pkg.Authorize(context.Background(), credentials)
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
