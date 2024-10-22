package main

import (
	"context"
	"fmt"
	"os"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"

	"github.com/spf13/cobra"
)

func initRootCmd(rootCmd *cobra.Command) (*help.HelpSystem, error) {
	helpSystem := help.NewHelpSystem()
	var err error
	// err := doc.AddDocToHelpSystem(helpSystem)
	cobra.CheckErr(err)

	helpSystem.SetupCobraRootCommand(rootCmd)

	err = clay.InitViper("catter", rootCmd)
	cobra.CheckErr(err)
	err = clay.InitLogger()
	cobra.CheckErr(err)

	return helpSystem, nil
}

func main() {
	ctx := context.Background()

	catterCmd, err := NewCatterCommand()
	cobra.CheckErr(err)

	cobraCmd, err := cli.BuildCobraCommandFromCommand(catterCmd,
		cli.WithCobraMiddlewaresFunc(getMiddlewares),
	)
	cobra.CheckErr(err)

	_, err = initRootCmd(cobraCmd)
	cobra.CheckErr(err)

	if err := cobraCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

func getMiddlewares(
	commandSettings *cli.GlazedCommandSettings,
	cmd *cobra.Command,
	args []string,
) ([]middlewares.Middleware, error) {
	return []middlewares.Middleware{
		middlewares.ParseFromCobraCommand(cmd),
		middlewares.GatherArguments(args),
		middlewares.GatherSpecificFlagsFromViper(
			[]string{"filter-profile"},
			parameters.WithParseStepSource("viper"),
		),
		middlewares.SetFromDefaults(),
	}, nil
}
