package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Review represents a single product review.
type Review struct {
	ReviewID     int64   `json:"review_id"`
	DisplayName  string  `json:"display_name"`
	Date         string  `json:"date"`
	ProductID    string  `json:"product_id"`
	Rating       float64 `json:"rating"`
	Comments     string  `json:"comments"`
	Public       int     `json:"public"`
	Response     string  `json:"response"`
	Location     string  `json:"location"`
	CustCareOpen int     `json:"custcareopen"`
}

// ShopperApprovedClient represents a client to access ShopperApproved API.
type ShopperApprovedClient struct {
	SiteID int
	Token  string
}

// ReviewRequestParams represents parameters for the API request to fetch product reviews.
type ReviewRequestParams struct {
	// ProductID represents the product ID or parent ID you would like reviews for.
	ProductID string
	// AsArray, if true, will return the response as a JSON array as opposed to the standard JSON object.
	AsArray *bool
	// Limit specifies how many reviews you want returned in the response. Larger numbers may cause a timeout.
	Limit *int
	// Page indicates which page you would like to request. The offset will be calculated by limit * page. Starts at 0.
	Page *int
	// From represents the date you would like to start the query with. It should be given in YYYY-MM-DD format. Defaults to 30 days prior to the current day.
	From *string
	// To represents the date you would like to end the query with. It should be given in YYYY-MM-DD format. Defaults to the current date.
	To *string
	// Sort indicates how you would like to sort the reviews. Values are newest, oldest, highest, lowest.
	Sort *string
	// Removed, if set to 1, will include reviews that have a 'removed' value equal to 1 if the review was removed and 0 if the review is active.
	Removed *int
}

// ReviewRequestOption defines the type of the function that will set the options.
type ReviewRequestOption func(*ReviewRequestParams)

// NewReviewRequestParams initializes a new ReviewRequestParams with given options.
func NewReviewRequestParams(productId string, opts ...ReviewRequestOption) *ReviewRequestParams {
	params := &ReviewRequestParams{
		ProductID: productId,
	}
	for _, opt := range opts {
		opt(params)
	}
	return params
}

// WithAsArray sets the AsArray in the request parameters.
func WithAsArray(asArray bool) ReviewRequestOption {
	return func(params *ReviewRequestParams) {
		params.AsArray = &asArray
	}
}

// WithLimit sets the Limit in the request parameters.
func WithLimit(limit int) ReviewRequestOption {
	return func(params *ReviewRequestParams) {
		params.Limit = &limit
	}
}

// WithPage sets the Page in the request parameters.
func WithPage(page int) ReviewRequestOption {
	return func(params *ReviewRequestParams) {
		params.Page = &page
	}
}

// WithFrom sets the From in the request parameters.
func WithFrom(from string) ReviewRequestOption {
	return func(params *ReviewRequestParams) {
		params.From = &from
	}
}

// WithTo sets the To in the request parameters.
func WithTo(to string) ReviewRequestOption {
	return func(params *ReviewRequestParams) {
		params.To = &to
	}
}

// WithSort sets the Sort in the request parameters.
func WithSort(sort string) ReviewRequestOption {
	return func(params *ReviewRequestParams) {
		params.Sort = &sort
	}
}

// WithRemoved sets the Removed in the request parameters.
func WithRemoved(removed int) ReviewRequestOption {
	return func(params *ReviewRequestParams) {
		params.Removed = &removed
	}
}

func (client *ShopperApprovedClient) FetchReviews(params *ReviewRequestParams) (map[string]Review, error) {
	const baseURL = "https://api.shopperapproved.com/products/reviews"

	url := fmt.Sprintf("%s/%d/%s", baseURL, client.SiteID, params.ProductID)

	queryParams := []string{fmt.Sprintf("token=%s", client.Token)}

	if params.AsArray != nil {
		queryParams = append(queryParams, fmt.Sprintf("asArray=%t", *params.AsArray))
	}
	if params.Limit != nil {
		queryParams = append(queryParams, fmt.Sprintf("limit=%d", *params.Limit))
	}
	if params.Page != nil {
		queryParams = append(queryParams, fmt.Sprintf("page=%d", *params.Page))
	}
	if params.From != nil {
		queryParams = append(queryParams, fmt.Sprintf("from=%s", *params.From))
	}
	if params.To != nil {
		queryParams = append(queryParams, fmt.Sprintf("to=%s", *params.To))
	}
	if params.Sort != nil {
		queryParams = append(queryParams, fmt.Sprintf("sort=%s", *params.Sort))
	}
	if params.Removed != nil {
		queryParams = append(queryParams, fmt.Sprintf("removed=%d", *params.Removed))
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

	var reviews map[string]Review
	if err := json.Unmarshal(body, &reviews); err != nil {
		return nil, err
	}

	return reviews, nil
}

func main() {
	// Initialize the client with siteID and token
	client := &ShopperApprovedClient{
		SiteID: 14431,
		Token:  "KhQxb0dP",
	}

	productID := "3699"
	reviews, err := client.FetchReviews(NewReviewRequestParams(productID, WithLimit(10), WithPage(0)))
	if err != nil {
		fmt.Printf("Failed to fetch reviews: %s\n", err)
		return
	}

	for _, review := range reviews {
		date, _ := time.Parse("Mon, 2 Jan 2006 15:04:05 MST", review.Date)
		fmt.Printf("Date: %s\nReviewer: %s\nRating: %.1f\nComment: %s\n\n", date.Format("2006-01-02"), review.DisplayName, review.Rating, review.Comments)
	}
}
