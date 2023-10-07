package pkg

import (
	"github.com/go-go-golems/glazed/pkg/types"
	"gopkg.in/yaml.v3"
	"reflect"
	"strings"
)

type MultiLineString string

func (s MultiLineString) MarshalYAML() (interface{}, error) {
	if strings.Contains(string(s), "\n") {
		// remove all whitespace preceding newlines
		// use the regex \s+\n to match all whitespace preceding a newline
		// and replace it with just the newline
		node := yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: string(s),
			Style: yaml.LiteralStyle, // Set style to Literal for multiline
		}
		return &node, nil
	}
	return string(s), nil
}

func WalkForMultiline(i interface{}) interface{} {
	val := reflect.ValueOf(i)

	// Handle maps
	if val.Kind() == reflect.Map {
		newMap := reflect.MakeMap(val.Type())
		for _, key := range val.MapKeys() {
			newValue := WalkForMultiline(val.MapIndex(key).Interface())
			newMap.SetMapIndex(key, reflect.ValueOf(newValue))
		}
		return newMap.Interface()
	}

	// Handle slices and arrays
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		newSlice := reflect.MakeSlice(val.Type(), val.Len(), val.Cap())
		for i := 0; i < val.Len(); i++ {
			newValue := WalkForMultiline(val.Index(i).Interface())
			newSlice.Index(i).Set(reflect.ValueOf(newValue))
		}
		return newSlice.Interface()
	}

	switch v_ := i.(type) {
	case types.Row:
		for pair := v_.Oldest(); pair != nil; pair = pair.Next() {
			k, v := pair.Key, pair.Value
			v_.Set(k, WalkForMultiline(v))
		}
	case []types.Row:
		for i, v__ := range v_ {
			v_[i] = WalkForMultiline(v__).(types.Row)
		}
		return v_
	case string:
		return MultiLineString(val.String())
	}

	return i
}
