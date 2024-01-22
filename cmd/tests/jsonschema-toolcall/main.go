package main

import (
	"github.com/go-go-golems/geppetto/pkg/helpers"
	"github.com/invopop/jsonschema"
)

// https://chat.openai.com/c/08d9679f-7f10-41b3-b777-bcb7ed479918

import (
	"encoding/json"
	"fmt"
)

func exampleFunction(a int, b string) {
	fmt.Printf("Function called with a = %d and b = %s\n", a, b)
}

func main() {
	// Example usage
	funcJson := `[123, "hello"]` // JSON representation of the arguments
	var funcArgs interface{}

	reflector := new(jsonschema.Reflector)
	schema, _ := helpers.GetFunctionParametersJsonSchema(reflector, exampleFunction)
	s, _ := json.MarshalIndent(schema, "", " ")
	fmt.Printf("json schema: \n%s\n", s)

	if err := json.Unmarshal([]byte(funcJson), &funcArgs); err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	if _, err := helpers.CallFunctionFromJson(exampleFunction, funcArgs); err != nil {
		fmt.Println("Error calling function:", err)
		return
	}
}
