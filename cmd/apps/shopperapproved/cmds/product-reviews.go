package cmds

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	pkg2 "github.com/go-go-golems/go-go-labs/cmd/apps/shopperapproved/pkg"
	"os"
	"strconv"
	"time"
)
import "github.com/pkg/errors"

type GetProductReviewsCommand struct {
	*cmds.CommandDescription
}

type GetAllProductReviewsCommand struct {
	*cmds.CommandDescription
}

func NewShopperApprovedLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"shopper-approved", "Shopper Approved",
		layers.WithFlags(
			parameters.NewParameterDefinition("site-id", parameters.ParameterTypeInteger, parameters.WithHelp("Shopper Approved site ID")),
			parameters.NewParameterDefinition("access-token", parameters.ParameterTypeString, parameters.WithHelp("Shopper Approved access token")),
		),
	)
}

func NewGetProductReviewsCommand() (*GetProductReviewsCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	shopperApprovedLayer, err := NewShopperApprovedLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Shopper Approved parameter layer")
	}

	return &GetProductReviewsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get-reviews",
			cmds.WithShort("Fetch a product review based on product ID"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("product-id", parameters.ParameterTypeString, parameters.WithHelp("ID of the product to fetch reviews for")),
				parameters.NewParameterDefinition("limit", parameters.ParameterTypeInteger, parameters.WithHelp("Limit on the number of reviews returned"), parameters.WithDefault(100)),
				parameters.NewParameterDefinition("page", parameters.ParameterTypeInteger, parameters.WithHelp("Page number for pagination")),
				parameters.NewParameterDefinition("from", parameters.ParameterTypeDate, parameters.WithHelp("Start date for the query in YYYY-MM-DD format")),
				parameters.NewParameterDefinition("to", parameters.ParameterTypeDate, parameters.WithHelp("End date for the query in YYYY-MM-DD format")),
				parameters.NewParameterDefinition("sort", parameters.ParameterTypeChoice, parameters.WithHelp("Sorting method: newest, oldest, highest, lowest")),
				parameters.NewParameterDefinition("removed", parameters.ParameterTypeInteger, parameters.WithHelp("Include reviews that are removed or not")),
			),
			cmds.WithLayers(shopperApprovedLayer, glazedParameterLayer),
		),
	}, nil
}

func GetAndCast[T any](ps map[string]interface{}, name string, default_ T) (T, bool, error) {
	if val, ok := ps[name]; ok {
		if castedVal, ok := val.(T); ok {
			return castedVal, true, nil
		}
		return default_, false, errors.Errorf("could not cast %s to %T", name, val)
	}
	return default_, false, nil
}

func GetAndCastPtr[T any](ps map[string]interface{}, name string, default_ *T) (*T, bool, error) {
	if val, ok := ps[name]; ok {
		if castedVal, ok := val.(T); ok {
			return &castedVal, true, nil
		}
		return default_, true, errors.Errorf("could not cast %s to %T", name, val)
	}
	return default_, false, nil
}

func (c *GetProductReviewsCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	params := &pkg2.ReviewRequestParams{
		ProductID: ps["product-id"].(string),
	}

	var err error
	params.Limit, _, err = GetAndCastPtr[int](ps, "limit", nil)
	if err != nil {
		return err
	}
	params.Page, _, err = GetAndCastPtr[int](ps, "page", nil)
	if err != nil {
		return err
	}

	params.From, _, err = GetAndCastPtr[time.Time](ps, "from", nil)
	if err != nil {
		return err
	}
	params.To, _, err = GetAndCastPtr[time.Time](ps, "to", nil)
	if err != nil {
		return err
	}
	params.Sort, _, err = GetAndCastPtr[string](ps, "sort", nil)
	if err != nil {
		return err
	}
	params.Removed, _, err = GetAndCastPtr[int](ps, "removed", nil)
	if err != nil {
		return err
	}

	client, err := getCredentials(ps)
	if err != nil {
		return err
	}

	reviews, err := client.FetchReviews(params)
	if err != nil {
		return err
	}

	for _, review := range reviews {
		row := types.NewRowFromStruct(&review, true)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func NewGetAllProductReviewsCommand() (*GetAllProductReviewsCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	shopperApprovedLayer, err := NewShopperApprovedLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Shopper Approved parameter layer")
	}

	return &GetAllProductReviewsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get-all-reviews",
			cmds.WithShort("Fetch all product reviews based on given criteria"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("product-id", parameters.ParameterTypeString, parameters.WithHelp("ID of the product to fetch reviews for")),
				parameters.NewParameterDefinition("limit", parameters.ParameterTypeInteger, parameters.WithHelp("Limit on the number of reviews returned"), parameters.WithDefault(100)),
				parameters.NewParameterDefinition("page", parameters.ParameterTypeInteger, parameters.WithHelp("Page number for pagination")),
				parameters.NewParameterDefinition("from", parameters.ParameterTypeDate, parameters.WithHelp("Start date for the query in YYYY-MM-DD format")),
				parameters.NewParameterDefinition("to", parameters.ParameterTypeDate, parameters.WithHelp("End date for the query in YYYY-MM-DD format")),
				parameters.NewParameterDefinition("sort", parameters.ParameterTypeChoice, parameters.WithHelp("Sorting method: newest, oldest, highest, lowest")),
				parameters.NewParameterDefinition("removed", parameters.ParameterTypeInteger, parameters.WithHelp("Include reviews that are removed or not")),
			),
			cmds.WithLayers(shopperApprovedLayer, glazedParameterLayer),
		),
	}, nil
}

func (c *GetAllProductReviewsCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	params := &pkg2.ReviewRequestParams{}

	var err error
	params.Limit, _, err = GetAndCastPtr[int](ps, "limit", nil)
	if err != nil {
		return err
	}
	params.Page, _, err = GetAndCastPtr[int](ps, "page", nil)
	if err != nil {
		return err
	}

	params.From, _, err = GetAndCastPtr[time.Time](ps, "from", nil)
	if err != nil {
		return err
	}
	params.To, _, err = GetAndCastPtr[time.Time](ps, "to", nil)
	if err != nil {
		return err
	}
	params.Sort, _, err = GetAndCastPtr[string](ps, "sort", nil)
	if err != nil {
		return err
	}
	params.Removed, _, err = GetAndCastPtr[int](ps, "removed", nil)
	if err != nil {
		return err
	}

	client, err := getCredentials(ps)
	if err != nil {
		return err
	}

	reviews, err := client.FetchReviews(params)
	if err != nil {
		return err
	}

	// Add each review as a row
	for _, review := range reviews {
		row := types.NewRowFromStruct(&review, true)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func getCredentials(ps map[string]interface{}) (*pkg2.ShopperApprovedClient, error) {
	siteId, ok := ps["site-id"].(int)
	if !ok {
		siteId_ := os.Getenv("SHOPPER_APPROVED_SITE_ID")
		if siteId_ == "" {
			return nil, errors.New("siteId is required")
		}
		var err error
		siteId, err = strconv.Atoi(siteId_)
		if err != nil {
			return nil, errors.Wrap(err, "siteId is required")
		}
	}
	accessToken, ok := ps["access-token"].(string)
	if !ok {
		accessToken = os.Getenv("SHOPPER_APPROVED_ACCESS_TOKEN")
		if accessToken == "" {
			return nil, errors.New("accessToken is required")
		}
	}

	client := &pkg2.ShopperApprovedClient{
		SiteID: siteId,
		Token:  accessToken,
	}
	return client, nil
}
