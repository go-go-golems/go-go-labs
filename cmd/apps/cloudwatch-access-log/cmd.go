package main

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"strings"
	"time"
)

type LogParserCommand struct {
	*cmds.CommandDescription
}

type LogEntry struct {
	Filename  string            `json:"filename"`
	Host      string            `json:"host"`
	Method    string            `json:"method"`
	Process   string            `json:"process"`
	Query     map[string]string `json:"-"`
	RawQuery  string            `json:"query"`
	Referer   string            `json:"referer"`
	RemoteIP  string            `json:"remoteIP"`
	Request   string            `json:"request"`
	Status    string            `json:"status"`
	Time      time.Time         `json:"-"`
	RawTime   string            `json:"time"`
	UniqueID  string            `json:"uniqueId"`
	UserAgent string            `json:"userAgent"`
}

func NewLogParserCommand() (*LogParserCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &LogParserCommand{
		CommandDescription: cmds.NewCommandDescription(
			"log-parser",
			cmds.WithShort("Parse log files"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"log",
					parameters.ParameterTypeFile,
					parameters.WithHelp("Path to the log file"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(
				glazedParameterLayer,
			),
		),
	}, nil
}

type LogParserSettings struct {
	LogFile *parameters.FileData `glazed.parameter:"log"`
}

func (c *LogParserCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &LogParserSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	file, err := os.Open(s.LogFile.Path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 {
			continue
		}

		var entry LogEntry
		err := json.Unmarshal([]byte(parts[2]), &entry)
		if err != nil {
			continue
		}

		entry.Time, _ = time.Parse(time.RFC3339, entry.RawTime)
		query, _ := url.ParseQuery(strings.TrimPrefix(entry.RawQuery, "?"))
		entry.Query = make(map[string]string)
		for k, v := range query {
			entry.Query[k] = v[0]
		}

		row := types.NewRow(
			types.MRP("filename", entry.Filename),
			types.MRP("host", entry.Host),
			types.MRP("method", entry.Method),
			types.MRP("process", entry.Process),
			types.MRP("query", entry.Query),
			types.MRP("referer", entry.Referer),
			types.MRP("remoteIP", entry.RemoteIP),
			types.MRP("request", entry.Request),
			types.MRP("status", entry.Status),
			types.MRP("time", entry.Time),
			types.MRP("uniqueID", entry.UniqueID),
			types.MRP("userAgent", entry.UserAgent),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
