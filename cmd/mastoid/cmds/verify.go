package cmds

import (
	"context"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/mastoid/pkg"
	"github.com/spf13/cobra"
)

var VerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verifies the credentials of a Mastodon instance",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		credentials, err := pkg.LoadCredentials()
		cobra.CheckErr(err)

		if credentials.Application.ClientID == "" || credentials.Application.ClientSecret == "" {
			fmt.Println("No client ID or secret found")
			return
		}

		client, err := pkg.CreateClient(credentials)
		cobra.CheckErr(err)

		if client.Config.AccessToken == "" {
			fmt.Println("No access token found")
			if credentials.GrantToken != "" {
				fmt.Println("Authenticating with grant token")
				err = client.AuthenticateToken(ctx, credentials.GrantToken, credentials.Application.RedirectURI)
				cobra.CheckErr(err)

				fmt.Println("Grant token authenticated")
			} else {
				fmt.Println("Authenticating with app")
				err = client.AuthenticateApp(ctx)
				cobra.CheckErr(err)
				fmt.Println("App authenticated")
			}
		} else {
			fmt.Printf("Access token found: %s\n", client.Config.AccessToken)
		}

		app, err := client.VerifyAppCredentials(ctx)
		cobra.CheckErr(err)

		fmt.Printf("App Name: %s\n", app.Name)
		fmt.Printf("App Website: %s\n", app.Website)
	},
}
