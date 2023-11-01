package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

func prettyPrintJSON(raw string) (string, error) {
	var buf bytes.Buffer
	err := json.Indent(&buf, []byte(raw), "", "  ")
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func processValue(data interface{}) interface{} {
	switch v := data.(type) {
	case map[interface{}]interface{}:
		for key, value := range v {
			v[key] = processValue(value)
		}
	case []interface{}:
		for i, item := range v {
			v[i] = processValue(item)
		}
	case string:
		if strings.HasPrefix(v, "{") || strings.HasPrefix(v, "[") {
			prettyJSON, err := prettyPrintJSON(v)
			if err == nil {
				return prettyJSON
			}
		}
	}
	return data
}

func main() {
	var yamlFile string

	var cmd = &cobra.Command{
		Use:   "transform",
		Short: "Transforms a YAML containing JSON strings to multi-line indented JSON",
		Run: func(cmd *cobra.Command, args []string) {
			content, err := os.ReadFile(yamlFile)
			if err != nil {
				panic(err)
			}

			var data interface{}
			err = yaml.Unmarshal(content, &data)
			if err != nil {
				panic(err)
			}

			transformedData := processValue(data)
			outContent, err := yaml.Marshal(transformedData)
			if err != nil {
				panic(err)
			}

			fmt.Print(string(outContent))
		},
	}

	cmd.Flags().StringVarP(&yamlFile, "file", "f", "", "YAML file to be transformed")
	_ = cmd.MarkFlagRequired("file")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
