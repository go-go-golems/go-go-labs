package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/middlewares/table"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	glazed_middlewares "github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	lua "github.com/yuin/gopher-lua"
)

// ParseNestedLuaTableToParsedLayers parses a nested Lua table into ParsedLayers
func ParseNestedLuaTableToParsedLayers(luaTable *lua.LTable, parameterLayers *layers.ParameterLayers) (*layers.ParsedLayers, error) {
	parsedLayers := layers.NewParsedLayers()
	var conversionErrors []string

	luaTable.ForEach(func(key, value lua.LValue) {
		if keyStr, ok := key.(lua.LString); ok {
			layerName := string(keyStr)
			layer, ok := parameterLayers.Get(layerName)
			if !ok {
				conversionErrors = append(conversionErrors, fmt.Sprintf("layer '%s' not found", layerName))
				return
			}

			if nestedTable, ok := value.(*lua.LTable); ok {
				parsedLayer, err := ParseLuaTableToLayer(nestedTable, layer)
				if err != nil {
					conversionErrors = append(conversionErrors, err.Error())
				} else {
					parsedLayers.Set(layerName, parsedLayer)
				}
			} else {
				conversionErrors = append(conversionErrors, fmt.Sprintf("invalid value for layer '%s': expected table, got %s", layerName, value.Type()))
			}
		}
	})

	if len(conversionErrors) > 0 {
		return nil, fmt.Errorf("parameter conversion errors: %s", strings.Join(conversionErrors, "; "))
	}

	return parsedLayers, nil
}

// ParseLuaTableToLayer parses a Lua table into a ParsedLayer
func ParseLuaTableToLayer(luaTable *lua.LTable, layer layers.ParameterLayer) (*layers.ParsedLayer, error) {
	params := make(map[string]interface{})
	var conversionErrors []string

	luaTable.ForEach(func(key, value lua.LValue) {
		if keyStr, ok := key.(lua.LString); ok {
			paramDef, _ := layer.GetParameterDefinitions().Get(string(keyStr))
			if paramDef != nil {
				convertedValue, err := ParseParameterFromLua(value, paramDef)
				if err != nil {
					conversionErrors = append(conversionErrors, err.Error())
				} else {
					params[string(keyStr)] = convertedValue
				}
			}
		}
	})

	if len(conversionErrors) > 0 {
		return nil, fmt.Errorf("parameter conversion errors: %s", strings.Join(conversionErrors, "; "))
	}

	// Parse parameters using the layer's definitions
	parsedParams, err := layer.GetParameterDefinitions().GatherParametersFromMap(params, true, parameters.WithParseStepSource("lua"))
	if err != nil {
		return nil, err
	}

	// Create a parsed layer
	return layers.NewParsedLayer(layer, layers.WithParsedParameters(parsedParams))
}

// Middleware to parse Lua table into a ParsedLayer
func ParseLuaTableMiddleware(luaTable *lua.LTable, layerName string) middlewares.Middleware {
	return func(next middlewares.HandlerFunc) middlewares.HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			// Look up the specific layer
			layer, ok := layers_.Get(layerName)
			if !ok {
				return fmt.Errorf("layer '%s' not found", layerName)
			}

			parsedLayer, err := ParseLuaTableToLayer(luaTable, layer)
			if err != nil {
				return err
			}

			parsedLayers.GetOrCreate(layer).MergeParameters(parsedLayer)
			return next(layers_, parsedLayers)
		}
	}
}

// Middleware to parse nested Lua table into ParsedLayers
func ParseNestedLuaTableMiddleware(luaTable *lua.LTable) middlewares.Middleware {
	return func(next middlewares.HandlerFunc) middlewares.HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			newParsedLayers, err := ParseNestedLuaTableToParsedLayers(luaTable, layers_)
			if err != nil {
				return err
			}

			// Merge the new parsed layers with the existing ones
			parsedLayers.Merge(newParsedLayers)

			return next(layers_, parsedLayers)
		}
	}
}

func ParseParameterFromLua(value lua.LValue, paramDef *parameters.ParameterDefinition) (interface{}, error) {
	switch paramDef.Type {
	case parameters.ParameterTypeString:
		if v, ok := value.(lua.LString); ok {
			return string(v), nil
		}
		return nil, fmt.Errorf("invalid type for parameter '%s': expected string, got %s", paramDef.Name, value.Type())
	case parameters.ParameterTypeInteger:
		if v, ok := value.(lua.LNumber); ok {
			return int(v), nil
		}
		return nil, fmt.Errorf("invalid type for parameter '%s': expected integer, got %s", paramDef.Name, value.Type())
	case parameters.ParameterTypeFloat:
		if v, ok := value.(lua.LNumber); ok {
			return float64(v), nil
		}
		return nil, fmt.Errorf("invalid type for parameter '%s': expected float, got %s", paramDef.Name, value.Type())
	case parameters.ParameterTypeBool:
		if v, ok := value.(lua.LBool); ok {
			return bool(v), nil
		}
		return nil, fmt.Errorf("invalid type for parameter '%s': expected boolean, got %s", paramDef.Name, value.Type())
	case parameters.ParameterTypeStringList:
		if tbl, ok := value.(*lua.LTable); ok {
			var list []string
			var invalidTypes []string
			tbl.ForEach(func(_, v lua.LValue) {
				if str, ok := v.(lua.LString); ok {
					list = append(list, string(str))
				} else {
					invalidTypes = append(invalidTypes, v.Type().String())
				}
			})
			if len(invalidTypes) > 0 {
				return nil, fmt.Errorf("invalid types in string list for parameter '%s': %v", paramDef.Name, invalidTypes)
			}
			return list, nil
		}
		return nil, fmt.Errorf("invalid type for parameter '%s': expected table (string list), got %s", paramDef.Name, value.Type())
	case parameters.ParameterTypeIntegerList:
		if tbl, ok := value.(*lua.LTable); ok {
			var list []int
			var invalidTypes []string
			tbl.ForEach(func(_, v lua.LValue) {
				if num, ok := v.(lua.LNumber); ok {
					list = append(list, int(num))
				} else {
					invalidTypes = append(invalidTypes, v.Type().String())
				}
			})
			if len(invalidTypes) > 0 {
				return nil, fmt.Errorf("invalid types in integer list for parameter '%s': %v", paramDef.Name, invalidTypes)
			}
			return list, nil
		}
		return nil, fmt.Errorf("invalid type for parameter '%s': expected table (integer list), got %s", paramDef.Name, value.Type())
	case parameters.ParameterTypeFloatList:
		if tbl, ok := value.(*lua.LTable); ok {
			var list []float64
			var invalidTypes []string
			tbl.ForEach(func(_, v lua.LValue) {
				if num, ok := v.(lua.LNumber); ok {
					list = append(list, float64(num))
				} else {
					invalidTypes = append(invalidTypes, v.Type().String())
				}
			})
			if len(invalidTypes) > 0 {
				return nil, fmt.Errorf("invalid types in float list for parameter '%s': %v", paramDef.Name, invalidTypes)
			}
			return list, nil
		}
		return nil, fmt.Errorf("invalid type for parameter '%s': expected table (float list), got %s", paramDef.Name, value.Type())
	case parameters.ParameterTypeChoice:
		if v, ok := value.(lua.LString); ok {
			choice := string(v)
			for _, c := range paramDef.Choices {
				if c == choice {
					return choice, nil
				}
			}
			return nil, fmt.Errorf("invalid choice '%s' for parameter '%s'", choice, paramDef.Name)
		}
		return nil, fmt.Errorf("invalid type for parameter '%s': expected string (choice), got %s", paramDef.Name, value.Type())
	case parameters.ParameterTypeChoiceList:
		if tbl, ok := value.(*lua.LTable); ok {
			var choices []string
			var invalidChoices []string
			var invalidTypes []string
			tbl.ForEach(func(_, v lua.LValue) {
				if str, ok := v.(lua.LString); ok {
					choice := string(str)
					valid := false
					for _, c := range paramDef.Choices {
						if c == choice {
							choices = append(choices, choice)
							valid = true
							break
						}
					}
					if !valid {
						invalidChoices = append(invalidChoices, choice)
					}
				} else {
					invalidTypes = append(invalidTypes, v.Type().String())
				}
			})
			if len(invalidTypes) > 0 {
				return nil, fmt.Errorf("invalid types in choice list for parameter '%s': %v", paramDef.Name, invalidTypes)
			}
			if len(invalidChoices) > 0 {
				return nil, fmt.Errorf("invalid choices %v for parameter '%s'", invalidChoices, paramDef.Name)
			}
			return choices, nil
		}
		return nil, fmt.Errorf("invalid type for parameter '%s': expected table (choice list), got %s", paramDef.Name, value.Type())
	case parameters.ParameterTypeDate:
		if v, ok := value.(lua.LString); ok {
			parsedDate, err := parameters.ParseDate(string(v))
			if err == nil {
				return parsedDate, nil
			}
			return nil, fmt.Errorf("invalid date '%s' for parameter '%s': %v", v, paramDef.Name, err)
		}
		return nil, fmt.Errorf("invalid type for parameter '%s': expected string (date), got %s", paramDef.Name, value.Type())
	case parameters.ParameterTypeKeyValue:
		if tbl, ok := value.(*lua.LTable); ok {
			keyValue := make(map[string]interface{})
			tbl.ForEach(func(k, v lua.LValue) {
				if key, ok := k.(lua.LString); ok {
					keyValue[string(key)] = LuaValueToInterface(v)
				}
			})
			return keyValue, nil
		}
		return nil, fmt.Errorf("invalid type for parameter '%s': expected table (key-value), got %s", paramDef.Name, value.Type())
	case parameters.ParameterTypeFile,
		parameters.ParameterTypeFileList,
		parameters.ParameterTypeObjectListFromFile,
		parameters.ParameterTypeObjectListFromFiles,
		parameters.ParameterTypeObjectFromFile,
		parameters.ParameterTypeStringFromFile,
		parameters.ParameterTypeStringFromFiles,
		parameters.ParameterTypeStringListFromFile,
		parameters.ParameterTypeStringListFromFiles:
		return nil, fmt.Errorf("parameter type '%s' for '%s' is not implemented for Lua conversion", paramDef.Type, paramDef.Name)
	}
	return nil, fmt.Errorf("unsupported parameter type '%s' for '%s'", paramDef.Type, paramDef.Name)
}

// LuaValueToInterface converts a Lua value to a Go interface{}
func LuaValueToInterface(value lua.LValue) interface{} {
	switch v := value.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		maxn := v.MaxN()
		if maxn == 0 { // Table is a map
			result := make(map[string]interface{})
			v.ForEach(func(key, value lua.LValue) {
				result[key.String()] = LuaValueToInterface(value)
			})
			return result
		} else { // Table is an array
			result := make([]interface{}, 0, maxn)
			for i := 1; i <= maxn; i++ {
				result = append(result, LuaValueToInterface(v.RawGetInt(i)))
			}
			return result
		}
	default:
		return v.String()
	}
}

// GlazedTableToLuaTable converts a Glazed table to a Lua table
func GlazedTableToLuaTable(L *lua.LState, glazedTable *types.Table) *lua.LTable {
	luaTable := L.CreateTable(len(glazedTable.Rows), 0)

	for i, row := range glazedTable.Rows {
		rowTable := L.CreateTable(0, len(glazedTable.Columns))
		for _, col := range glazedTable.Columns {
			value, ok := row.Get(col)
			if !ok {
				continue
			}
			rowTable.RawSetString(col, InterfaceToLuaValue(L, value))
		}
		luaTable.RawSetInt(i+1, rowTable)
	}

	return luaTable
}

// InterfaceToLuaValue converts a Go interface{} to a Lua value
func InterfaceToLuaValue(L *lua.LState, value interface{}) lua.LValue {
	switch v := value.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(v)
	case int:
		return lua.LNumber(v)
	case int64:
		return lua.LNumber(v)
	case float64:
		return lua.LNumber(v)
	case string:
		return lua.LString(v)
	case []interface{}:
		table := L.CreateTable(len(v), 0)
		for i, item := range v {
			table.RawSetInt(i+1, InterfaceToLuaValue(L, item))
		}
		return table
	case map[string]interface{}:
		table := L.CreateTable(0, len(v))
		for key, item := range v {
			table.RawSetString(key, InterfaceToLuaValue(L, item))
		}
		return table
	default:
		return lua.LString(fmt.Sprintf("%v", v))
	}
}

// ParsedLayerToLuaTable converts a ParsedLayer to a Lua table
func ParsedLayerToLuaTable(L *lua.LState, parsedLayer *layers.ParsedLayer) *lua.LTable {
	luaTable := L.CreateTable(0, len(parsedLayer.Parameters.ToMap()))

	parsedLayer.Parameters.ForEach(func(name string, param *parameters.ParsedParameter) {
		luaTable.RawSetString(name, InterfaceToLuaValue(L, param.Value))
	})

	return luaTable
}

// ParsedLayersToLuaTable converts ParsedLayers to a nested Lua table
func ParsedLayersToLuaTable(L *lua.LState, parsedLayers *layers.ParsedLayers) *lua.LTable {
	luaTable := L.CreateTable(0, parsedLayers.Len())

	parsedLayers.ForEach(func(layerName string, parsedLayer *layers.ParsedLayer) {
		layerTable := ParsedLayerToLuaTable(L, parsedLayer)
		luaTable.RawSetString(layerName, layerTable)
	})

	return luaTable
}

// CallGlazedCommandFromLua executes a GlazeCommand with parameters from a Lua table
func CallGlazedCommandFromLua(L *lua.LState, cmd cmds.GlazeCommand, luaTable *lua.LTable) (*types.Table, error) {
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
	err := middlewares.ExecuteMiddlewares(cmd.Description().Layers, parsedLayers, middlewares_...)
	if err != nil {
		return nil, fmt.Errorf("error executing middlewares: %v", err)
	}

	glazedLayer, ok := parsedLayers.Get(settings.GlazedSlug)
	if !ok {
		return nil, fmt.Errorf("glazed layer not found")
	}
	gp, err := settings.SetupTableProcessor(glazedLayer, glazed_middlewares.WithTableMiddleware(&table.NullTableMiddleware{}))
	if err != nil {
		return nil, fmt.Errorf("error setting up table processor: %v", err)
	}

	ctx := context.Background()

	// Run the command with the parsed layers
	err = cmd.RunIntoGlazeProcessor(ctx, parsedLayers, gp)
	if err != nil {
		return nil, fmt.Errorf("error running command: %v", err)
	}
	err = gp.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("error closing processor: %v", err)
	}

	return gp.Table, nil
}

// LuaCallGlazedCommand is a Lua-callable wrapper for CallGlazedCommandFromLua
func LuaCallGlazedCommand(L *lua.LState) int {
	// Get the GlazeCommand from the first argument (userdata)
	cmdUD := L.CheckUserData(1)
	cmd, ok := cmdUD.Value.(cmds.GlazeCommand)
	if !ok {
		L.ArgError(1, "GlazeCommand expected")
		return 0
	}

	// Get the Lua table from the second argument
	luaTable := L.CheckTable(2)

	// Call the Go function
	result, err := CallGlazedCommandFromLua(L, cmd, luaTable)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert the result to a Lua table
	luaResult := GlazedTableToLuaTable(L, result)
	L.Push(luaResult)
	return 1
}

// RegisterGlazedCommand registers a GlazeCommand in the Lua state
func RegisterGlazedCommand(L *lua.LState, cmd cmds.GlazeCommand) {
	desc := cmd.Description()
	name := desc.Name

	// Create a new function that wraps the command
	fn := L.NewFunction(func(L *lua.LState) int {
		// Get the Lua table from the first argument
		luaTable := L.CheckTable(1)

		// Call the Go function
		result, err := CallGlazedCommandFromLua(L, cmd, luaTable)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert the result to a Lua table
		luaResult := GlazedTableToLuaTable(L, result)
		L.Push(luaResult)
		return 1
	})

	// Convert command name to a valid Lua identifier
	luaName := strings.ReplaceAll(name, "-", "_")

	// Register the function in the global environment
	L.SetGlobal(luaName, fn)

	// Update the parameter information global name
	paramsGlobalName := luaName + "_params"

	// Register the function in the global environment
	L.SetGlobal(name, fn)

	// Optionally, you can also register parameter information
	defaultLayer, ok := desc.GetDefaultLayer()
	if !ok {
		// TODO error handling
		return
	}
	paramsTable := L.CreateTable(0, len(defaultLayer.GetParameterDefinitions().ToList()))
	defaultLayer.GetParameterDefinitions().ForEach(func(param *parameters.ParameterDefinition) {
		paramInfo := L.CreateTable(0, 4)
		paramInfo.RawSetString("name", lua.LString(param.Name))
		paramInfo.RawSetString("type", lua.LString(string(param.Type)))
		paramInfo.RawSetString("description", lua.LString(param.Help))
		paramInfo.RawSetString("default", InterfaceToLuaValue(L, param.Default))
		paramsTable.RawSetString(param.Name, paramInfo)
	})
	L.SetGlobal(paramsGlobalName, paramsTable)
}
