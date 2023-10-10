package main

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-labs/cmd/shopperapproved/cmds"
	"github.com/spf13/cobra"
)

type ShopperApprovedConfig struct {
	SiteID string
	Token  string
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "shopperApproved",
		Short: "Shopper Approved fetcher",
	}

	helpSystem := help.NewHelpSystem()

	helpSystem.SetupCobraRootCommand(rootCmd)

	productCmd := &cobra.Command{
		Use:   "products",
		Short: "Shopper Approved products",
	}

	getProductReviewsCommand, err := cmds.NewGetProductReviewCommand()
	cobra.CheckErr(err)
	cobraCommand, err := cli.BuildCobraCommandFromGlazeCommand(getProductReviewsCommand)
	cobra.CheckErr(err)
	productCmd.AddCommand(cobraCommand)

	getAllProductReviewsCommand, err := cmds.NewGetAllProductReviewsCommand()
	cobra.CheckErr(err)
	cobraCommand, err = cli.BuildCobraCommandFromGlazeCommand(getAllProductReviewsCommand)
	cobra.CheckErr(err)
	productCmd.AddCommand(cobraCommand)

	rootCmd.AddCommand(productCmd)

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
