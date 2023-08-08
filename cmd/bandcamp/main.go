package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"text/template"
)

type SimpleResult struct {
	AlbumID     int64  `json:"album_id"`
	ArtID       int64  `json:"art_id"`
	BandID      int64  `json:"band_id"`
	BandName    string `json:"band_name"`
	ImgID       int64  `json:"img_id"`
	ItemURLPath string `json:"item_url_path"`
	AlbumName   string `json:"album_name"`
}

type Playlist struct {
	LinkColor       string         `json:"linkColor"`
	BackgroundColor string         `json:"backgroundColor"`
	Tracks          []SimpleResult `json:"tracks"`
}

const iframeTmpl = `<iframe style="border: 0; width: 100%%; height: 42px; background-color: {{.BackgroundColor}};" src="https://bandcamp.com/EmbeddedPlayer/album={{.AlbumID}}/size=small/bgcol={{.BackgroundColor}}/linkcol={{.LinkColor}}/transparent=true/" seamless><a href="https://bandcamp.com/{{.ItemURLPath}}">{{.AlbumName}} by {{.BandName}}</a></iframe>`

func (p *Playlist) Render() (string, error) {
	var out bytes.Buffer

	tmpl, err := template.New("playlist").Parse(iframeTmpl)
	if err != nil {
		return "", err
	}

	for _, track := range p.Tracks {

		// A struct that fulfills the template
		data := struct {
			BackgroundColor string
			AlbumID         int64
			LinkColor       string
			ItemURLPath     string
			AlbumName       string
			BandName        string
		}{
			p.BackgroundColor,
			track.AlbumID,
			p.LinkColor,
			track.ItemURLPath,
			track.AlbumName,
			track.BandName,
		}

		// Apply the data to the template
		if err := tmpl.Execute(&out, data); err != nil {
			return "", err
		}

		// Add a separator between tracks
		if err := out.WriteByte('\n'); err != nil {
			return "", err
		}
	}
	return out.String(), nil
}

func generatePlaylists(jsonData []byte) ([]string, error) {
	var playlists []Playlist
	err := json.Unmarshal(jsonData, &playlists)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, playlist := range playlists {
		for _, track := range playlist.Tracks {
			iframe := fmt.Sprintf(`<iframe style="border: 0; width: 100%%; height: 42px; background-color: %s;" src="https://bandcamp.com/EmbeddedPlayer/album=%d/size=small/bgcol=%s/linkcol=%s/transparent=true/" seamless><a href="https://bandcamp.com/%s">%s by %s</a></iframe>`,
				playlist.BackgroundColor,
				track.AlbumID,
				playlist.BackgroundColor,
				playlist.LinkColor,
				track.ItemURLPath,
				track.AlbumName,
				track.BandName)
			result = append(result, iframe)
		}
	}
	return result, nil
}

type SearchType string

const (
	Album SearchType = "a"
	Band  SearchType = "b"
	Track SearchType = "t"
	All   SearchType = ""
)

// Client is a simple HTTP client.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

// NewClient returns a new Client.
func NewClient() *Client {
	return &Client{
		HTTPClient: http.DefaultClient,
		BaseURL:    "https://bandcamp.com/api/bcsearch_public_api/1",
	}
}

type SearchResponse struct {
	Auto struct {
		Results      []Result `json:"results"`
		StatParams   string   `json:"stat_params_for_tag"`
		ResponseTime int      `json:"time_ms"`
	} `json:"auto"`
	Genre struct{} `json:"genre"`
	Tag   struct {
		Count        int `json:"count"`
		Matches      []struct{}
		ResponseTime int `json:"time_ms"`
	} `json:"tag"`
}

type Result struct {
	AlbumID     int64  `json:"album_id"`
	AlbumName   string `json:"album_name"`
	ArtID       int64  `json:"art_id"`
	BandID      int64  `json:"band_id"`
	BandName    string `json:"band_name"`
	ID          int64  `json:"id"`
	Img         string `json:"img"`
	ImgID       int64  `json:"img_id"`
	ItemURLPath string `json:"item_url_path"`
	ItemURLRoot string `json:"item_url_root"`
	Name        string `json:"name"`
	StatParams  string `json:"stat_params"`
	Type        string `json:"type"`
}

type SearchRequest struct {
	FanID        *int       `json:"fan_id"`
	FullPage     bool       `json:"full_page"`
	SearchFilter SearchType `json:"search_filter"`
	SearchText   string     `json:"search_text"`
}

func (c *Client) Search(ctx context.Context, query string, filter SearchType) (*SearchResponse, error) {
	url := "https://bandcamp.com/api/bcsearch_public_api/1/autocomplete_elastic"

	searchReq := &SearchRequest{
		FanID:        nil,
		FullPage:     false,
		SearchFilter: filter,
		SearchText:   query,
	}

	b, err := json.Marshal(searchReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var searchResp SearchResponse
	s, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()
	if err := json.NewDecoder(bytes.NewBuffer(s)).Decode(&searchResp); err != nil {
		return nil, err
	}

	return &searchResp, nil
}

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

			for _, result := range searchResp.Auto.Results {
				switch result.Type {
				case "a":
					fmt.Printf("Type: Album\n")
				case "t":
					fmt.Printf("Type: Track\n")
				case "b":
					fmt.Printf("Type: Band\n")
				default:
					fmt.Printf("Type: %s\n", result.Type)
				}
				fmt.Printf("Album Name: %s\n", result.AlbumName)
				fmt.Printf("Band Name: %s\n", result.BandName)
				fmt.Printf("Name: %s\n", result.Name)
				fmt.Printf("URL: %s%s\n\n", result.ItemURLRoot, result.ItemURLPath)
			}
		},
	}

	rootCmd.Flags().StringP("filter", "f", "", "filter search results by type (album, band, track)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("failed to execute command")
	}
}
