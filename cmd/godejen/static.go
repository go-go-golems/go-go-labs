package main

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
)

func genStaticPsYAML() {
	f := jen.NewFile("main") // this is the package name

	// 1. Define the constant for the query.
	f.Const().Id("psCommandQuery").Op("=").Lit("queryTemplate\n")

	// 2. Define PsCommand struct.
	f.Type().Id("PsCommand").Struct(
		jen.Op("*").Id("SqlCommand"),
	)

	// 3. Define PsCommandParameters struct.
	flags := []struct {
		Name string
		Type string
	}{
		{"MysqlUser", "[]string"},
		{"UserLike", "string"},
		{"Db", "string"},
		{"DbLike", "string"},
		{"State", "[]string"},
		{"InfoLike", "string"},
		{"ShortInfo", "bool"},
		{"MediumInfo", "bool"},
		{"FullInfo", "bool"},
	}

	psCommandParameters := jen.Type().Id("PsCommandParameters").StructFunc(func(g *jen.Group) {
		for _, flag := range flags {
			g.Id(flag.Name).Id(flag.Type).Tag(
				map[string]string{"glazed.parameter": strcase.ToSnake(flag.Name)})
		}
	})

	f.Add(psCommandParameters)

	// 4. Define Run method for the PsCommand
	f.Func().Params(jen.Id("p").Op("*").Id("PsCommand")).Id("Run").
		Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("db").Op("*").Qual("github.com/jmoiron/sqlx", "DB"),
			jen.Id("params").Op("*").Id("PsCommandParameters"),
			jen.Id("gp").Id("middlewares.Processor"),
		).Error().Block(
		jen.List(jen.Id("ps"), jen.Id("err")).Op(":=").Id("parameters").Dot("StructToMap").Call(jen.Id("params")),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(
			jen.Return(jen.Id("err")),
		),
		jen.Id("err").Op("=").Id("p").Dot("SqlCommand").Dot("RunQueryIntoGlaze").Call(jen.Id("ctx"), jen.Id("db"), jen.Id("ps"), jen.Id("gp")),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(
			jen.Return(jen.Id("err")),
		),
		jen.Return(jen.Nil()),
	)

	// 5. Define NewPSCommand function
	f.Func().Id("NewPSCommand").Params().Params(jen.Op("*").Id("PsCommand"), jen.Error()).Block(
		jen.Id("desc").Op(":=").Id("cmds").Dot("NewCommandDescription").Call(
			jen.Lit("ps"),
			jen.Id("cmds").Dot("WithFlags").Call(
				jen.Id("parameters").Dot("NewParameterDefinition").Call(
					jen.Lit("mysql_user"),
					jen.Id("parameters").Dot("ParameterTypeStringList"),
					jen.Id("parameters").Dot("WithHelp").Call(jen.Lit("Filter by user(s)")),
				),
			),
		),
		jen.List(jen.Id("psSqlCommand"), jen.Id("err")).Op(":=").Id("NewSqlCommand").Call(
			jen.Id("desc"),
			jen.Id("WithQuery").Call(jen.Id("psCommandQuery")),
		),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(
			jen.Return(jen.Nil(), jen.Id("err")),
		),
		jen.Return(jen.Op("&").Id("PsCommand").Values(jen.Dict{
			jen.Id("SqlCommand"): jen.Id("psSqlCommand"),
		}), jen.Nil()),
	)

	fmt.Printf("%#v", f)
}
