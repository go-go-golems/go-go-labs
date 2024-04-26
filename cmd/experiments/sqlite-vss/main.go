package main

import (
	"context"
	"fmt"
	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sqlite-vss/cmds"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sqlite-vss/pkg"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"time"

	_ "github.com/asg017/sqlite-vss/bindings/go"
	_ "github.com/mattn/go-sqlite3"
)

// #cgo LDFLAGS: -L../../../thirdparty/sqlite-vss-libs/ -Wl,-undefined,dynamic_lookup
import "C"

func createRootCmd() *cobra.Command {
	helpSystem := help.NewHelpSystem()

	rootCmd := &cobra.Command{
		Use:   "sqlite-vss",
		Short: "Play with sqlite VSS",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			err := clay.InitLogger()
			cobra.CheckErr(err)
		},
	}

	helpSystem.SetupCobraRootCommand(rootCmd)

	err := clay.InitViper("sqlite-vss", rootCmd)
	cobra.CheckErr(err)
	err = clay.InitLogger()
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

	initDocumentCommand, err := cmds.NewIndexDocumentCommand(e)
	cobra.CheckErr(err)
	initDocumentCmd, err := cli.BuildCobraCommandFromGlazeCommand(initDocumentCommand)
	cobra.CheckErr(err)
	rootCmd.AddCommand(initDocumentCmd)

	ctx := context.Background()

	err = e.IndexDocument(ctx, "A big headline", "Today's news is horrible.", time.Now())
	if err != nil {
		log.Fatal().Err(err).Msg("could not index document")
	}

	err = e.Search(ctx, "what is going on?")
	if err != nil {
		log.Fatal().Err(err).Msg("could not search")
	}

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
