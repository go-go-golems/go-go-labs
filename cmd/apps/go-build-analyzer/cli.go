package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func runCLI(ctx context.Context) {
    rootCmd := &cobra.Command{
        Use:   "go-build-analyzer",
        Short: "Analyze Go build tool invocations and timings",
        Long:  "Toolexec wrapper and query CLI for analyzing Go build timings stored in SQLite.",
    }

    // Build commands
    commands := []cmds.Command{ }

    // runs new
    if c, err := NewRunsNewCommand(); err == nil {
        commands = append(commands, c)
    } else {
        fmt.Fprintln(os.Stderr, "error creating command runs new:", err)
    }
    // runs list
    if c, err := NewRunsListCommand(); err == nil {
        commands = append(commands, c)
    } else {
        fmt.Fprintln(os.Stderr, "error creating command runs list:", err)
    }
    // invocations list
    if c, err := NewInvocationsListCommand(); err == nil {
        commands = append(commands, c)
    } else {
        fmt.Fprintln(os.Stderr, "error creating command invocations list:", err)
    }
    // stats packages
    if c, err := NewStatsPackagesCommand(); err == nil {
        commands = append(commands, c)
    } else {
        fmt.Fprintln(os.Stderr, "error creating command stats packages:", err)
    }

    // Add as cobra subcommands
    for _, gc := range commands {
        cobraCmd, err := cli.BuildCobraCommand(gc,
            cli.WithParserConfig(cli.CobraParserConfig{
                ShortHelpLayers: []string{layers.DefaultSlug},
                MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
            }),
        )
        if err != nil {
            fmt.Fprintln(os.Stderr, "error building cobra command:", err)
            os.Exit(1)
        }
        rootCmd.AddCommand(cobraCmd)
    }

    // Enhanced help system
    helpSystem := help.NewHelpSystem()
    help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

// runs new
type RunsNewCommand struct {
    *cmds.CommandDescription
}

type RunsNewSettings struct {
    Comment string `glazed.parameter:"comment"`
    PrintEnv bool   `glazed.parameter:"print-env"`
}

func NewRunsNewCommand() (*RunsNewCommand, error) {
    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, errors.WithStack(err)
    }
    commandSettingsLayer, err := cli.NewCommandSettingsLayer()
    if err != nil {
        return nil, errors.WithStack(err)
    }

    cd := cmds.NewCommandDescription(
        "runs-new",
        cmds.WithShort("Create a new build run and return its id"),
        cmds.WithFlags(
            parameters.NewParameterDefinition("comment", parameters.ParameterTypeString, parameters.WithDefault(""), parameters.WithHelp("Optional run comment")),
            parameters.NewParameterDefinition("print-env", parameters.ParameterTypeBool, parameters.WithDefault(false), parameters.WithHelp("Print shell export for TOOLEXEC_RUN_ID")),
        ),
        cmds.WithLayersList(glazedLayer, commandSettingsLayer),
    )
    return &RunsNewCommand{CommandDescription: cd}, nil
}

var _ cmds.GlazeCommand = &RunsNewCommand{}

func (c *RunsNewCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    s := &RunsNewSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return errors.WithStack(err)
    }

    db, err := openDB(ctx)
    if err != nil {
        return errors.WithStack(err)
    }
    defer db.Close()

    id, err := insertRun(ctx, db, s.Comment)
    if err != nil {
        return errors.WithStack(err)
    }

    if s.PrintEnv {
        fmt.Printf("export TOOLEXEC_RUN_ID=%d\n", id)
    }

    row := types.NewRow(
        types.MRP("run_id", id),
        types.MRP("comment", s.Comment),
    )
    return gp.AddRow(ctx, row)
}

// runs list
type RunsListCommand struct {
    *cmds.CommandDescription
}

func NewRunsListCommand() (*RunsListCommand, error) {
    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, errors.WithStack(err)
    }
    commandSettingsLayer, err := cli.NewCommandSettingsLayer()
    if err != nil {
        return nil, errors.WithStack(err)
    }

    cd := cmds.NewCommandDescription(
        "runs-list",
        cmds.WithShort("List recorded runs"),
        cmds.WithLayersList(glazedLayer, commandSettingsLayer),
    )
    return &RunsListCommand{CommandDescription: cd}, nil
}

var _ cmds.GlazeCommand = &RunsListCommand{}

func (c *RunsListCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    db, err := openDB(ctx)
    if err != nil {
        return errors.WithStack(err)
    }
    defer db.Close()

    rows, err := db.QueryContext(ctx, `SELECT id, ts_unix, comment FROM runs ORDER BY id DESC`)
    if err != nil {
        return errors.WithStack(err)
    }
    defer rows.Close()

    for rows.Next() {
        var id int64
        var ts int64
        var comment string
        if err := rows.Scan(&id, &ts, &comment); err != nil {
            return errors.WithStack(err)
        }
        r := types.NewRow(
            types.MRP("run_id", id),
            types.MRP("ts_unix", ts),
            types.MRP("comment", comment),
        )
        if err := gp.AddRow(ctx, r); err != nil {
            return errors.WithStack(err)
        }
    }
    return nil
}

// invocations list
type InvocationsListCommand struct {
    *cmds.CommandDescription
}

type InvocationsListSettings struct {
    RunID  int64  `glazed.parameter:"run-id"`
    Tool   string `glazed.parameter:"tool"`
    Pkg    string `glazed.parameter:"pkg"`
    Limit  int    `glazed.parameter:"limit"`
}

func NewInvocationsListCommand() (*InvocationsListCommand, error) {
    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, errors.WithStack(err)
    }
    commandSettingsLayer, err := cli.NewCommandSettingsLayer()
    if err != nil {
        return nil, errors.WithStack(err)
    }

    cd := cmds.NewCommandDescription(
        "invocations-list",
        cmds.WithShort("List tool invocations, optionally filtered"),
        cmds.WithFlags(
            parameters.NewParameterDefinition("run-id", parameters.ParameterTypeInteger, parameters.WithDefault(0), parameters.WithHelp("Filter by run id (>0)")),
            parameters.NewParameterDefinition("tool", parameters.ParameterTypeString, parameters.WithDefault(""), parameters.WithHelp("Filter by tool name")),
            parameters.NewParameterDefinition("pkg", parameters.ParameterTypeString, parameters.WithDefault(""), parameters.WithHelp("Filter by package")),
            parameters.NewParameterDefinition("limit", parameters.ParameterTypeInteger, parameters.WithDefault(200), parameters.WithHelp("Limit number of rows")),
        ),
        cmds.WithLayersList(glazedLayer, commandSettingsLayer),
    )
    return &InvocationsListCommand{CommandDescription: cd}, nil
}

var _ cmds.GlazeCommand = &InvocationsListCommand{}

func (c *InvocationsListCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    s := &InvocationsListSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return errors.WithStack(err)
    }

    db, err := openDB(ctx)
    if err != nil {
        return errors.WithStack(err)
    }
    defer db.Close()

    q := "SELECT id, run_id, ts_unix, tool, tool_path, pkg, status, elapsed_ms, args, os, arch, cwd, out, importcfg, embedcfg, buildid, goversion, lang, concurrency, complete, pack, source_count, flags_json FROM invocations"
    var where []string
    var args []any
    if s.RunID > 0 {
        where = append(where, "run_id = ?")
        args = append(args, s.RunID)
    }
    if s.Tool != "" {
        where = append(where, "tool = ?")
        args = append(args, s.Tool)
    }
    if s.Pkg != "" {
        where = append(where, "pkg = ?")
        args = append(args, s.Pkg)
    }
    if len(where) > 0 {
        q += " WHERE " + stringsJoin(where, " AND ")
    }
    q += " ORDER BY id DESC"
    if s.Limit > 0 {
        q += fmt.Sprintf(" LIMIT %d", s.Limit)
    }

    rows, err := db.QueryContext(ctx, q, args...)
    if err != nil {
        return errors.WithStack(err)
    }
    defer rows.Close()

    for rows.Next() {
        var id, runID, ts int64
        var tool, toolPath, pkg string
        var status, elapsed int64
        var argstr string
        var osStr, archStr, cwd, out, importcfg, embedcfg, buildid, goversion, lang string
        var concurrency, complete, pack, sourceCount int64
        var flagsJSON string
        if err := rows.Scan(&id, &runID, &ts, &tool, &toolPath, &pkg, &status, &elapsed, &argstr, &osStr, &archStr, &cwd, &out, &importcfg, &embedcfg, &buildid, &goversion, &lang, &concurrency, &complete, &pack, &sourceCount, &flagsJSON); err != nil {
            return errors.WithStack(err)
        }
        r := types.NewRow(
            types.MRP("id", id),
            types.MRP("run_id", runID),
            types.MRP("ts_unix", ts),
            types.MRP("tool", tool),
            types.MRP("tool_path", toolPath),
            types.MRP("pkg", pkg),
            types.MRP("status", status),
            types.MRP("elapsed_ms", elapsed),
            types.MRP("args", argstr),
            types.MRP("os", osStr),
            types.MRP("arch", archStr),
            types.MRP("cwd", cwd),
            types.MRP("out", out),
            types.MRP("importcfg", importcfg),
            types.MRP("embedcfg", embedcfg),
            types.MRP("buildid", buildid),
            types.MRP("goversion", goversion),
            types.MRP("lang", lang),
            types.MRP("concurrency", concurrency),
            types.MRP("complete", complete),
            types.MRP("pack", pack),
            types.MRP("source_count", sourceCount),
            types.MRP("flags_json", flagsJSON),
        )
        if err := gp.AddRow(ctx, r); err != nil {
            return errors.WithStack(err)
        }
    }
    return nil
}

// stats packages (sum elapsed for compile by pkg)
type StatsPackagesCommand struct {
    *cmds.CommandDescription
}

type StatsPackagesSettings struct {
    RunID int64 `glazed.parameter:"run-id"`
    Tool  string `glazed.parameter:"tool"`
    Limit int   `glazed.parameter:"limit"`
}

func NewStatsPackagesCommand() (*StatsPackagesCommand, error) {
    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, errors.WithStack(err)
    }
    commandSettingsLayer, err := cli.NewCommandSettingsLayer()
    if err != nil {
        return nil, errors.WithStack(err)
    }

    cd := cmds.NewCommandDescription(
        "stats-packages",
        cmds.WithShort("Aggregate elapsed time by package (default tool=compile)"),
        cmds.WithFlags(
            parameters.NewParameterDefinition("run-id", parameters.ParameterTypeInteger, parameters.WithDefault(0), parameters.WithHelp("Filter by run id (>0)")),
            parameters.NewParameterDefinition("tool", parameters.ParameterTypeString, parameters.WithDefault("compile"), parameters.WithHelp("Tool to aggregate (compile/link/asm/etc)")),
            parameters.NewParameterDefinition("limit", parameters.ParameterTypeInteger, parameters.WithDefault(50), parameters.WithHelp("Limit number of rows")),
        ),
        cmds.WithLayersList(glazedLayer, commandSettingsLayer),
    )
    return &StatsPackagesCommand{CommandDescription: cd}, nil
}

var _ cmds.GlazeCommand = &StatsPackagesCommand{}

func (c *StatsPackagesCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    s := &StatsPackagesSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return errors.WithStack(err)
    }

    db, err := openDB(ctx)
    if err != nil {
        return errors.WithStack(err)
    }
    defer db.Close()

    q := "SELECT pkg, SUM(elapsed_ms) AS total_ms, COUNT(*) AS num FROM invocations WHERE tool = ?"
    var args []any
    args = append(args, s.Tool)
    if s.RunID > 0 {
        q += " AND run_id = ?"
        args = append(args, s.RunID)
    }
    q += " GROUP BY pkg ORDER BY total_ms DESC"
    if s.Limit > 0 {
        q += fmt.Sprintf(" LIMIT %d", s.Limit)
    }

    rows, err := db.QueryContext(ctx, q, args...)
    if err != nil {
        return errors.WithStack(err)
    }
    defer rows.Close()

    for rows.Next() {
        var pkg string
        var totalMs int64
        var num int64
        if err := rows.Scan(&pkg, &totalMs, &num); err != nil {
            return errors.WithStack(err)
        }
        r := types.NewRow(
            types.MRP("pkg", pkg),
            types.MRP("total_ms", totalMs),
            types.MRP("num", num),
        )
        if err := gp.AddRow(ctx, r); err != nil {
            return errors.WithStack(err)
        }
    }
    return nil
}

// Utility: join without importing strings (we already have a strings usage in another file)
func stringsJoin(parts []string, sep string) string {
    if len(parts) == 0 {
        return ""
    }
    out := parts[0]
    for i := 1; i < len(parts); i++ {
        out += sep + parts[i]
    }
    return out
}


