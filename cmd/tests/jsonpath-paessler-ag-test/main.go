package main

import (
	"encoding/json"
	"fmt"
	"github.com/PaesslerAG/jsonpath"
	"os"
)

func main() {
	v := interface{}(nil)
	err := json.Unmarshal([]byte(`{"welcome":{"message":["Good Morning", "Hello World!"]}}`), &v)
	if err != nil {
		panic(err)
	}

	welcome, err := jsonpath.Get("$.welcome.message[1]", v)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(welcome) // Output: Hello World!

	data := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "foo",
				},
			},
			map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "bar",
				},
			},
			123,
		},
	}

	res, err := jsonpath.Get("$.items[*]", data)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}
