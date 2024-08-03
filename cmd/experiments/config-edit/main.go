package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: program <config_file_path> <new_repository_path>")
		os.Exit(1)
	}

	configPath := os.Args[1]
	newRepoPath := os.Args[2]

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	var root yaml.Node
	err = yaml.Unmarshal(data, &root)
	if err != nil {
		fmt.Printf("Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	if findAndUpdateRepositoriesNode(&root, newRepoPath) {
		// Write the updated YAML back to the file
		updatedData, err := yaml.Marshal(&root)
		if err != nil {
			fmt.Printf("Error marshaling YAML: %v\n", err)
			os.Exit(1)
		}

		err = os.WriteFile(configPath, updatedData, 0644)
		if err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Config file updated successfully")
	} else {
		fmt.Println("'repositories' key not found in the config file")
	}
}

func findAndUpdateRepositoriesNode(node *yaml.Node, newRepoPath string) bool {
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]

			if key.Value == "repositories" {
				// Print out the values in value.Content before appending the new node
				fmt.Println("Current repository paths:")
				for _, contentNode := range value.Content {
					fmt.Printf("- %s (kind: %s, style: %s)\n", contentNode.Value, getKindString(contentNode.Kind), getStyleString(contentNode.Style))
				}

				// Add the new repository path to the list
				newNode := &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: newRepoPath,
				}
				value.Content = append(value.Content, newNode)
				return true
			}
		}
	} else if node.Kind == yaml.SequenceNode || node.Kind == yaml.DocumentNode {
		for _, child := range node.Content {
			if findAndUpdateRepositoriesNode(child, newRepoPath) {
				return true
			}
		}
	}

	return false
}

func getKindString(kind yaml.Kind) string {
	switch kind {
	case yaml.DocumentNode:
		return "DocumentNode"
	case yaml.SequenceNode:
		return "SequenceNode"
	case yaml.MappingNode:
		return "MappingNode"
	case yaml.ScalarNode:
		return "ScalarNode"
	case yaml.AliasNode:
		return "AliasNode"
	default:
		return "UnknownNode"
	}
}

func getStyleString(style yaml.Style) string {
	switch style {
	case yaml.LiteralStyle:
		return "LiteralStyle"
	case yaml.FlowStyle:
		return "FlowStyle"
	case yaml.FoldedStyle:
		return "FoldedStyle"
	case yaml.DoubleQuotedStyle:
		return "DoubleQuotedStyle"
	case yaml.SingleQuotedStyle:
		return "SingleQuotedStyle"
	case yaml.TaggedStyle:
		return "TaggedStyle"
	default:
		return fmt.Sprintf("UnknownStyle(%d)", style)
	}
}
