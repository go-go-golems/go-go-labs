package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type YAMLCommandLoader struct{}

func NewYAMLCommandLoader() *YAMLCommandLoader {
	return &YAMLCommandLoader{}
}

type YAMLCommandDescription struct {
	Name      string                            `yaml:"name"`
	Short     string                            `yaml:"short"`
	Long      string                            `yaml:"long,omitempty"`
	Flags     []*parameters.ParameterDefinition `yaml:"flags,omitempty"`
	Arguments []*parameters.ParameterDefinition `yaml:"arguments,omitempty"`
}

func (l *YAMLCommandLoader) LoadCommands(
	f fs.FS,
	entryName string,
	options []cmds.CommandDescriptionOption,
	aliasOptions []alias.Option,
) ([]cmds.Command, error) {
	file, err := f.Open(entryName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var yamlCmd YAMLCommandDescription
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&yamlCmd)
	if err != nil {
		return nil, err
	}

	cmdOptions := []cmds.CommandDescriptionOption{
		cmds.WithShort(yamlCmd.Short),
		cmds.WithLong(yamlCmd.Long),
		cmds.WithFlags(yamlCmd.Flags...),
		cmds.WithArguments(yamlCmd.Arguments...),
	}
	cmdOptions = append(cmdOptions, options...)

	cmd := cmds.NewCommandDescription(yamlCmd.Name, cmdOptions...)

	glazedCmd := cmds.NewCommandFromDescription(cmd, func(
		ctx context.Context,
		parsedLayers *layers.ParsedLayers,
		gp middlewares.Processor,
	) error {
		type GenerateSettings struct {
			Count   int    `glazed.parameter:"count"`
			Verbose bool   `glazed.parameter:"verbose"`
			Prefix  string `glazed.parameter:"prefix"`
		}

		settings := &GenerateSettings{}
		if err := parsedLayers.InitializeStruct("default", settings); err != nil {
			return err
		}

		for i := 1; i <= settings.Count; i++ {
			user := types.NewRow(
				types.MRP("id", i),
				types.MRP("name", settings.Prefix+"-"+strconv.Itoa(i)),
				types.MRP("email", "user"+strconv.Itoa(i)+"@example.com"),
			)

			if settings.Verbose {
				user.Set("debug", "Verbose mode enabled")
			}

			if err := gp.AddRow(ctx, user); err != nil {
				return err
			}
		}

		return nil
	})

	return []cmds.Command{glazedCmd}, nil
}

func (l *YAMLCommandLoader) IsFileSupported(f fs.FS, fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	return ext == ".yaml" || ext == ".yml"
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "glazed-cli",
		Short: "A CLI application using Glazed",
	}

	// Initialize the help system
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Create a YAML loader
	yamlLoader := NewYAMLCommandLoader()

	// Load commands from YAML
	commands, err := yamlLoader.LoadCommands(os.DirFS("."), "commands.yaml", nil, nil)
	if err != nil {
		fmt.Println("Error loading commands:", err)
		os.Exit(1)
	}

	// Add loaded commands to root
	for _, cmd := range commands {
		cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
		if err != nil {
			fmt.Println("Error building Cobra command:", err)
			os.Exit(1)
		}
		rootCmd.AddCommand(cobraCmd)
	}

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
