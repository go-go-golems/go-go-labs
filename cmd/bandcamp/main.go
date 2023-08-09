package main

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "bancamp_search",
		Short: "Search bandcamp",
		Long:  `Search for music on bandcamp`,
		Run: func(cmd *cobra.Command, args []string) {

			client := NewClient()

			filter, _ := cmd.Flags().GetString("filter")

			if len(args) == 0 {
				log.Fatal().Msg("please provide a search keyword")
			}

			searchResp, err := client.Search(context.Background(), args[0], SearchType(filter))
			if err != nil {
				log.Fatal().Err(err).Msg("failed to search")
			}

			results := searchResp.Auto.Results
			p := tea.NewProgram(NewModel(client, results), tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				fmt.Printf("Alas, there's been an error: %v", err)
				os.Exit(1)
			}

			selectedResult := results[0]

			playlist := &PlaylistSection{
				Tracks: []*Result{
					selectedResult,
				},
				LinkColor:       "white",
				BackgroundColor: "black",
			}

			s, err := playlist.Render()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to render playlist")
			}
			fmt.Println(s)

			//for _, result := range searchResp.Auto.Results {
			//	switch result.Type {
			//	case "a":
			//		fmt.Printf("Type: Album\n")
			//	case "t":
			//		fmt.Printf("Type: Track\n")
			//	case "b":
			//		fmt.Printf("Type: Band\n")
			//	default:
			//		fmt.Printf("Type: %s\n", result.Type)
			//	}
			//	fmt.Printf("Album Name: %s\n", result.AlbumName)
			//	fmt.Printf("Band Name: %s\n", result.BandName)
			//	fmt.Printf("Name: %s\n", result.Name)
			//	fmt.Printf("URL: %s%s\n\n", result.ItemURLRoot, result.ItemURLPath)
			//}
		},
	}

	rootCmd.Flags().StringP("filter", "f", "", "filter search results by type (album, band, track)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("failed to execute command")
	}
}
