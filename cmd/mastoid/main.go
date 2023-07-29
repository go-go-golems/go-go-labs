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

	"github.com/pkg/browser"

	"github.com/mattn/go-mastodon"
	"github.com/spf13/cobra"
	"github.com/yardbirdsax/bubblewrap"
)

// Credentials stores the result of an oauth flow.
type Credentials struct {
	Server      string
	GrantToken  string
	Application *mastodon.Application
	AccessToken string
}

var rootCmd = &cobra.Command{
	Use:   "mastoid",
	Short: "mastoid is a CLI app to interact with Mastodon",
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verifies the credentials of a Mastodon instance",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		credentials, err := loadCredentials()
		cobra.CheckErr(err)

		if credentials.Application.ClientID == "" || credentials.Application.ClientSecret == "" {
			fmt.Println("No client ID or secret found")
			return
		}

		client, err := createClient(credentials)
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

var threadCmd = &cobra.Command{
	Use:   "thread",
	Short: "Retrieves a thread from a Mastodon instance",
	Run: func(cmd *cobra.Command, args []string) {
		statusID, _ := cmd.Flags().GetString("status-id")
		verbose, _ := cmd.Flags().GetBool("verbose")
		withHtml, _ := cmd.Flags().GetBool("withHtml")
		withHeader, _ := cmd.Flags().GetBool("with-header")

		// extract statusID from URL if we have a URL
		if strings.Contains(statusID, "http") {
			statusID = strings.Split(statusID, "/")[4]
		}

		ctx := context.Background()

		credentials, err := loadCredentials()
		cobra.CheckErr(err)

		client, err := createClient(credentials)
		cobra.CheckErr(err)
		if client.Config.AccessToken == "" {
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

var authorizeCmd = &cobra.Command{
	Use:   "authorize",
	Short: "Authorize the app with the Mastodon instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		credentials, err := loadCredentials()
		if err != nil {
			return fmt.Errorf("Error loading credentials: %w", err)
		}

		err = authorize(context.Background(), credentials)
		if err != nil {
			return fmt.Errorf("Error authorizing app: %w", err)
		}

		err = storeCredentials(credentials)
		if err != nil {
			return fmt.Errorf("Error storing credentials: %w", err)
		}
		fmt.Printf("Grant Token: %s\n", credentials.GrantToken)
		fmt.Printf("Access Token: %s\n", credentials.AccessToken)

		return nil
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

		credentials := &Credentials{
			Server:      server,
			Application: app,
		}

		fmt.Printf("App registration successful!\n")
		fmt.Printf("Client ID: %s\n", app.ClientID)
		fmt.Printf("Client Secret: %s\n", app.ClientSecret)
		fmt.Printf("Auth URI: %s\n", app.AuthURI)
		fmt.Printf("Redirect URI: %s\n", app.RedirectURI)

		err = authorize(ctx, credentials)
		if err != nil {
			return fmt.Errorf("Error authorizing app: %w", err)
		}

		err = storeCredentials(credentials)
		if err != nil {
			return fmt.Errorf("Error storing credentials: %w", err)
		}

		fmt.Printf("Grant Token: %s\n", credentials.GrantToken)
		fmt.Printf("Access Token: %s\n", credentials.AccessToken)

		return nil
	},
}

func authorize(ctx context.Context, credentials_ *Credentials) error {
	// open app.AuthURI in browser
	err := browser.OpenURL(credentials_.Application.AuthURI)
	if err != nil {
		return fmt.Errorf("Error opening browser: %w", err)
	}

	isCodeValid := false

	var grantToken string

	for !isCodeValid {
		grantToken, err = bubblewrap.Input("Enter the code from the browser: ")
		if err != nil {
			return fmt.Errorf("Error reading user input: %w", err)
		}

		credentials_.GrantToken = grantToken
		fmt.Printf("Grant Token: %s\n", credentials_.GrantToken)

		client := mastodon.NewClient(&mastodon.Config{
			Server:       credentials_.Server,
			ClientID:     credentials_.Application.ClientID,
			ClientSecret: credentials_.Application.ClientSecret,
		})

		err = client.AuthenticateApp(ctx)
		if err != nil {
			fmt.Printf("Error authenticating app: %s\n", err)
			isCodeValid = false
			continue
		}

		err = client.AuthenticateToken(ctx, grantToken, credentials_.Application.RedirectURI)
		if err != nil {
			fmt.Printf("Error authenticating token: %s\n", err)
			isCodeValid = false
			continue
		}
		credentials_.AccessToken = client.Config.AccessToken
		fmt.Printf("Access Token: %s\n", credentials_.AccessToken)

		credentials, err := client.VerifyAppCredentials(ctx)
		if err != nil {
			fmt.Printf("Error verifying credentials: %s\n", err)
			isCodeValid = false
			continue
		}

		fmt.Printf("Website: %s\n", credentials.Website)
		fmt.Printf("Name: %s\n", credentials.Name)
		isCodeValid = true
	}

	return nil
}

func initConfig() {
	viper.SetConfigName("config")                           // config file name without extension
	viper.SetConfigType("yaml")                             // or viper.SetConfigType("YAML")
	viper.AddConfigPath(filepath.Join("$HOME", ".mastoid")) // path to look for the config file in
	_ = viper.ReadInConfig()                                // read in config file
}

func storeCredentials(credentials *Credentials) error {
	viper.Set("client_id", credentials.Application.ClientID)
	viper.Set("client_secret", credentials.Application.ClientSecret)
	viper.Set("auth_uri", credentials.Application.AuthURI)
	viper.Set("redirect_uri", credentials.Application.RedirectURI)
	viper.Set("grant_token", credentials.GrantToken)
	viper.Set("access_token", credentials.AccessToken)
	viper.Set("server", credentials.Server)

	err := viper.WriteConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// config file not found, create it
		err = viper.SafeWriteConfig()
		if err != nil {
			return fmt.Errorf("Error writing config: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error writing config: %w", err)
	}
	return nil
}

func loadCredentials() (*Credentials, error) {
	clientId := viper.GetString("client_id")
	clientSecret := viper.GetString("client_secret")
	authUri := viper.GetString("auth_uri")
	redirectUri := viper.GetString("redirect_uri")
	grantToken := viper.GetString("grant_token")
	accessToken := viper.GetString("access_token")
	server := viper.GetString("server")

	// check that they are valid
	if clientId == "" || clientSecret == "" {
		return nil, fmt.Errorf("no credentials found")
	}

	app := &Credentials{
		Server:      server,
		GrantToken:  grantToken,
		AccessToken: accessToken,
		Application: &mastodon.Application{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			AuthURI:      authUri,
			RedirectURI:  redirectUri,
		},
	}
	return app, nil
}

func createClient(credentials *Credentials) (*mastodon.Client, error) {
	config := &mastodon.Config{
		Server:       credentials.Server,
		ClientID:     credentials.Application.ClientID,
		ClientSecret: credentials.Application.ClientSecret,
		AccessToken:  credentials.AccessToken,
	}
	return mastodon.NewClient(config), nil
}

func main() {
	initConfig()

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

	rootCmd.AddCommand(authorizeCmd)

	rootCmd.AddCommand(verifyCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
