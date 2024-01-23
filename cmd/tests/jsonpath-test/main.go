package main

import (
	"fmt"
	"k8s.io/client-go/util/jsonpath"
	"os"
)

func main() {
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
	j := jsonpath.New("test")
	err := j.Parse("{.items[*]}")
	if err != nil {
		panic(err)
	}
	results, err := j.FindResults(data)
	if err != nil {
		panic(err)
	}
	for _, result := range results {
		err = j.PrintResults(os.Stdout, result)
		fmt.Println()
	}
	if err != nil {
		panic(err)
	}
}
