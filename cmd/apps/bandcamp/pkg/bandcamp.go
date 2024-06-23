package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

type SearchType string

const (
	FilterAlbum SearchType = "a"
	FilterBand  SearchType = "b"
	FilterTrack SearchType = "t"
	FilterAll   SearchType = ""
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
		Results      []*Result `json:"results"`
		StatParams   string    `json:"stat_params_for_tag"`
		ResponseTime int       `json:"time_ms"`
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
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
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
