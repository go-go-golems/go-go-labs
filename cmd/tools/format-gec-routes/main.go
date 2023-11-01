package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
	"text/template"
)

// Define the structures
type DocString struct {
	Annotations      map[string]interface{} `yaml:"annotations"`
	MethodOrField    string                 `yaml:"methodorfield"`
	FullDocString    string                 `yaml:"fulldocstring"`
	ContentDocString string                 `yaml:"contentdocstring"`
}

// Template for generating markdown
var markdownTemplate = `
# Table of Contents

{{ range . -}}
{{ if .Annotations.url -}}
- [{{ (index .Annotations.url 0) }}](#{{ (index .Annotations.url 0) }}){{- end }}
{{ end }}

# API Endpoints

{{ range . -}}
{{if .Annotations.url }}## {{ (index .Annotations.url 0) }}

{{ .ContentDocString | parseContent }}
{{ end }}
{{ end }}
`

func parseContentDocString(content string) string {
	lines := strings.Split(content, "\n")
	var parsedLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "@") {
			parsedLines = append(parsedLines, "- `"+trimmedLine+"`")
		} else {
			parsedLines = append(parsedLines, trimmedLine)
		}
	}

	return strings.Join(parsedLines, "\n")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <yaml file>")
		return
	}
	// Read the YAML data
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	// Parse the YAML data into our structures
	var annotations []DocString
	err = yaml.Unmarshal(data, &annotations)
	if err != nil {
		panic(err)
	}

	funcMap := template.FuncMap{
		"parseContent": parseContentDocString,
	}
	// Generate markdown using the template
	tmpl, err := template.New("markdown").Funcs(funcMap).Parse(markdownTemplate)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(os.Stdout, annotations)
	if err != nil {
		panic(err)
	}
}
