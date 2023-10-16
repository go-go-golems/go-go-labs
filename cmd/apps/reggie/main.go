package main

import (
	"fmt"
	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "reggie",
	Short: "reggie is a tool to run a set of regexps against and document",
}

func main() {
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	err := clay.InitViper("reggie", rootCmd)
	if err != nil {
		panic(err)
	}
	err = clay.InitLogger()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error initializing logger: %s\n", err)
		os.Exit(1)
	}

	_ = rootCmd.Execute()
}
