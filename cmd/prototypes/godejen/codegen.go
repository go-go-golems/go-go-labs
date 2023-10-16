package main

import (
	"github.com/dave/jennifer/jen"
	"github.com/go-go-golems/sqleton/pkg/cmds"
	"github.com/iancoleman/strcase"
)

type SqlCommandCodeGenerator struct {
	PackageName string
	SplitFiles  bool
}

func (s *SqlCommandCodeGenerator) GenerateCommandCode(
	cmd *cmds.SqlCommand,
) *jen.File {
	f := jen.NewFile(s.PackageName)

	// 1. Define the constant for the query.
	cmdName := strcase.ToLowerCamel(cmd.Name)

	f.Const().Id(strcase.ToCamel(cmdName) + "CommandQuery").Op("=").Lit(cmd.Query)

	// TODO(manuel, 2023-09-02) Add subqueries
	f.Const().Id(strcase.ToCamel(cmdName) + "CommandSubQueries").Op("=").Lit(cmd.SubQueries)

	// 2. Define struct.
	f.Type().Id(strcase.ToCamel(cmdName) + "Command").Struct(
		jen.Op("*").Id("SqlCommand"),
	)

	// 3. Define struct.
	psCommandParameters := jen.Type().Id(strcase.ToCamel(cmdName) + "CommandParameters").StructFunc(func(g *jen.Group) {
		for _, flag := range cmd.Flags {
			g.Id(strcase.ToCamel(flag.Name)).Id(string(flag.Type)).Tag(
				map[string]string{"glazed.parameter": strcase.ToSnake(flag.Name)})
		}
	})

	f.Add(psCommandParameters)

	// 4. Define Run method.
	f.Func().Params(jen.Id("p").Op("*").Id(strcase.ToCamel(cmdName)+"Command")).Id("Run").
		Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("db").Op("*").Qual("github.com/jmoiron/sqlx", "DB"),
			jen.Id("params").Op("*").Id(strcase.ToCamel(cmdName)+"CommandParameters"),
			jen.Id("gp").Id("middlewares.Processor"),
		).Error().Block(
		jen.List(jen.Id(strcase.ToCamel(cmdName)), jen.Id("err")).Op(":=").Id("parameters").Dot("StructToMap").Call(jen.Id("params")),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(
			jen.Return(jen.Id("err")),
		),
		jen.Id("err").Op("=").Id("p").Dot("SqlCommand").Dot("RunQueryIntoGlaze").Call(jen.Id("ctx"), jen.Id("db"), jen.Id(strcase.ToCamel(cmdName)), jen.Id("gp")),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(
			jen.Return(jen.Id("err")),
		),
		jen.Return(jen.Nil()),
	)

	// 5. Define New function.
	f.Func().Id("New"+strcase.ToCamel(cmdName)+"Command").Params().
		Params(jen.Op("*").Id(strcase.ToCamel(cmdName)+"Command"), jen.Error()).
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
				jen.Lit(cmdName),
				jen.Id("cmds").Dot("WithFlags").Call(jen.Id("flagDefs")),
			)
			g.List(jen.Id(strcase.ToCamel(cmdName)+"SqlCommand"), jen.Id("err")).Op(":=").Id("NewSqlCommand").Call(
				jen.Id("desc"),
				jen.Id("WithQuery").Call(jen.Id(strcase.ToCamel(cmdName)+"CommandQuery")),
			)
			g.If(jen.Id("err").Op("!=").Nil()).Block(
				jen.Return(jen.Nil(), jen.Id("err")),
			)
			g.Return(jen.Op("&").Id(strcase.ToCamel(cmdName)+"Command").Values(jen.Dict{
				jen.Id("SqlCommand"): jen.Id(strcase.ToCamel(cmdName) + "SqlCommand"),
			}), jen.Nil())
		})

	return f
}
