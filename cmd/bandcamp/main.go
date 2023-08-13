package main

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/bandcamp/pkg"
	"github.com/go-go-golems/go-go-labs/cmd/bandcamp/ui/playlist"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "bancamp_search",
		Short: "Search bandcamp", Long: `Search for music on bandcamp`,
		Run: func(cmd *cobra.Command, args []string) {

			client := pkg.NewClient()

			filter, _ := cmd.Flags().GetString("filter")

			if len(args) == 0 {
				log.Fatal().Msg("please provide a search keyword")
			}

			searchResp, err := client.Search(context.Background(), args[0], pkg.SearchType(filter))
			if err != nil {
				log.Fatal().Err(err).Msg("failed to search")
			}

			results := searchResp.Auto.Results[:3]
			tracks_ := make([]*pkg.Track, len(results))

			for i, result := range results {
				tracks_[i] = &pkg.Track{
					BackgroundColor: "black",
					LinkColor:       "white",
					AlbumID:         result.AlbumID,
					Name:            result.Name,
					BandName:        result.BandName,
					ItemURLPath:     result.ItemURLPath,
				}
			}

			playlist_ := &pkg.Playlist{
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
