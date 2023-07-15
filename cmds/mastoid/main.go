package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"os"
	"strings"

	"github.com/mattn/go-mastodon"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mastoid",
	Short: "mastoid is a CLI app to interact with Mastodon",
}

var threadCmd = &cobra.Command{
	Use:   "thread",
	Short: "Retrieves a thread from a Mastodon instance",
	Run: func(cmd *cobra.Command, args []string) {
		instance, _ := cmd.Flags().GetString("instance")
		statusID, _ := cmd.Flags().GetString("status-id")
		verbose, _ := cmd.Flags().GetBool("verbose")
		html, _ := cmd.Flags().GetBool("html")

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

		printThread(status, context, verbose, html)
	},
}

func printStatus(status *mastodon.Status, verbose bool, html bool) {
	if verbose {
		fmt.Printf("Status ID: %s\n", status.ID)
		fmt.Printf("Created at: %v\n", status.CreatedAt)
		if html {
			fmt.Printf("Content: %s\n", status.Content)
		} else {
			fmt.Printf("Content: %s\n", convertHTMLToPlainText(status.Content))
		}
		fmt.Println("-----------------")
	} else {
		if html {
			fmt.Println(status.Content)
		} else {
			fmt.Println(convertHTMLToPlainText(status.Content))
		}
	}
}

func printThread(status *mastodon.Status, context *mastodon.Context, verbose bool, html bool) {
	printStatus(status, verbose, html)

	for _, ancestor := range context.Ancestors {
		if verbose {
			fmt.Println("--AN--")
		}
		printStatus(ancestor, verbose, html)
	}

	if verbose {
		fmt.Println("--OR--")
	}

	for _, descendant := range context.Descendants {
		if verbose {
			fmt.Println("--DE--")
		}
		printStatus(descendant, verbose, html)
	}
}

func convertHTMLToPlainText(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	var f func(*html.Node)
	var output strings.Builder
	blockTags := map[string]struct{}{
		"p":       {},
		"div":     {},
		"br":      {},
		"article": {},
		"section": {},
		"li":      {},
	}

	f = func(n *html.Node) {
		// If the current node is a text node
		if n.Type == html.TextNode {
			output.WriteString(n.Data)
		}

		// If the current node is one of the block tags types
		if _, present := blockTags[n.Data]; present {
			output.WriteRune('\n')
		}

		// Recurse on the children nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
	return output.String()
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

func storeCredentials(app *mastodon.Application) error {
	file, err := os.OpenFile("credentials.json", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	encoder := json.NewEncoder(file)
	return encoder.Encode(app)
}

func loadCredentials() (*mastodon.Application, error) {
	file, err := os.Open("credentials.json")
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	decoder := json.NewDecoder(file)

	var app *mastodon.Application
	err = decoder.Decode(&app)
	if err != nil {
		return nil, err
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
	threadCmd.Flags().StringP("instance", "i", "https://hachyderm.io", "Mastodon instance")
	threadCmd.Flags().StringP("status-id", "s", "", "Status ID")
	threadCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	threadCmd.Flags().Bool("html", false, "HTML output")
	rootCmd.AddCommand(threadCmd)

	registerCmd.Flags().StringP("client-name", "n", "mastoid", "Client name")
	registerCmd.Flags().StringP("redirect-uris", "r", "urn:ietf:wg:oauth:2.0:oob", "Redirect URIs")
	registerCmd.Flags().StringP("scopes", "s", "read write follow", "Scopes")
	registerCmd.Flags().StringP("website", "w", "", "Website")
	registerCmd.Flags().StringP("server", "v", "https://hachyderm.io", "Mastodon instance")

	rootCmd.AddCommand(registerCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
