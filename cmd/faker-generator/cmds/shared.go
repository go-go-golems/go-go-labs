package cmds

import (
	"github.com/go-faker/faker/v4"
	"github.com/go-go-golems/go-emrichen/pkg/emrichen"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// --- Custom Tag Handlers ---

func handleFakerName(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.MappingNode && node.Kind != yaml.ScalarNode {
		// Allow simple !FakerName or !FakerName {}
	}
	// We ignore any parameters for now in this PoC

	name := faker.Name()
	return emrichen.ValueToNode(name)
}

func handleFakerEmail(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.MappingNode && node.Kind != yaml.ScalarNode {
		// Allow simple !FakerEmail or !FakerEmail {}
	}
	// We ignore any parameters for now in this PoC

	email := faker.Email()
	return emrichen.ValueToNode(email)
}

// --- Interpreter Factory ---

// NewFakerInterpreter creates a new Emrichen interpreter with the standard Faker tags registered.
func NewFakerInterpreter() (*emrichen.Interpreter, error) {
	fakerTags := emrichen.TagFuncMap{
		"!FakerName":  handleFakerName,
		"!FakerEmail": handleFakerEmail,
		// Add more tags here as needed
	}

	ei, err := emrichen.NewInterpreter(
		emrichen.WithAdditionalTags(fakerTags),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create emrichen interpreter with faker tags")
	}
	return ei, nil
}
