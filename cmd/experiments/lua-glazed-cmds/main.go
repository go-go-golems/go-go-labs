package main

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/spf13/cobra"
	"math/rand"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	glazed_middlewares "github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
)

type AnimalListCommand struct {
	*cmds.CommandDescription
}

type AnimalListSettings struct {
	Count int `glazed.parameter:"count"`
}

func NewAnimalListCommand() (*AnimalListCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &AnimalListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"animal-list",
			cmds.WithShort("List random animals"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"count",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Number of animals to list"),
					parameters.WithDefault(5),
				),
			),
			cmds.WithLayersList(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (c *AnimalListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp glazed_middlewares.Processor,
) error {
	s := &AnimalListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	animals := []string{"Lion", "Elephant", "Giraffe", "Zebra", "Monkey", "Penguin", "Kangaroo", "Koala", "Tiger", "Bear"}

	for i := 0; i < s.Count; i++ {
		animalIndex := rand.Intn(len(animals))
		row := types.NewRow(
			types.MRP("id", i+1),
			types.MRP("animal", animals[animalIndex]),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	L := lua.NewState()
	defer L.Close()

	animalListCmd, err := NewAnimalListCommand()
	cobra.CheckErr(err)

	gp := glazed_middlewares.NewTableProcessor(glazed_middlewares.WithPrependTableMiddleware(&table.NullTableMiddleware{}))

	defaultLayer, ok := animalListCmd.Layers.Get(layers.DefaultSlug)
	if !ok {
		panic("default layer not found")
	}
	cobra.CheckErr(err)
	parsedParameters, err := defaultLayer.GetParameterDefinitions().GatherParametersFromMap(map[string]interface{}{
		"count": 5,
	}, true)
	cobra.CheckErr(err)

	defaultParsedLayer, err := layers.NewParsedLayer(defaultLayer, layers.WithParsedParameters(parsedParameters))
	cobra.CheckErr(err)
	parsedLayers := layers.NewParsedLayers(layers.WithParsedLayer(layers.DefaultSlug, defaultParsedLayer))

	ctx := context.Background()

	err = animalListCmd.RunIntoGlazeProcessor(ctx, parsedLayers, gp)
	cobra.CheckErr(err)
	err = gp.Close(ctx)
	cobra.CheckErr(err)

	tbl := gp.Table
	for _, row := range tbl.Rows {
		for _, column := range tbl.Columns {
			v, _ := row.Get(column)
			fmt.Println(column, ":", v)
		}
	}

	// Lua script to create a table
	script := `
		params = {
			name = "John Doe",
			age = 30
		}
	`
	if err := L.DoString(script); err != nil {
		panic(err)
	}

	// Define parameter definitions
	paramDefs := parameters.NewParameterDefinitions(
		parameters.WithParameterDefinitionList([]*parameters.ParameterDefinition{
			parameters.NewParameterDefinition("name", parameters.ParameterTypeString),
			parameters.NewParameterDefinition("age", parameters.ParameterTypeInteger),
		}),
	)

	// Create a parameter layer
	layer, err := layers.NewParameterLayer("user", "User Information",
		layers.WithDescription("Parameters related to user information"),
		layers.WithParameterDefinitions(paramDefs.ToList()...),
	)
	if err != nil {
		panic(err)
	}

	// Create parameter layers and add the created layer
	parameterLayers := layers.NewParameterLayers(
		layers.WithLayers(layer),
	)

	// Create parsedLayers
	parsedLayers = layers.NewParsedLayers()

	luaTable := L.GetGlobal("params").(*lua.LTable)

	// Execute middlewares
	err = middlewares.ExecuteMiddlewares(parameterLayers, parsedLayers,
		ParseLuaTableMiddleware(luaTable, "user"),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Parsed parameters:", parsedLayers.GetDataMap())
}
