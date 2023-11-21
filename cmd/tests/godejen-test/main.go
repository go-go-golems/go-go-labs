package main

import (
	"context"
	"fmt"
	"github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/clay/pkg/sql"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	sqleton_cmds "github.com/go-go-golems/sqleton/pkg/cmds"
	"github.com/go-go-golems/sqleton/pkg/flags"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
)

//go:generate sqleton codegen ps.yaml

type PsCommandWrap struct {
	Command             *PsCommand
	description         *cmds.CommandDescription
	dbConnectionFactory sqleton_cmds.DBConnectionFactory
}

func NewPsCommandWrap() (*PsCommandWrap, error) {
	psCommand, err := NewPsCommand()
	if err != nil {
		return nil, err
	}
	origDesc := psCommand.CommandDescription

	sqlHelpersLayer, err := flags.NewSqlHelpersParameterLayer()
	if err != nil {
		return nil, err
	}
	sqlLayer, err := sql.NewSqlConnectionParameterLayer()
	if err != nil {
		return nil, err
	}
	dbtLayer, err := sql.NewDbtParameterLayer()
	if err != nil {
		return nil, err
	}
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	layers := []layers.ParameterLayer{
		sqlHelpersLayer,
		glazedLayer,
		sqlLayer,
		dbtLayer,
	}

	desc := origDesc.Clone(true, cmds.WithLayers(layers...))

	return &PsCommandWrap{
		psCommand,
		desc,
		sql.OpenDatabaseFromDefaultSqlConnectionLayer,
	}, nil
}

var _ cmds.GlazeCommand = (*PsCommandWrap)(nil)

func (p *PsCommandWrap) Description() *cmds.CommandDescription {
	return p.description
}

func (p *PsCommandWrap) ToYAML(w io.Writer) error {
	return p.Command.ToYAML(w)
}

func (p *PsCommandWrap) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	if p.dbConnectionFactory == nil {
		return fmt.Errorf("dbConnectionFactory is not set")
	}

	// at this point, the factory can probably be passed the sql-connection parsed layer
	db, err := p.dbConnectionFactory(parsedLayers)
	if err != nil {
		return err
	}
	defer func(db *sqlx.DB) {
		_ = db.Close()
	}(db)

	err = db.PingContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "Could not ping database")
	}

	psParameters := &PsCommandParameters{}
	err = parameters.InitializeStructFromParameters(psParameters, ps)
	if err != nil {
		return err
	}
	return p.Command.RunIntoGlazed(ctx, db, psParameters, gp)
}

func main() {
	psCommand, err := NewPsCommandWrap()
	cobra.CheckErr(err)
	cobraPsCommand, err := cli.BuildCobraCommandFromGlazeCommand(psCommand)
	cobra.CheckErr(err)
	err = pkg.InitViper("sqleton", cobraPsCommand)
	cobra.CheckErr(err)

	err = cobraPsCommand.Execute()
	cobra.CheckErr(err)
}
