// nolint
package main

import (
	"fmt"

	_ "github.com/asg017/sqlite-vss/bindings/go"
	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sqlite-vss/cmds"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sqlite-vss/pkg"
	geppetto_cmds "github.com/go-go-golems/pinocchio/pkg/cmds"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// #cgo LDFLAGS: -L../../../thirdparty/sqlite-vss-libs/ -Wl,-undefined,dynamic_lookup
import "C"

func createRootCmd() *cobra.Command {
	helpSystem := help.NewHelpSystem()

	rootCmd := &cobra.Command{
		Use:   "sqlite-vss",
		Short: "Play with sqlite VSS",
	}

	helpSystem.SetupCobraRootCommand(rootCmd)

	err := clay.InitViper("sqlite-vss", rootCmd)
	cobra.CheckErr(err)

	return rootCmd
}

func main() {
	rootCmd := createRootCmd()

	e, err := pkg.NewEmbedder("file:test.db")
	if err != nil {
		log.Fatal().Err(err).Msg("could not create embedder")
	}
	defer e.Close()

	fmt.Println(e.VSSVersion())
	err = e.Init()
	if err != nil {
		log.Fatal().Err(err).Msg("could not initialize embedder")
	}

	//ctx := context.Background()

	// load glaze help system
	//helpSystem := help.NewHelpSystem()
	//err = doc.AddDocToHelpSystem(helpSystem)
	//cobra.CheckErr(err)
	//err = e.IndexHelpSystem(ctx, helpSystem)
	//cobra.CheckErr(err)
	//_ = ctx

	initDocumentCommand, err := cmds.NewIndexDocumentCommand(e)
	cobra.CheckErr(err)
	initDocumentCmd, err := cli.BuildCobraCommandFromGlazeCommand(initDocumentCommand)
	cobra.CheckErr(err)
	rootCmd.AddCommand(initDocumentCmd)

	searchCommand, err := cmds.NewSearchCommand(e)
	cobra.CheckErr(err)
	searchCmd, err := geppetto_cmds.BuildCobraCommandWithGeppettoMiddlewares(searchCommand)
	cobra.CheckErr(err)
	rootCmd.AddCommand(searchCmd)

	answerCommand, err := cmds.NewAnswerQuestionCommand()
	cobra.CheckErr(err)
	answerCmd, err := geppetto_cmds.BuildCobraCommandWithGeppettoMiddlewares(answerCommand)
	cobra.CheckErr(err)
	rootCmd.AddCommand(answerCmd)

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
