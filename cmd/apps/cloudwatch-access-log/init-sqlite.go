package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"io"
)

type InitSQLiteCommand struct {
	*cmds.CommandDescription
}

type InitSQLiteSettings struct {
	DBFile string `glazed.parameter:"db-file"`
}

func NewInitSQLiteCommand() (*InitSQLiteCommand, error) {
	return &InitSQLiteCommand{
		CommandDescription: cmds.NewCommandDescription(
			"init-sqlite",
			cmds.WithShort("Initialize SQLite database"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"db-file",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to the SQLite database file"),
					parameters.WithRequired(true),
				),
			),
		),
	}, nil
}

func (c *InitSQLiteCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	s := &InitSQLiteSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", s.DBFile)
	if err != nil {
		return errors.Wrap(err, "failed to open SQLite database")
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS log_entries (
            id INTEGER PRIMARY KEY ,
            filename TEXT,
            host TEXT,
            method TEXT,
            process TEXT,
            query TEXT,
            referer TEXT,
            remoteIP TEXT,
            request TEXT,
            status TEXT,
            time TEXT,
            uniqueID TEXT,
            userAgent TEXT
        )
    `)
	if err != nil {
		return errors.Wrap(err, "failed to create log_entries table")
	}

	_, _ = fmt.Fprintf(w, "Successfully initialized SQLite database at %s\n", s.DBFile)
	_, _ = fmt.Fprintf(w, "Created table 'log_entries' with the specified schema\n")

	return nil
}
