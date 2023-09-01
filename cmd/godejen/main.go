package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/go-go-golems/glazed/pkg/cmds"
	cmds2 "github.com/go-go-golems/sqleton/pkg/cmds"
	"github.com/iancoleman/strcase"
)

//go:embed "ps.yaml"
var psYaml []byte

func LoadPS() (*cmds2.SqlCommand, error) {
	loader := &cmds2.SqlCommandLoader{
		DBConnectionFactory: nil,
	}

	// create reader from psYaml
	r := bytes.NewReader(psYaml)
	cmds_, err := loader.LoadCommandFromYAML(r)
	if err != nil {
		return nil, err
	}
	if len(cmds_) != 1 {
		return nil, fmt.Errorf("expected exactly one command, got %d", len(cmds_))
	}
	return cmds_[0].(*cmds2.SqlCommand), nil
}

func GenerateCommandCode(cmd *cmds.CommandDescription, query string) {
	f := jen.NewFile("main")

	// 1. Define the constant for the query.
	f.Const().Id(strcase.ToCamel(cmd.Name) + "CommandQuery").Op("=").Lit(query)

	// 2. Define struct.
	f.Type().Id(strcase.ToCamel(cmd.Name) + "Command").Struct(
		jen.Op("*").Id("SqlCommand"),
	)

	// 3. Define struct.
	psCommandParameters := jen.Type().Id(strcase.ToCamel(cmd.Name) + "CommandParameters").StructFunc(func(g *jen.Group) {
		for _, flag := range cmd.Flags {
			g.Id(strcase.ToCamel(flag.Name)).Id(string(flag.Type)).Tag(
				map[string]string{"glazed.parameter": strcase.ToSnake(flag.Name)})
		}
	})

	f.Add(psCommandParameters)

	// 4. Define Run method.
	f.Func().Params(jen.Id("p").Op("*").Id(strcase.ToCamel(cmd.Name)+"Command")).Id("Run").
		Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("db").Op("*").Qual("github.com/jmoiron/sqlx", "DB"),
			jen.Id("params").Op("*").Id(strcase.ToCamel(cmd.Name)+"CommandParameters"),
			jen.Id("gp").Id("middlewares.Processor"),
		).Error().Block(
		jen.List(jen.Id(strcase.ToCamel(cmd.Name)), jen.Id("err")).Op(":=").Id("parameters").Dot("StructToMap").Call(jen.Id("params")),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(
			jen.Return(jen.Id("err")),
		),
		jen.Id("err").Op("=").Id("p").Dot("SqlCommand").Dot("RunQueryIntoGlaze").Call(jen.Id("ctx"), jen.Id("db"), jen.Id(strcase.ToCamel(cmd.Name)), jen.Id("gp")),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(
			jen.Return(jen.Id("err")),
		),
		jen.Return(jen.Nil()),
	)

	// 5. Define New function.
	f.Func().Id("New"+strcase.ToCamel(cmd.Name)+"Command").Params().
		Params(jen.Op("*").Id(strcase.ToCamel(cmd.Name)+"Command"), jen.Error()).
		BlockFunc(func(g *jen.Group) {
			g.Var().Id("flagDefs").Op("=").
				Index().Op("*").
				Qual("github.com/go-go-golems/glazed/pkg/cmds/parameters", "ParameterDefinition").
				ValuesFunc(func(g *jen.Group) {
					for _, flag := range cmd.Flags {
						def := jen.Id("parameters").Dot("NewParameterDefinition").Call(
							jen.Add(
								jen.Line(),
								jen.Lit(flag.Name),
							),
							jen.Add(
								jen.Line(),
								jen.Lit(string(flag.Type)),
							),
							jen.Add(
								jen.Line(),
								jen.Id("parameters").Dot("WithHelp").Call(jen.Lit(flag.Help)),
							),
							jen.Line(),
						)
						g.Add(jen.Line(), def)
					}
				})

			// Add other contents for this block...
			g.Id("desc").Op(":=").Id("cmds").Dot("NewCommandDescription").Call(
				jen.Lit(cmd.Name),
				jen.Id("cmds").Dot("WithFlags").Call(jen.Id("flagDefs")),
			)
			g.List(jen.Id(strcase.ToCamel(cmd.Name)+"SqlCommand"), jen.Id("err")).Op(":=").Id("NewSqlCommand").Call(
				jen.Id("desc"),
				jen.Id("WithQuery").Call(jen.Id(strcase.ToCamel(cmd.Name)+"CommandQuery")),
			)
			g.If(jen.Id("err").Op("!=").Nil()).Block(
				jen.Return(jen.Nil(), jen.Id("err")),
			)
			g.Return(jen.Op("&").Id(strcase.ToCamel(cmd.Name)+"Command").Values(jen.Dict{
				jen.Id("SqlCommand"): jen.Id(strcase.ToCamel(cmd.Name) + "SqlCommand"),
			}), jen.Nil())
		})

	fmt.Printf("%#v", f) // Print the generated code
}

func main() {
	//genStaticPsYAML()

	cmd, err := LoadPS()
	if err != nil {
		panic(err)
	}

	GenerateCommandCode(cmd.Description(), cmd.Query)
}
