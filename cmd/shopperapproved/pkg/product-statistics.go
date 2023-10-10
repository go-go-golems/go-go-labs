package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ProductStatistics represents aggregated product feedback statistics.
type ProductStatistics struct {
	CertificateURL string `json:"certificate_url,omitempty"`
	ProductTotals  struct {
		TotalReviews      int     `json:"total_reviews"`
		AverageRating     float64 `json:"average_rating"`
		TotalWithComments int     `json:"total_with_comments"`
	} `json:"product_totals"`
	SiteTotals struct {
		TotalReviews      int     `json:"total_reviews"`
		AverageRating     float64 `json:"average_rating"`
		TotalWithComments int     `json:"total_with_comments"`
	} `json:"site_totals,omitempty"`
	ProductAggregates map[string]struct {
		TotalReviews      int     `json:"total_reviews"`
		AverageRating     float64 `json:"average_rating"`
		TotalWithComments int     `json:"total_with_comments"`
	} `json:"product_totals,omitempty"`
}

// AggregateRequestParams represents parameters for the API request to fetch aggregate product statistics.
type AggregateRequestParams struct {
	ProductID  *int
	ByMatchKey *bool
	AsArray    *bool
	SiteOnly   *bool
	FastMode   *bool
}

// AggregateRequestOption defines the type of the function that will set the options.
type AggregateRequestOption func(*AggregateRequestParams)

// NewAggregateRequestParams initializes a new AggregateRequestParams with given options.
func NewAggregateRequestParams(opts ...AggregateRequestOption) *AggregateRequestParams {
	params := &AggregateRequestParams{}
	for _, opt := range opts {
		opt(params)
	}
	return params
}

// WithProductID sets the ProductID for the AggregateRequestParams.
func WithProductID(productID int) AggregateRequestOption {
	return func(params *AggregateRequestParams) {
		params.ProductID = &productID
	}
}

func (client *ShopperApprovedClient) FetchAggregateStatistics(params *AggregateRequestParams) (*ProductStatistics, error) {
	const baseURL = "https://api.shopperapproved.com/aggregates/products"

	var url string
	if params.ProductID != nil {
		url = fmt.Sprintf("%s/%d/%d", baseURL, client.SiteID, *params.ProductID)
	} else {
		url = fmt.Sprintf("%s/%d", baseURL, client.SiteID)
	}

	queryParams := []string{fmt.Sprintf("token=%s", client.Token)}

	if params.AsArray != nil {
		queryParams = append(queryParams, fmt.Sprintf("asArray=%t", *params.AsArray))
	}
	if params.ByMatchKey != nil {
		queryParams = append(queryParams, fmt.Sprintf("by_match_key=%t", *params.ByMatchKey))
	}
	if params.SiteOnly != nil {
		queryParams = append(queryParams, fmt.Sprintf("siteOnly=%t", *params.SiteOnly))
	}
	if params.FastMode != nil {
		queryParams = append(queryParams, fmt.Sprintf("fastmode=%t", *params.FastMode))
	}

	// Join the base URL with the query parameters
	url += "?" + strings.Join(queryParams, "&")

	// Make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	// Read and unmarshal the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var stats ProductStatistics
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}
