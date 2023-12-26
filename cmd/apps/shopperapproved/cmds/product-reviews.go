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

var _ cmds.GlazeCommand = (*GetProductReviewsCommand)(nil)
var _ cmds.GlazeCommand = (*GetAllProductReviewsCommand)(nil)

type GetAllProductReviewsCommand struct {
	*cmds.CommandDescription
}

func NewShopperApprovedLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"shopper-approved", "Shopper Approved",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition("site-id", parameters.ParameterTypeInteger, parameters.WithHelp("Shopper Approved site ID")),
			parameters.NewParameterDefinition("access-token", parameters.ParameterTypeString, parameters.WithHelp("Shopper Approved access token")),
		),
	)
}

type ShopperApprovedSettings struct {
	SiteID      *int    `glazed.parameter:"site-id"`
	AccessToken *string `glazed.parameter:"access-token"`
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

type GetProductReviewsSettings struct {
	ProductID string     `glazed.parameter:"product-id"`
	Limit     *int       `glazed.parameter:"limit"`
	Page      *int       `glazed.parameter:"page"`
	From      *time.Time `glazed.parameter:"from"`
	To        *time.Time `glazed.parameter:"to"`
	Sort      *string    `glazed.parameter:"sort"`
	Removed   *int       `glazed.parameter:"removed"`
}

func (c *GetProductReviewsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &GetProductReviewsSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}
	params := &pkg2.ReviewRequestParams{
		ProductID: s.ProductID,
		Limit:     s.Limit,
		Page:      s.Page,
		From:      s.From,
		To:        s.To,
		Sort:      s.Sort,
		Removed:   s.Removed,
	}

	saSettings := &ShopperApprovedSettings{}
	err = parsedLayers.InitializeStruct("shopper-approved", saSettings)
	if err != nil {
		return err
	}
	client, err := getCredentials(saSettings)
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

func (c *GetAllProductReviewsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &GetProductReviewsSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	params := &pkg2.ReviewRequestParams{
		ProductID: s.ProductID,
		Limit:     s.Limit,
		Page:      s.Page,
		From:      s.From,
		To:        s.To,
		Sort:      s.Sort,
		Removed:   s.Removed,
	}

	saSettings := &ShopperApprovedSettings{}
	err = parsedLayers.InitializeStruct("shopper-approved", saSettings)
	if err != nil {
		return err
	}

	client, err := getCredentials(saSettings)
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

func getCredentials(approvedSettings *ShopperApprovedSettings) (*pkg2.ShopperApprovedClient, error) {
	if approvedSettings.SiteID == nil {
		siteId_ := os.Getenv("SHOPPER_APPROVED_SITE_ID")
		if siteId_ == "" {
			return nil, errors.New("siteId is required")
		}
		siteId, err := strconv.Atoi(siteId_)
		if err != nil {
			return nil, errors.Wrap(err, "siteId is required")
		}
		approvedSettings.SiteID = &siteId
	}
	if approvedSettings.AccessToken == nil {
		accessToken := os.Getenv("SHOPPER_APPROVED_ACCESS_TOKEN")
		if accessToken == "" {
			return nil, errors.New("accessToken is required")
		}
		approvedSettings.AccessToken = &accessToken
	}

	client := &pkg2.ShopperApprovedClient{
		SiteID: *approvedSettings.SiteID,
		Token:  *approvedSettings.AccessToken,
	}
	return client, nil
}
