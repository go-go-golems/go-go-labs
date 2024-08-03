package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"time"
)

// Define a struct with time.Time and *time.Time fields
type MyStruct struct {
	SimpleTime  time.Time  `yaml:"simple_time"`
	PointerTime *time.Time `yaml:"pointer_time"`
}

func main() {
	// Initialize the struct with current time
	now := time.Now()
	//myStruct := MyStruct{
	//	SimpleTime:  now,
	//	PointerTime: &now, // Initialize the pointer field
	//}

	t := interface{}(now)
	t2 := t.(time.Time)
	t3 := interface{}(interface{}(&t2))

	h := map[string]interface{}{
		"simple_time":  t2,
		"pointer_time": t3,
	}

	// Marshal the struct to YAML
	yamlData, err := yaml.Marshal(h)
	if err != nil {
		fmt.Printf("Error marshaling to YAML: %v\n", err)
		return
	}

	// Print the YAML output
	fmt.Println(string(yamlData))
}
