package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/spf13/cobra"

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
	habitats := []string{"Savanna", "Jungle", "Forest", "Arctic", "Desert", "Grassland", "Mountains", "Wetlands"}
	diets := []string{"Carnivore", "Herbivore", "Omnivore"}

	for i := 0; i < s.Count; i++ {
		animalIndex := rand.Intn(len(animals))
		habitatIndex := rand.Intn(len(habitats))
		dietIndex := rand.Intn(len(diets))
		weight := rand.Float64()*1000 + 1 // Random weight between 1 and 1001 kg

		row := types.NewRow(
			types.MRP("id", i+1),
			types.MRP("animal", animals[animalIndex]),
			types.MRP("habitat", habitats[habitatIndex]),
			types.MRP("diet", diets[dietIndex]),
			types.MRP("weight_kg", fmt.Sprintf("%.2f", weight)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// PrintGlazedTableInLua converts a Glazed table to a Lua table and prints it using a Lua script
func PrintGlazedTableInLua(glazedTable *types.Table) error {
	L := lua.NewState()
	defer L.Close()

	// Convert Glazed table to Lua table
	luaTable := GlazedTableToLuaTable(L, glazedTable)

	// Set the Lua table as a global variable
	L.SetGlobal("glazed_table", luaTable)

	// Lua script to print the table
	script := `
		function print_table(t, indent)
			indent = indent or ""
			for k, v in pairs(t) do
				if type(v) == "table" then
					print(indent .. tostring(k) .. ":")
					print_table(v, indent .. "  ")
				else
					print(indent .. tostring(k) .. ": " .. tostring(v))
				end
			end
		end

		print("Glazed Table Contents:")
		print_table(glazed_table)
	`

	// Execute the Lua script
	if err := L.DoString(script); err != nil {
		return fmt.Errorf("error executing Lua script: %v", err)
	}

	return nil
}

func main() {
	L := lua.NewState()
	defer L.Close()
	//
	//fmt.Println("Step 1: Run AnimalListCommand")
	//fmt.Println("---")
	//// Step 1: Run AnimalListCommand
	//runAnimalListCommand(L)
	//
	//fmt.Println("\nStep 2: Handle Lua table parsing")
	//fmt.Println("---")
	//// Step 2: Handle Lua table parsing
	//handleLuaTableParsing(L)
	//
	//fmt.Println("\nStep 3: Pass parsed layers to Lua")
	//fmt.Println("---")
	//// Step 3: Pass parsed layers to Lua
	//passParsedLayersToLua(L)
	//
	//fmt.Println("\nStep 4: Test nested Lua table with AnimalListCommand")
	//fmt.Println("---")
	//// Step 4: Test nested Lua table with AnimalListCommand
	//testNestedLuaTableWithAnimalListCommand(L)

	fmt.Println("\nStep 5: Test registered command")
	fmt.Println("---")
	// Step 5: Test registered command
	testRegisteredCommand()
}

func runAnimalListCommand(L *lua.LState) {
	animalListCmd, err := NewAnimalListCommand()
	cobra.CheckErr(err)

	gp := glazed_middlewares.NewTableProcessor(glazed_middlewares.WithTableMiddleware(&table.NullTableMiddleware{}))

	parsedLayers := createParsedLayers(animalListCmd)

	ctx := context.Background()

	err = animalListCmd.RunIntoGlazeProcessor(ctx, parsedLayers, gp)
	cobra.CheckErr(err)
	err = gp.Close(ctx)
	cobra.CheckErr(err)

	if err := PrintGlazedTableInLua(gp.Table); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func createParsedLayers(cmd *AnimalListCommand) *layers.ParsedLayers {
	defaultLayer, ok := cmd.Layers.Get(layers.DefaultSlug)
	if !ok {
		panic("default layer not found")
	}
	parsedParameters, err := defaultLayer.GetParameterDefinitions().GatherParametersFromMap(map[string]interface{}{
		"count": 5,
	}, true)
	cobra.CheckErr(err)

	defaultParsedLayer, err := layers.NewParsedLayer(defaultLayer, layers.WithParsedParameters(parsedParameters))
	cobra.CheckErr(err)
	return layers.NewParsedLayers(layers.WithParsedLayer(layers.DefaultSlug, defaultParsedLayer))
}

func handleLuaTableParsing(L *lua.LState) {
	script := `
		params = {
			name = "John Doe",
			age = 30
		}
	`
	if err := L.DoString(script); err != nil {
		panic(err)
	}

	paramDefs := parameters.NewParameterDefinitions(
		parameters.WithParameterDefinitionList([]*parameters.ParameterDefinition{
			parameters.NewParameterDefinition("name", parameters.ParameterTypeString),
			parameters.NewParameterDefinition("age", parameters.ParameterTypeInteger),
		}),
	)

	layer, err := layers.NewParameterLayer("user", "User Information",
		layers.WithDescription("Parameters related to user information"),
		layers.WithParameterDefinitions(paramDefs.ToList()...),
	)
	if err != nil {
		panic(err)
	}

	parameterLayers := layers.NewParameterLayers(
		layers.WithLayers(layer),
	)

	parsedLayers := layers.NewParsedLayers()

	luaTable := L.GetGlobal("params").(*lua.LTable)

	err = middlewares.ExecuteMiddlewares(parameterLayers, parsedLayers,
		ParseLuaTableMiddleware(luaTable, "user"),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Parsed parameters:", parsedLayers.GetDataMap())
}

func passParsedLayersToLua(L *lua.LState) {
	// Assuming you have parsedLayers available from the previous step
	parsedLayers := createDemoParsedLayers()
	luaTable := ParsedLayersToLuaTable(L, parsedLayers)

	L.SetGlobal("parsed_layers", luaTable)

	script := `
    for layer_name, layer_data in pairs(parsed_layers) do
        print("Layer: " .. layer_name)
        for param_name, param_value in pairs(layer_data) do
            print("  " .. param_name .. ": " .. tostring(param_value))
        end
    end
`

	if err := L.DoString(script); err != nil {
		fmt.Printf("Error executing Lua script: %v\n", err)
	}

	cmd, _ := NewAnimalListCommand()

	animalParsedLayers := createParsedLayers(cmd)

	luaTable = ParsedLayersToLuaTable(L, animalParsedLayers)

	L.SetGlobal("parsed_layers", luaTable)

	if err := L.DoString(script); err != nil {
		fmt.Printf("Error executing Lua script: %v\n", err)
	}
}

func testNestedLuaTableWithAnimalListCommand(L *lua.LState) {
	animalListCmd, err := NewAnimalListCommand()
	cobra.CheckErr(err)

	// Create a nested Lua table
	script := `
		params = {
			default = {
				count = 3
			},
			glazed = {
				fields = {"animal", "diet"}
			}
		}
	`
	if err := L.DoString(script); err != nil {
		panic(err)
	}

	luaTable := L.GetGlobal("params").(*lua.LTable)

	// Create parsed layers
	parsedLayers := layers.NewParsedLayers()

	// Define middlewares
	middlewares_ := []middlewares.Middleware{
		// Parse from Lua table (highest priority)
		ParseNestedLuaTableMiddleware(luaTable),
		// Set defaults (lowest priority)
		middlewares.SetFromDefaults(parameters.WithParseStepSource("defaults")),
	}

	// Execute middlewares
	err = middlewares.ExecuteMiddlewares(animalListCmd.Layers, parsedLayers, middlewares_...)
	if err != nil {
		panic(err)
	}

	glazedLayer, ok := parsedLayers.Get(settings.GlazedSlug)
	if !ok {
		panic("Glazed layer not found")
	}
	gp, err := settings.SetupTableProcessor(glazedLayer, glazed_middlewares.WithTableMiddleware(&table.NullTableMiddleware{}))
	cobra.CheckErr(err)

	ctx := context.Background()

	// Run the command with the parsed layers
	err = animalListCmd.RunIntoGlazeProcessor(ctx, parsedLayers, gp)
	cobra.CheckErr(err)
	err = gp.Close(ctx)
	cobra.CheckErr(err)

	fmt.Println("\nTesting AnimalListCommand with nested Lua table:")
	if err := PrintGlazedTableInLua(gp.Table); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func testRegisteredCommand() {
	L := lua.NewState()
	defer L.Close()

	// Create and register your GlazeCommands
	animalListCmd, _ := NewAnimalListCommand()
	RegisterGlazedCommand(L, animalListCmd)

	// Run a Lua script that uses the registered command
	script := `
		local params = {
			default = {
				count = 3
			},
			glazed = {
				fields = {"animal", "diet"}
			}
		}
		local result = animal_list(params)
		for i, row in ipairs(result) do
			print(string.format("Animal %d: %s, Diet: %s", i, row.animal, row.diet))
		end

		-- You can also access parameter information
		print("Parameters for animal_list command:")
		for name, info in pairs(animal_list_params) do
			print(string.format("  %s (%s): %s", name, info.type, info.description))
		end
	`
	if err := L.DoString(script); err != nil {
		fmt.Printf("Error executing Lua script: %v\n", err)
	}
}
