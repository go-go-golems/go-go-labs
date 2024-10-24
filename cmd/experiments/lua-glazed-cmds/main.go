package main

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/yuin/gopher-lua"
)

// Middleware to parse Lua table into a ParsedLayer
func ParseLuaTableMiddleware(L *lua.LState, tableName string, paramDefs *parameters.ParameterDefinitions) middlewares.Middleware {
	return func(next middlewares.HandlerFunc) middlewares.HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			luaTable := L.GetGlobal(tableName).(*lua.LTable)
			params := make(map[string]interface{})
			luaTable.ForEach(func(key, value lua.LValue) {
				if keyStr, ok := key.(lua.LString); ok {
					paramDef, _ := paramDefs.Get(string(keyStr))
					params[string(keyStr)] = luaValueToGo(value, paramDef)
				}
			})

			// Look up the specific layer
			layer, ok := layers_.Get("user")
			if !ok {
				return fmt.Errorf("layer 'user' not found")
			}

			// Parse parameters using the layer's definitions
			parsedParams, err := layer.GetParameterDefinitions().GatherParametersFromMap(params, true)
			if err != nil {
				return err
			}

			// Create a parsed layer and merge it into parsedLayers
			parsedLayer, err := layers.NewParsedLayer(layer, layers.WithParsedParameters(parsedParams))
			if err != nil {
				return err
			}

			parsedLayers.GetOrCreate(layer).MergeParameters(parsedLayer)
			return next(layers_, parsedLayers)
		}
	}
}

// Extended luaValueToGo function
func luaValueToGo(value lua.LValue, paramDef *parameters.ParameterDefinition) interface{} {
	switch paramDef.Type {
	case parameters.ParameterTypeString:
		if v, ok := value.(lua.LString); ok {
			return string(v)
		}
	case parameters.ParameterTypeInteger:
		if v, ok := value.(lua.LNumber); ok {
			return int(v)
		}
	case parameters.ParameterTypeFloat:
		if v, ok := value.(lua.LNumber); ok {
			return float64(v)
		}
	case parameters.ParameterTypeBool:
		if v, ok := value.(lua.LBool); ok {
			return bool(v)
		}
	case parameters.ParameterTypeStringList:
		if tbl, ok := value.(*lua.LTable); ok {
			var list []string
			tbl.ForEach(func(_, v lua.LValue) {
				if str, ok := v.(lua.LString); ok {
					list = append(list, string(str))
				}
			})
			return list
		}
	case parameters.ParameterTypeIntegerList:
		if tbl, ok := value.(*lua.LTable); ok {
			var list []int
			tbl.ForEach(func(_, v lua.LValue) {
				if num, ok := v.(lua.LNumber); ok {
					list = append(list, int(num))
				}
			})
			return list
		}
	case parameters.ParameterTypeFloatList:
		if tbl, ok := value.(*lua.LTable); ok {
			var list []float64
			tbl.ForEach(func(_, v lua.LValue) {
				if num, ok := v.(lua.LNumber); ok {
					list = append(list, float64(num))
				}
			})
			return list
		}
	case parameters.ParameterTypeDate:
		if v, ok := value.(lua.LString); ok {
			parsedDate, err := parameters.ParseDate(string(v))
			if err == nil {
				return parsedDate
			}
		}
	// Add more cases for other parameter types as needed
	default:
		return nil
	}
	return nil
}

func main() {
	L := lua.NewState()
	defer L.Close()

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
	parsedLayers := layers.NewParsedLayers()

	// Execute middlewares
	err = middlewares.ExecuteMiddlewares(parameterLayers, parsedLayers,
		ParseLuaTableMiddleware(L, "params", paramDefs),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Parsed parameters:", parsedLayers.GetDataMap())
}
