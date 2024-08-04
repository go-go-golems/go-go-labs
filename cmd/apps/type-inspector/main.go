package main

import (
	"fmt"
	"reflect"
)

func analyzeValue(v reflect.Value, indent string) {
	t := v.Type()
	fmt.Printf("%sType: %v\n", indent, t)
	fmt.Printf("%sKind: %v\n", indent, v.Kind())

	//exhaustive:ignore
	switch v.Kind() {
	case reflect.Struct:
		fmt.Printf("%sFields:\n", indent)
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fmt.Printf("%s  %s:\n", indent, t.Field(i).Name)
			analyzeValue(field, indent+"    ")
		}

	case reflect.Slice, reflect.Array:
		fmt.Printf("%sLength: %d\n", indent, v.Len())
		fmt.Printf("%sCap: %d\n", indent, v.Cap())
		fmt.Printf("%sElements:\n", indent)
		for i := 0; i < v.Len(); i++ {
			fmt.Printf("%s  [%d]:\n", indent, i)
			analyzeValue(v.Index(i), indent+"    ")
		}

	case reflect.Map:
		fmt.Printf("%sLength: %d\n", indent, v.Len())
		fmt.Printf("%sEntries:\n", indent)
		for _, key := range v.MapKeys() {
			fmt.Printf("%s  Key:\n", indent)
			analyzeValue(key, indent+"    ")
			fmt.Printf("%s  Value:\n", indent)
			analyzeValue(v.MapIndex(key), indent+"    ")
		}

	case reflect.Ptr:
		if v.IsNil() {
			fmt.Printf("%sNil pointer\n", indent)
		} else {
			fmt.Printf("%sPointer to:\n", indent)
			analyzeValue(v.Elem(), indent+"  ")
		}

	case reflect.Interface:
		if v.IsNil() {
			fmt.Printf("%sNil interface\n", indent)
		} else {
			fmt.Printf("%sInterface value:\n", indent)
			analyzeValue(v.Elem(), indent+"  ")
		}

	case reflect.Func:
		fmt.Printf("%sFunction\n", indent)
		fmt.Printf("%s  IsVariadic: %v\n", indent, t.IsVariadic())
		fmt.Printf("%s  NumIn: %d\n", indent, t.NumIn())
		fmt.Printf("%s  NumOut: %d\n", indent, t.NumOut())

	case reflect.Chan:
		fmt.Printf("%sChannel\n", indent)
		fmt.Printf("%s  Direction: %v\n", indent, t.ChanDir())

	default:
		fmt.Printf("%sValue: %v\n", indent, v.Interface())
	}

	// Print common attributes
	fmt.Printf("%sCanAddr: %v\n", indent, v.CanAddr())
	fmt.Printf("%sCanSet: %v\n", indent, v.CanSet())
	fmt.Printf("%sIsValid: %v\n", indent, v.IsValid())
}

func Analyze(x interface{}) {
	v := reflect.ValueOf(x)
	analyzeValue(v, "")
}

func main() {
	// Example usage
	type Person struct {
		Name string
		Age  int
	}

	data := map[string]interface{}{
		"integer": 42,
		"float":   3.14,
		"string":  "hello",
		"slice":   []int{1, 2, 3},
		"struct":  Person{Name: "Alice", Age: 30},
		"map":     map[string]int{"a": 1, "b": 2},
	}

	Analyze(data)
}
