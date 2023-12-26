package main

import (
	"context"
	"fmt"
	ahocorasick "github.com/BobuSumisu/aho-corasick"
	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// findAllMarkdownFiles walks the given directory and returns the paths of all markdown files found.
func findAllMarkdownFiles(directory string) ([]string, error) {
	var ret []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".md") {
			ret = append(ret, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ret, err
}

// getFileBaseNames returns the base name of the given file.
func getFileBaseNames(files []string) *orderedmap.OrderedMap[string, string] {
	ret := orderedmap.New[string, string]()
	for _, file := range files {
		base := filepath.Base(file)
		ret.Set(file, strings.TrimSuffix(base, ".md"))
	}
	return ret
}

var rootCmd = &cobra.Command{
	Use:   "note-linker",
	Short: "note-linker is a CLI app to create note links for markdown files",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// reinitialize the logger because we can now parse --log-level and co
		// from the command line flag
		err := clay.InitLogger()
		cobra.CheckErr(err)
	},
}

type ListCmd struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*ListCmd)(nil)

type ListSettings struct {
	Directories []string `glazed.parameter:"directories"`
}

func (l *ListCmd) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ListSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	for _, directory := range s.Directories {
		fileNames, err := findAllMarkdownFiles(directory)
		if err != nil {
			return err
		}
		baseNames := getFileBaseNames(fileNames)
		for pair := baseNames.Oldest(); pair != nil; pair = pair.Next() {
			err = gp.AddRow(
				ctx,
				types.NewRow(
					types.MRP("title", pair.Value),
					types.MRP("file", pair.Key),
				))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func NewListCommand() (*ListCmd, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &ListCmd{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List all note titles that could be linked to"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"directories",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Directories to search for markdown files"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayers(glazedParameterLayer),
		),
	}, nil
}

func buildAhoCorasickTrie(entries []string) *ahocorasick.Trie {
	builder := ahocorasick.NewTrieBuilder()
	builder.AddStrings(entries)
	return builder.Build()
}

func NewLinkNotesCommand() *cobra.Command {
	ret := &cobra.Command{
		Use:   "link-notes",
		Args:  cobra.MinimumNArgs(1),
		Short: "Link notes by adding note links to markdown files, based on the files found in the given directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			directories, _ := cmd.Flags().GetStringSlice("directories")

			titles := []string{}
			for _, directory := range directories {
				fileNames, err := findAllMarkdownFiles(directory)
				if err != nil {
					return err
				}
				baseNames := getFileBaseNames(fileNames)
				for pair := baseNames.Oldest(); pair != nil; pair = pair.Next() {
					titles = append(titles, strings.ToLower(pair.Value))
				}
			}

			trie := buildAhoCorasickTrie(titles)

			for _, file := range args {
				content, err := os.ReadFile(file)
				if err != nil {
					return err
				}

				lowerCaseContent := strings.ToLower(string(content))

				matches := trie.Match([]byte(lowerCaseContent))
				for _, match := range matches {
					pos := int(match.Pos())
					s := match.MatchString()
					originalString := content[pos : pos+len(s)]

					// check if the original string is surrounded by non-word characters
					if pos > 0 && unicode.IsLetter(rune(content[pos-1])) {
						continue
					}
					if pos+len(s) < len(content) && unicode.IsLetter(rune(content[pos+len(s)])) {
						continue
					}

					fmt.Printf("Found \"%s\" at offset %d\n", string(originalString), pos)
				}
			}

			return nil
		},
	}

	ret.Flags().StringSlice("directories", []string{}, "Directories to search for markdown files")

	return ret
}

func initRootCmd() (*help.HelpSystem, error) {
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	err := clay.InitViper("mastoid", rootCmd)
	if err != nil {
		return nil, err
	}
	err = clay.InitLogger()
	if err != nil {
		return nil, err
	}

	return helpSystem, nil
}

func main() {
	_, err := initRootCmd()
	cobra.CheckErr(err)

	listCmd, err := NewListCommand()
	cobra.CheckErr(err)
	command, err := cli.BuildCobraCommandFromGlazeCommand(listCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	linkNotesCmd := NewLinkNotesCommand()
	rootCmd.AddCommand(linkNotesCmd)

	_ = rootCmd.Execute()
}
