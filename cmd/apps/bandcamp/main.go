package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill/message"
	tea "github.com/charmbracelet/bubbletea"
	pkg2 "github.com/go-go-golems/go-go-labs/cmd/apps/bandcamp/pkg"
	machinery2 "github.com/go-go-golems/go-go-labs/cmd/apps/bandcamp/pkg/machinery"
	"github.com/go-go-golems/go-go-labs/cmd/apps/bandcamp/ui/playlist"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "bancamp_search",
		Short: "Search bandcamp", Long: `Search for music on bandcamp`,
		Run: func(cmd *cobra.Command, args []string) {
			client := pkg2.NewClient()
			machine, err := machinery2.NewMachine()
			httpServer := machinery2.NewHTTPServer()
			cobra.CheckErr(err)

			machine.Router.AddNoPublisherHandler(
				"httpServer",
				"playlist",
				machine.PubSub,
				func(msg *message.Message) error {
					playlist := &pkg2.Playlist{}
					if err := json.Unmarshal(msg.Payload, playlist); err != nil {
						return err
					}

					httpServer.HandlePlaylist(playlist)
					return nil
				},
			)

			filter, _ := cmd.Flags().GetString("filter")

			if len(args) == 0 {
				log.Fatal().Msg("please provide a search keyword")
			}

			searchResp, err := client.Search(context.Background(), args[0], pkg2.SearchType(filter))
			if err != nil {
				log.Fatal().Err(err).Msg("failed to search")
			}

			results := searchResp.Auto.Results[:3]
			tracks_ := make([]*pkg2.Track, len(results))

			for i, result := range results {
				tracks_[i] = &pkg2.Track{
					BackgroundColor: "black",
					LinkColor:       "white",
					AlbumID:         result.AlbumID,
					Name:            result.Name,
					BandName:        result.BandName,
					ItemURLPath:     result.ItemURLPath,
				}
			}

			// TODO(manuel, 2023-08-16) A cool feature would be to expose the playlist
			// as a render webpage immediately, so that one can see the final result.

			playlist_ := &pkg2.Playlist{
				Title:       "Summer Playlist",
				Description: "Foobar playlist",
				Tracks:      tracks_,
			}

			m := playlist.NewModel(playlist_)

			p := tea.NewProgram(m, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				fmt.Printf("Alas, there's been an error: %v", err)
				os.Exit(1)
			}

			s, err := playlist_.Render()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to render playlist")
			}
			fmt.Println(s)
		},
	}

	rootCmd.Flags().StringP("filter", "f", "", "filter search results by type (album, band, track)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("failed to execute command")
	}
}
