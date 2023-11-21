package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/dave/jennifer/jen"
	cmds2 "github.com/go-go-golems/sqleton/pkg/cmds"
	"github.com/go-go-golems/sqleton/pkg/codegen"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var rootCmd = &cobra.Command{
	Use:   "program [file...]",
	Short: "A program to load YAML from files and pass the parsed command to GenerateCommandCode",
	Long:  `This is a program to load YAML files passed in as parameters and pass each parsed command to GenerateCommandCode`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := processFiles(context.Background(), args)
		if err != nil {
			return err
		}

		for path, file := range res {
			s := file.GoString()
			// store in path.go after removing .yaml
			p, _ := strings.CutSuffix(path, ".yaml")
			p = p + ".go"

			fmt.Println("Writing to", p)
			_ = os.WriteFile(p, []byte(s), 0644)
		}

		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func processFiles(ctx context.Context, files []string) (map[string]*jen.File, error) {
	s := &codegen.SqlCommandCodeGenerator{
		PackageName: "main",
	}

	results := make(map[string]*jen.File)

	for _, fileName := range files {
		// passing fileName to avoid the go routine range variable issue
		file := fileName
		// Check for context cancellation

		psYaml, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		loader := &cmds2.SqlCommandLoader{
			DBConnectionFactory: nil,
		}

		// create reader from psYaml
		r := bytes.NewReader(psYaml)
		cmds_, err := loader.LoadCommandFromYAML(r)
		if err != nil {
			return nil, err
		}
		if len(cmds_) != 1 {
			return nil, fmt.Errorf("expected exactly one command, got %d", len(cmds_))
		}
		cmd := cmds_[0].(*cmds2.SqlCommand)

		f, err := s.GenerateCommandCode(cmd)
		if err != nil {
			return nil, err
		}
		results[file] = f
	}

	return results, nil
}
