package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Options struct {
	MaxStringLength int
	MaxDepth        int
	MaxArrayLength  int
	OutputFormat    string
}

var options Options

var rootCmd = &cobra.Command{
	Use:   "json-shortener [file]",
	Short: "Shortens JSON/YAML content with configurable limits",
	Long: `A utility to shorten JSON/YAML content by limiting string length, recursion depth, and array size.
It reads from a file or stdin and outputs the shortened content to stdout.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var reader io.Reader
		if len(args) == 0 || args[0] == "-" {
			reader = os.Stdin
		} else {
			file, err := os.Open(args[0])
			if err != nil {
				return errors.Wrap(err, "failed to open input file")
			}
			defer file.Close()
			reader = file
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			return errors.Wrap(err, "failed to read input")
		}

		var parsedData interface{}
		inputFormat := ""

		// Try parsing as JSON first
		err = json.Unmarshal(data, &parsedData)
		if err == nil {
			inputFormat = "json"
		} else {
			// If JSON fails, try YAML
			err = yaml.Unmarshal(data, &parsedData)
			if err != nil {
				return errors.Wrap(err, "failed to parse input as JSON or YAML")
			}
			inputFormat = "yaml"
		}

		// Default to input format if not specified
		if options.OutputFormat == "auto" {
			options.OutputFormat = inputFormat
		}

		// Process the data with our shortening rules
		shortenedData := shortenValue(parsedData, 0)

		// Output the shortened data
		var output []byte
		if options.OutputFormat == "json" {
			output, err = json.MarshalIndent(shortenedData, "", "  ")
		} else {
			output, err = yaml.Marshal(shortenedData)
		}

		if err != nil {
			return errors.Wrap(err, "failed to marshal output")
		}

		fmt.Println(string(output))
		return nil
	},
}

// shortenValue recursively shortens a value based on the configured options
func shortenValue(value interface{}, depth int) interface{} {
	// Check depth limit
	if depth >= options.MaxDepth {
		return "[max depth]"
	}

	switch v := value.(type) {
	case string:
		return shortenString(v)
	case map[string]interface{}:
		return shortenMap(v, depth)
	case map[interface{}]interface{}:
		return shortenInterfaceMap(v, depth)
	case []interface{}:
		return shortenArray(v, depth)
	default:
		return v
	}
}

// shortenString truncates a string to the configured maximum length
func shortenString(s string) string {
	if len(s) <= options.MaxStringLength {
		return s
	}
	return s[:options.MaxStringLength] + "..."
}

// shortenMap shortens a map's values recursively
func shortenMap(m map[string]interface{}, depth int) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = shortenValue(v, depth+1)
	}
	return result
}

// shortenInterfaceMap handles YAML maps with interface{} keys
func shortenInterfaceMap(m map[interface{}]interface{}, depth int) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		// Convert key to string, using reflection for proper type handling
		var keyStr string
		switch kv := k.(type) {
		case string:
			keyStr = kv
		default:
			keyStr = fmt.Sprintf("%v", k)
		}
		result[keyStr] = shortenValue(v, depth+1)
	}
	return result
}

// shortenArray shortens an array to the configured maximum length
func shortenArray(a []interface{}, depth int) []interface{} {
	if len(a) <= options.MaxArrayLength {
		result := make([]interface{}, len(a))
		for i, v := range a {
			result[i] = shortenValue(v, depth+1)
		}
		return result
	}

	// Create a shortened array with the first N elements plus a count message
	result := make([]interface{}, options.MaxArrayLength+1)
	for i := 0; i < options.MaxArrayLength; i++ {
		result[i] = shortenValue(a[i], depth+1)
	}
	result[options.MaxArrayLength] = fmt.Sprintf("... (%d more items)", len(a)-options.MaxArrayLength)
	return result
}

func init() {
	rootCmd.Flags().IntVarP(&options.MaxStringLength, "string-length", "s", 50, "Maximum string length")
	rootCmd.Flags().IntVarP(&options.MaxDepth, "depth", "d", 10, "Maximum recursion depth")
	rootCmd.Flags().IntVarP(&options.MaxArrayLength, "array-length", "a", 10, "Maximum array length")
	rootCmd.Flags().StringVarP(&options.OutputFormat, "format", "f", "auto", "Output format (json, yaml, auto)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
