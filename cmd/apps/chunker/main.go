package main

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/tiktoken-go/tokenizer"
	"log"
)

func SplitString(input string, separators []string, model tokenizer.Model) (string, string, error) {
	codec, err := tokenizer.ForModel(model)
	if err != nil {
		return "", "", fmt.Errorf("Error getting codec: %v", err)
	}

	dfa := computeDFA(separators, codec)

	tokenIds, _, err := codec.Encode(input)
	if err != nil {
		return "", "", fmt.Errorf("Error encoding text: %v", err)
	}

	headIds, tailIds := splitTokenIdsByDFA(dfa, tokenIds, 10, codec)

	headString, err := codec.Decode(headIds)
	if err != nil {
		return "", "", fmt.Errorf("Error decoding headIds: %v", err)
	}

	tailString, err := codec.Decode(tailIds)
	if err != nil {
		return "", "", fmt.Errorf("Error decoding tailIds: %v", err)
	}

	return headString, tailString, nil
}

type SplitStringCommand struct {
	*cmds.CommandDescription
	Text       string   `flag:"text, t" help:"The text to split"`
	Separators []string `flag:"separators, s" help:"The separators to use"`
	Model      string   `flag:"model, m" help:"The model to use" default:"GPT4"`
}

func NewSplitStringCommand() (*SplitStringCommand, error) {
	return &SplitStringCommand{
		CommandDescription: cmds.NewCommandDescription(
			"split-string",
			cmds.WithShort("Split a string by specified separators"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"text",
					parameters.ParameterTypeString,
					parameters.WithHelp("The text to split"),
				),
				parameters.NewParameterDefinition(
					"separators",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("The separators to use"),
					parameters.WithDefault([]string{".", "!", "?"}),
				),
				parameters.NewParameterDefinition(
					"model",
					parameters.ParameterTypeString,
					parameters.WithHelp("The model to use"),
					parameters.WithDefault("GPT4"),
				),
			),
		),
	}, nil
}

func main() {
	separators := []string{".", "!", "?", "\n\n"}

	text := "This is a sentence. Another sentence! Yet another one?"
	head, tail, err := SplitString(text, separators, tokenizer.GPT4)
	if err != nil {
		log.Fatalf("Error in SplitString: %v", err)
	}

	fmt.Println("Head:", head)
	fmt.Println("Tail:", tail)
}
