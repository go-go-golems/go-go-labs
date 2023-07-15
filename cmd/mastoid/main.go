package main

import (
	"context"
	"fmt"
	"github.com/go-go-golems/go-go-labs/pkg/render"
	"github.com/go-go-golems/go-go-labs/pkg/render/html"
	"github.com/go-go-golems/go-go-labs/pkg/render/plaintext"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-mastodon"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mastoid",
	Short: "mastoid is a CLI app to interact with Mastodon",
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verifies the credentials of a Mastodon instance",
	Run: func(cmd *cobra.Command, args []string) {
		instance, _ := cmd.Flags().GetString("instance")

		ctx := context.Background()

		client, err := createClient(instance)
		cobra.CheckErr(err)

		err = client.AuthenticateApp(ctx)
		cobra.CheckErr(err)

		app, err := client.VerifyAppCredentials(ctx)
		cobra.CheckErr(err)

		fmt.Printf("App Name: %s\n", app.Name)
		fmt.Printf("App Website: %s\n", app.Website)
	},
}

var threadCmd = &cobra.Command{
	Use:   "thread",
	Short: "Retrieves a thread from a Mastodon instance",
	Run: func(cmd *cobra.Command, args []string) {
		instance, _ := cmd.Flags().GetString("instance")
		statusID, _ := cmd.Flags().GetString("status-id")
		verbose, _ := cmd.Flags().GetBool("verbose")
		withHtml, _ := cmd.Flags().GetBool("withHtml")
		withHeader, _ := cmd.Flags().GetBool("with-header")

		// extract statusID from URL if we have a URL
		if strings.Contains(statusID, "http") {
			statusID = strings.Split(statusID, "/")[4]
		}

		ctx := context.Background()

		client, err := createClient(instance)
		cobra.CheckErr(err)

		err = client.AuthenticateApp(ctx)
		cobra.CheckErr(err)

		app, err := client.VerifyAppCredentials(ctx)
		cobra.CheckErr(err)

		// display app information
		if verbose {
			fmt.Printf("App Name: %s\n", app.Name)
			fmt.Printf("App Website: %s\n", app.Website)
		}

		status, err := client.GetStatus(ctx, mastodon.ID(statusID))
		cobra.CheckErr(err)

		context, err := client.GetStatusContext(ctx, status.ID)
		cobra.CheckErr(err)

		var renderer render.Renderer

		if withHtml {
			renderer = html.NewRenderer(
				html.WithVerbose(verbose),
				html.WithHeader(withHeader),
			)
		} else {
			renderer = plaintext.NewRenderer(
				plaintext.WithVerbose(verbose),
				plaintext.WithHeader(withHeader),
			)
		}

		err = renderer.RenderThread(os.Stdout, status, context)
		cobra.CheckErr(err)
	},
}

var registerCmd = &cobra.Command{
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

		err = storeCredentials(app)
		if err != nil {
			return fmt.Errorf("Error storing credentials: %w", err)
		}

		fmt.Printf("App registration successful!\n")
		fmt.Printf("Client ID: %s\n", app.ClientID)
		fmt.Printf("Client Secret: %s\n", app.ClientSecret)

		return nil
	},
}

func initConfig() {
	viper.SetConfigName("config")                           // config file name without extension
	viper.SetConfigType("yaml")                             // or viper.SetConfigType("YAML")
	viper.AddConfigPath(filepath.Join("$HOME", ".mastoid")) // path to look for the config file in
	_ = viper.ReadInConfig()                                // read in config file
}

func storeCredentials(app *mastodon.Application) error {
	viper.Set("client_id", app.ClientID)
	viper.Set("client_secret", app.ClientSecret)

	return viper.WriteConfig()
}

func loadCredentials() (*mastodon.Application, error) {
	clientId := viper.GetString("client_id")
	clientSecret := viper.GetString("client_secret")

	// check that they are valid
	if clientId == "" || clientSecret == "" {
		return nil, fmt.Errorf("no credentials found")
	}

	app := &mastodon.Application{
		ClientID:     clientId,
		ClientSecret: clientSecret,
	}
	return app, nil
}

func createClient(instance string) (*mastodon.Client, error) {
	credentials, err := loadCredentials()
	if err != nil {
		return nil, err
	}
	config := &mastodon.Config{
		Server:       instance,
		ClientID:     credentials.ClientID,
		ClientSecret: credentials.ClientSecret,
	}
	return mastodon.NewClient(config), nil
}

func main() {
	initConfig()

	threadCmd.Flags().StringP("instance", "i", "https://hachyderm.io", "Mastodon instance")
	threadCmd.Flags().StringP("status-id", "s", "", "Status ID")
	threadCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	threadCmd.Flags().Bool("html", false, "HTML output")
	threadCmd.Flags().Bool("with-header", true, "Print header")
	rootCmd.AddCommand(threadCmd)

	registerCmd.Flags().StringP("client-name", "n", "mastoid", "Client name")
	registerCmd.Flags().StringP("redirect-uris", "r", "urn:ietf:wg:oauth:2.0:oob", "Redirect URIs")
	registerCmd.Flags().StringP("scopes", "s", "read write follow", "Scopes")
	registerCmd.Flags().StringP("website", "w", "", "Website")
	registerCmd.Flags().StringP("server", "v", "https://hachyderm.io", "Mastodon instance")

	rootCmd.AddCommand(registerCmd)

	verifyCmd.Flags().StringP("instance", "i", "https://hachyderm.io", "Mastodon instance")
	rootCmd.AddCommand(verifyCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
