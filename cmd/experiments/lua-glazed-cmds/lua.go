package main

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/yuin/gopher-lua"
	"strings"
)

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
	parsedParams, err := layer.GetParameterDefinitions().GatherParametersFromMap(params, true)
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
				return fmt.Errorf("layer 'user' not found")
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
