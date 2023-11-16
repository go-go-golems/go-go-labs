package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
)

type Query struct {
	PathRegexes      []*regexp.Regexp
	OperationRegexes []*regexp.Regexp
	TypeRegexes      []*regexp.Regexp
}

func compileRegexes(regexStrs []string) []*regexp.Regexp {
	var regexes []*regexp.Regexp
	for _, str := range regexStrs {
		regex, err := regexp.Compile(str)
		if err != nil {
			fmt.Printf("Invalid regular expression: %v\n", err)
			os.Exit(1)
		}
		regexes = append(regexes, regex)
	}
	return regexes
}

// findroutesCmd represents the findroutes command
var findroutesCmd = &cobra.Command{
	Use:   "findroutes [regexp]",
	Short: "Finds routes in an OpenAPI file matching a regular expression",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regex := args[0]
		openAPIFile, _ := cmd.Flags().GetString("openapi")

		// Load the OpenAPI file
		data, err := os.ReadFile(openAPIFile)
		if err != nil {
			fmt.Printf("Error reading OpenAPI file: %v\n", err)
			os.Exit(1)
		}

		// Parse the OpenAPI file
		swagger, err := openapi3.NewLoader().LoadFromData(data)
		if err != nil {
			fmt.Printf("Error parsing OpenAPI file: %v\n", err)
			os.Exit(1)
		}

		q := Query{
			PathRegexes: compileRegexes([]string{regex}),
		}

		q.ProcessOpenAPIFile(swagger)
	},
}

func (q *Query) ProcessOpenAPIFile(swagger *openapi3.T) {
	// Process each path in the OpenAPI document
	for path, pathItem := range swagger.Paths {
		if q.matchesAny(path, q.PathRegexes) {
			fmt.Println("Path:", path)
			q.printTypesForPathItem(pathItem)
		}
	}
}

func (q *Query) matchesAny(text string, regexes []*regexp.Regexp) bool {
	if len(regexes) == 0 {
		return true
	}
	for _, regex := range regexes {
		if regex.MatchString(text) {
			return true
		}
	}
	return false
}

func (q *Query) printTypesForPathItem(pathItem *openapi3.PathItem) {
	operations := map[string]*openapi3.Operation{
		"GET":     pathItem.Get,
		"PUT":     pathItem.Put,
		"POST":    pathItem.Post,
		"DELETE":  pathItem.Delete,
		"OPTIONS": pathItem.Options,
		"HEAD":    pathItem.Head,
		"PATCH":   pathItem.Patch,
		"TRACE":   pathItem.Trace,
	}

	for opName, operation := range operations {
		if operation != nil && q.matchesAny(opName, q.OperationRegexes) {
			fmt.Println("  Operation:", operation.OperationID)
			q.printTypesForOperation(operation)
		}
	}
}

func (q *Query) printTypesForOperation(operation *openapi3.Operation) {
	// Print parameter types
	for _, parameter := range operation.Parameters {
		if ref := parameter.Value.Schema.Ref; ref != "" && q.matchesAny(ref, q.TypeRegexes) {
			fmt.Println("    Parameter type:", getRefName(ref))
		}
	}

	// Print request body types
	if operation.RequestBody != nil {
		for _, mediaType := range operation.RequestBody.Value.Content {
			if ref := mediaType.Schema.Ref; ref != "" && q.matchesAny(ref, q.TypeRegexes) {
				fmt.Println("    Request body type:", getRefName(ref))
			}
		}
	}

	// Print response types
	for _, response := range operation.Responses {
		for _, mediaType := range response.Value.Content {
			if ref := mediaType.Schema.Ref; ref != "" && q.matchesAny(ref, q.TypeRegexes) {
				fmt.Println("    Response type:", getRefName(ref))
			}
		}
	}
}

func getRefName(ref string) string {
	// Extracts the last part of the $ref path which is typically the type name
	if lastIndex := regexp.MustCompile(`/`).FindAllStringIndex(ref, -1); lastIndex != nil {
		return ref[lastIndex[len(lastIndex)-1][1]:]
	}
	return ref
}

func init() {
	rootCmd.AddCommand(findroutesCmd)

	// Here you will define your flags and configuration settings.
	findroutesCmd.PersistentFlags().String("openapi", "", "Path to OpenAPI file")
	err := findroutesCmd.MarkPersistentFlagRequired("openapi")
	cobra.CheckErr(err)
}

var rootCmd = &cobra.Command{}

func main() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}
