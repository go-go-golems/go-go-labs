package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <log_file>")
		os.Exit(1)
	}

	logFile := os.Args[1]
	file, err := os.Open(logFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error reading line: %v\n", err)
			continue
		}

		parts := strings.SplitN(line, " event=", 2)
		if len(parts) != 2 {
			continue
		}

		jsonStr := strings.TrimSpace(parts[1])
		var data interface{}
		err = json.Unmarshal([]byte(jsonStr), &data)
		if err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			continue
		}

		yamlData, err := yaml.Marshal(data)
		if err != nil {
			fmt.Printf("Error converting to YAML: %v\n", err)
			continue
		}

		fmt.Println(string(yamlData))
		fmt.Println("---") // YAML document separator
	}
}
