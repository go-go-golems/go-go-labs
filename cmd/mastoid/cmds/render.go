package cmds

import (
	"context"
	"encoding/json"
	"github.com/go-go-golems/go-go-labs/cmd/mastoid/pkg"
	"github.com/go-go-golems/go-go-labs/cmd/mastoid/pkg/render"
	"github.com/go-go-golems/go-go-labs/cmd/mastoid/pkg/render/html"
	"github.com/go-go-golems/go-go-labs/cmd/mastoid/pkg/render/plaintext"
	"github.com/mattn/go-mastodon"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

var RenderCmd = &cobra.Command{
	Use:   "render",
	Short: "Retrieves a thread from a Mastodon instance and renders it",
	Run: func(cmd *cobra.Command, args []string) {
		statusID, _ := cmd.Flags().GetString("status-id")
		verbose, _ := cmd.Flags().GetBool("verbose")
		output, _ := cmd.Flags().GetString("output")
		withHeader, _ := cmd.Flags().GetBool("with-header")

		statusID = pkg.ExtractID(statusID)
		if statusID == "" {
			log.Error().Msg("No status ID provided")
			os.Exit(1)
		}

		ctx := context.Background()

		credentials, err := pkg.LoadCredentials()
		cobra.CheckErr(err)

		client, err := pkg.CreateClientAndAuthenticate(ctx, credentials)
		cobra.CheckErr(err)

		status, err := client.GetStatus(ctx, mastodon.ID(statusID))
		if err != nil {
			log.Error().Err(err).Str("statusId", statusID).Msg("Could not get status")
		}
		cobra.CheckErr(err)

		context, err := client.GetStatusContext(ctx, status.ID)
		cobra.CheckErr(err)

		var renderer render.Renderer

		switch output {
		case "json":
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			err = encoder.Encode(context)
			cobra.CheckErr(err)
			return
		case "html":

			renderer = html.NewRenderer(
				html.WithVerbose(verbose),
				html.WithHeader(withHeader),
			)
		case "text":
			renderer = plaintext.NewRenderer(
				plaintext.WithVerbose(verbose),
				plaintext.WithHeader(withHeader),
			)

		case "markdown":
			renderer = plaintext.NewRenderer(
				plaintext.WithVerbose(verbose),
				plaintext.WithHeader(withHeader),
				plaintext.WithMarkdown(),
			)
		default:
			cobra.CheckErr("Unknown output format")
		}

		err = renderer.RenderThread(os.Stdout, status, context)
		cobra.CheckErr(err)
	},
}
