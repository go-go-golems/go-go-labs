package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/dave/jennifer/jen"
	cmds2 "github.com/go-go-golems/sqleton/pkg/cmds"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
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
			fmt.Printf("File: %s\n", path)
			fmt.Println(file.GoString())
		}

		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

type FileResult struct {
	Path string
	File *jen.File
}

func processFiles(ctx context.Context, files []string) (map[string]*jen.File, error) {
	// Use errgroup with context cancellation
	g, ctx := errgroup.WithContext(ctx)
	resultChan := make(chan FileResult, len(files))

	defer close(resultChan)

	s := &SqlCommandCodeGenerator{
		PackageName: "main",
		SplitFiles:  false,
	}

	for _, fileName := range files {
		// passing fileName to avoid the go routine range variable issue
		file := fileName
		g.Go(func() error {
			// Check for context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			psYaml, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}

			loader := &cmds2.SqlCommandLoader{
				DBConnectionFactory: nil,
			}

			// create reader from psYaml
			r := bytes.NewReader(psYaml)
			cmds_, err := loader.LoadCommandFromYAML(r)
			if err != nil {
				return err
			}
			if len(cmds_) != 1 {
				return fmt.Errorf("expected exactly one command, got %d", len(cmds_))
			}
			cmd := cmds_[0].(*cmds2.SqlCommand)

			f := s.GenerateCommandCode(cmd)
			// send the result to the result channel
			resultChan <- FileResult{Path: file, File: f}

			return nil
		})
	}

	// Wait for all files to be processed
	if err := g.Wait(); err != nil {
		return nil, err
	}

	results := make(map[string]*jen.File)
	for fr := range resultChan {
		results[fr.Path] = fr.File
	}

	return results, nil
}
