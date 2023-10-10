package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Review represents a single product review.
// Review represents a single product review.
type Review struct {
	ReviewID        int64   `json:"review_id,omitempty"`
	ProductReviewID int64   `json:"product_review_id,omitempty"`
	DisplayName     string  `json:"display_name"`
	EmailAddress    string  `json:"email_address,omitempty"`
	OrderID         string  `json:"order_id,omitempty"`
	Date            string  `json:"date"`
	ReviewDate      string  `json:"review_date,omitempty"`
	LastModified    string  `json:"last_modified,omitempty"`
	ProductID       string  `json:"product_id"`
	Product         string  `json:"product,omitempty"`
	Rating          float64 `json:"rating"`
	Comments        string  `json:"comments"`
	Public          int     `json:"public,omitempty"`
	VisibleToPublic int     `json:"visible_to_public,omitempty"`
	Heading         string  `json:"heading,omitempty"`
	Verified        int     `json:"verified,omitempty"`
	Response        string  `json:"response,omitempty"`
	Location        string  `json:"location"`
	CustCareOpen    int     `json:"custcareopen"`
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

func WithAsArray(asArray bool) ReviewRequestOption {
	return func(params *ReviewRequestParams) {
		params.AsArray = &asArray
	}
}

func WithLimit(limit int) ReviewRequestOption {
	return func(params *ReviewRequestParams) {
		params.Limit = &limit
	}
}

func WithPage(page int) ReviewRequestOption {
	return func(params *ReviewRequestParams) {
		params.Page = &page
	}
}

func WithFrom(from string) ReviewRequestOption {
	return func(params *ReviewRequestParams) {
		params.From = &from
	}
}

func WithTo(to string) ReviewRequestOption {
	return func(params *ReviewRequestParams) {
		params.To = &to
	}
}

func WithSort(sort string) ReviewRequestOption {
	return func(params *ReviewRequestParams) {
		params.Sort = &sort
	}
}

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

func (client *ShopperApprovedClient) FetchAllProductReviews(params *ReviewRequestParams) (map[string]Review, error) {
	const baseURL = "https://api.shopperapproved.com/products/reviews"

	url := fmt.Sprintf("%s/%d", baseURL, client.SiteID)

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
