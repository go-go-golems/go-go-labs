package main

import (
	"fmt"
	cmds "github.com/go-go-golems/go-go-labs/cmd/mastoid/cmds"
	"github.com/go-go-golems/go-go-labs/cmd/mastoid/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mastoid",
	Short: "mastoid is a CLI app to interact with Mastodon",
}

func main() {
	pkg.InitConfig()

	cmds.ThreadCmd.Flags().StringP("status-id", "s", "", "Status ID")
	cmds.ThreadCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	cmds.ThreadCmd.Flags().String("output", "markdown", "Output format (html, text, markdown, json)")
	cmds.ThreadCmd.Flags().Bool("with-header", true, "Print header")
	rootCmd.AddCommand(cmds.ThreadCmd)

	cmds.RegisterCmd.Flags().StringP("client-name", "n", "mastoid", "Client name")
	cmds.RegisterCmd.Flags().StringP("redirect-uris", "r", "urn:ietf:wg:oauth:2.0:oob", "Redirect URIs")
	cmds.RegisterCmd.Flags().StringP("scopes", "s", "read", "Scopes")
	cmds.RegisterCmd.Flags().StringP("website", "w", "", "Website")
	cmds.RegisterCmd.Flags().StringP("server", "v", "https://hachyderm.io", "Mastodon instance")

	rootCmd.AddCommand(cmds.RegisterCmd)

	rootCmd.AddCommand(cmds.AuthorizeCmd)

	rootCmd.AddCommand(cmds.VerifyCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
