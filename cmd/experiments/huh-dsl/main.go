package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
)

func main() {
	yamlData := `
name: Burger Order Form
theme: Charm
groups:
  - name: Burger Selection
    fields:
      - type: select
        key: burger
        title: Choose your burger
        options:
          - label: Charmburger Classic
            value: classic
          - label: Chickwich
            value: chickwich
  - name: Order Details
    fields:
      - type: input
        key: name
        title: What's your name?
        validation:
          - condition: Frank
            error: Sorry, we don't serve customers named Frank.
      - type: confirm
        key: discount
        title: Would you like 15% off?
`

	var form Form
	err := yaml.Unmarshal([]byte(yamlData), &form)
	if err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}

	values, err := form.Run()
	if err != nil {
		log.Fatalf("Error running form: %v", err)
	}

	fmt.Println("Form Results:")
	for key, value := range values {
		fmt.Printf("%s: %v\n", key, value)
	}
}
