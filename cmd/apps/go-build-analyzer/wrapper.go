package main

import (
	"database/sql"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// isWrapperInvocation returns true if the program seems to be invoked by `go build -toolexec`.
func isWrapperInvocation() bool {
    // When used as toolexec, os.Args will include at least: [wrapper, /path/to/tool, ...]
    if len(os.Args) < 2 {
        return false
    }
    toolPath := os.Args[1]
    base := filepath.Base(toolPath)
    if base == "" {
        return false
    }
    // Heuristic: If first arg basename matches a known tool or contains path separators, treat as wrapper
    switch base {
    case "compile", "asm", "pack", "link", "cgo", "cover", "vet":
        return true
    }
    if strings.Contains(toolPath, "/") || strings.Contains(toolPath, "\\") {
        return true
    }
    return false
}

func runWrapper() {
    args := os.Args[1:]
    if len(args) == 0 {
        os.Exit(1)
    }

    toolPath := args[0]
    tool := filepath.Base(toolPath)

    // Extract details from tool-specific args
    toolArgs := []string{}
    if len(args) > 1 {
        toolArgs = args[1:]
    }
    details := parseToolArgs(tool, toolArgs)

    // Execute the tool
    start := time.Now()
    cmd := exec.Command(toolPath, toolArgs...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    err := cmd.Run()
    elapsed := time.Since(start)
    status := exitStatus(err)

    // Best-effort logging that doesn't affect the build
    _ = logInvocation(tool, toolPath, details, status, elapsed, args)

    os.Exit(status)
}

func exitStatus(err error) int {
    if err == nil {
        return 0
    }
    // If it's an *ExitError, use its code; otherwise 1
    if ee, ok := err.(*exec.ExitError); ok {
        if ws, ok := ee.Sys().(interface{ ExitStatus() int }); ok {
            return ws.ExitStatus()
        }
    }
    return 1
}

type parsedArgs struct {
    Pkg         string            `json:"pkg"`
    Out         string            `json:"out"`
    ImportCfg   string            `json:"importcfg"`
    EmbedCfg    string            `json:"embedcfg"`
    BuildID     string            `json:"buildid"`
    GoVersion   string            `json:"goversion"`
    Lang        string            `json:"lang"`
    Concurrency int               `json:"concurrency"`
    Complete    bool              `json:"complete"`
    Pack        bool              `json:"pack"`
    SourceCount int               `json:"source_count"`
    Flags       map[string]string `json:"flags"`
}

func parseToolArgs(tool string, toolArgs []string) parsedArgs {
    d := parsedArgs{Flags: map[string]string{}}
    for i := 0; i < len(toolArgs); i++ {
        a := toolArgs[i]
        if !strings.HasPrefix(a, "-") {
            d.SourceCount++
            continue
        }
        key := a
        val := ""
        if eq := strings.IndexByte(a, '='); eq >= 0 {
            key = a[:eq]
            val = a[eq+1:]
        } else {
            if i+1 < len(toolArgs) && !strings.HasPrefix(toolArgs[i+1], "-") {
                val = toolArgs[i+1]
                i++
            }
        }
        for strings.HasPrefix(key, "-") {
            key = key[1:]
        }
        switch key {
        case "p":
            d.Pkg = val
        case "o":
            d.Out = val
        case "importcfg":
            d.ImportCfg = val
        case "embedcfg":
            d.EmbedCfg = val
        case "buildid":
            d.BuildID = val
        case "goversion":
            d.GoVersion = val
        case "lang":
            d.Lang = val
        case "c":
            if n, err := strconv.Atoi(val); err == nil {
                d.Concurrency = n
            }
        case "complete":
            d.Complete = true
        case "pack":
            d.Pack = true
        default:
            if val == "" {
                d.Flags[key] = "true"
            } else {
                d.Flags[key] = val
            }
        }
    }
    return d
}

func logInvocation(tool, toolPath string, details parsedArgs, status int, elapsed time.Duration, fullArgs []string) error {
    dbPath := getDBPath()
    dsn := dbPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout=5000"
    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        return errors.WithStack(err)
    }
    defer db.Close()

    if err := ensureSchema(db); err != nil {
        return errors.WithStack(err)
    }

    // Determine run_id from env
    runID, hasRun := currentRunID()

    // Insert invocation row
    argsJoined := strings.Join(fullArgs, " ")
    ts := time.Now().Unix()
    goos := runtime.GOOS
    goarch := runtime.GOARCH
    cwd, _ := os.Getwd()
    flagsJSON := ""
    if b, err := json.Marshal(details); err == nil {
        flagsJSON = string(b)
    }

    if hasRun {
        _, err = db.Exec(`
INSERT INTO invocations (
  run_id, ts_unix, tool, tool_path, pkg, status, elapsed_ms, args,
  os, arch, cwd,
  out, importcfg, embedcfg, buildid, goversion, lang, concurrency, complete, pack, source_count, flags_json
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, runID, ts, tool, toolPath, details.Pkg, status, elapsed.Milliseconds(), argsJoined,
            goos, goarch, cwd,
            details.Out, details.ImportCfg, details.EmbedCfg, details.BuildID, details.GoVersion, details.Lang, details.Concurrency, boolToInt(details.Complete), boolToInt(details.Pack), details.SourceCount, flagsJSON)
    } else {
        _, err = db.Exec(`
INSERT INTO invocations (
  ts_unix, tool, tool_path, pkg, status, elapsed_ms, args,
  os, arch, cwd,
  out, importcfg, embedcfg, buildid, goversion, lang, concurrency, complete, pack, source_count, flags_json
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, ts, tool, toolPath, details.Pkg, status, elapsed.Milliseconds(), argsJoined,
            goos, goarch, cwd,
            details.Out, details.ImportCfg, details.EmbedCfg, details.BuildID, details.GoVersion, details.Lang, details.Concurrency, boolToInt(details.Complete), boolToInt(details.Pack), details.SourceCount, flagsJSON)
    }
    if err != nil {
        return errors.WithStack(err)
    }
    return nil
}

func boolToInt(b bool) int { if b { return 1 }; return 0 }


