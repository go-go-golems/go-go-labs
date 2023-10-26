package pkg

import (
	"github.com/dave/jennifer/jen"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// Things to build:
//    ParameterDefinition list to a struct

type Foobar struct {
	I int `json:"i"`
}

func ParameterDefinitionToStructMember(p *parameters.ParameterDefinition) *jen.Statement {
	return jen.Id(p.Name).Id(string(p.Type)).Tag(map[string]string{"json": p.Name})
}
