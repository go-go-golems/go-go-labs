package main

import (
	"encoding/json"
	"fmt"
)

// printJSON prints data as formatted JSON
func printJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}
