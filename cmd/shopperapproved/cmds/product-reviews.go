package cmds

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/shopperapproved/pkg"
)
import "github.com/pkg/errors"

type GetProductReviewCommand struct {
	*cmds.CommandDescription
}

type GetAllProductReviewsCommand struct {
	*cmds.CommandDescription
}

func NewShopperApprovedLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"shopper-approved", "Shopper Approved",
		layers.WithFlags(
			parameters.NewParameterDefinition("siteId", parameters.ParameterTypeInteger, parameters.WithHelp("Shopper Approved site ID"),
				parameters.WithRequired(true)),
			parameters.NewParameterDefinition("accessToken", parameters.ParameterTypeString, parameters.WithHelp("Shopper Approved access token"),
				parameters.WithRequired(true)),
		),
	)
}

func NewGetProductReviewCommand() (*GetProductReviewCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	shopperApprovedLayer, err := NewShopperApprovedLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Shopper Approved parameter layer")
	}

	return &GetProductReviewCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get-reviews",
			cmds.WithShort("Fetch a product review based on product ID"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("productID", parameters.ParameterTypeString, parameters.WithHelp("ID of the product to fetch reviews for")),
				parameters.NewParameterDefinition("limit", parameters.ParameterTypeInteger, parameters.WithHelp("Limit on the number of reviews returned"), parameters.WithDefault(10)),
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

func (c *GetProductReviewCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	params := &pkg.ReviewRequestParams{
		ProductID: ps["productID"].(string),
	}

	if asArray, ok := ps["asArray"]; ok {
		params.AsArray = asArray.(*bool)
	}
	if limit, ok := ps["limit"]; ok {
		params.Limit = limit.(*int)
	}
	if page, ok := ps["page"]; ok {
		params.Page = page.(*int)
	}
	if from, ok := ps["from"]; ok {
		params.From = from.(*string)
	}
	if to, ok := ps["to"]; ok {
		params.To = to.(*string)
	}
	if sort, ok := ps["sort"]; ok {
		params.Sort = sort.(*string)
	}
	if removed, ok := ps["removed"]; ok {
		params.Removed = removed.(*int)
	}

	siteId, ok := ps["siteId"].(int)
	if !ok {
		return errors.New("siteId is required")
	}
	accessToken, ok := ps["accessToken"].(string)
	if !ok {
		return errors.New("accessToken is required")
	}

	client := pkg.ShopperApprovedClient{
		SiteID: siteId,
		Token:  accessToken,
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
				parameters.NewParameterDefinition("productID", parameters.ParameterTypeString, parameters.WithHelp("ID of the product to fetch reviews for")),
				parameters.NewParameterDefinition("limit", parameters.ParameterTypeInteger, parameters.WithHelp("Limit on the number of reviews returned"), parameters.WithDefault(10)),
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
	params := &pkg.ReviewRequestParams{}

	if productID, ok := ps["productID"]; ok {
		params.ProductID = productID.(string)
	}

	if limit, ok := ps["limit"]; ok {
		params.Limit = limit.(*int)
	}
	if page, ok := ps["page"]; ok {
		params.Page = page.(*int)
	}
	if from, ok := ps["from"]; ok {
		params.From = from.(*string)
	}
	if to, ok := ps["to"]; ok {
		params.To = to.(*string)
	}
	if sort, ok := ps["sort"]; ok {
		params.Sort = sort.(*string)
	}
	if removed, ok := ps["removed"]; ok {
		params.Removed = removed.(*int)
	}

	siteId, ok := ps["siteId"].(int)
	if !ok {
		return errors.New("siteId is required")
	}
	accessToken, ok := ps["accessToken"].(string)
	if !ok {
		return errors.New("accessToken is required")
	}

	client := pkg.ShopperApprovedClient{
		SiteID: siteId,
		Token:  accessToken,
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
