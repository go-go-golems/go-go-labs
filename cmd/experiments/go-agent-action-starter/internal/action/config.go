package action

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Inputs combines all configuration toggles passed by the Action metadata.
type Inputs struct {
	TriggerPhrase      string
	LabelTrigger       string
	AssigneeTrigger    string
	GuidelinesPath     string
	IncludePatch       bool
	IncludeFileContent bool
	IncludeRepoGlobs   []string
	MaxFileBytes       int
	MaxChangedFiles    int

	ToolMode    string
	ToolURL     string
	ToolMethod  string
	ToolHeaders map[string]string
	ToolToken   string
	ToolCmd     string
	ToolArgs    []string
	WorkingDir  string

	OutputMode  string
	MaxComments int
	GitHubToken string
}

// ParseInputs wires CLI args and INPUT_* environment variables to the
// canonical Inputs struct. The lookup function is passed so tests can stub env.
func ParseInputs(args []string, lookup func(string) string) (*Inputs, error) {
	in := &Inputs{
		ToolHeaders: map[string]string{},
	}
	fs := flag.NewFlagSet("agent-action", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var headersJSON string
	var toolArgsJSON string
	var repoGlobsRaw string

	fs.StringVar(&in.TriggerPhrase, "trigger_phrase", envOrDefault(lookup, "INPUT_TRIGGER_PHRASE", "@agent"), "")
	fs.StringVar(&in.LabelTrigger, "label_trigger", lookup("INPUT_LABEL_TRIGGER"), "")
	fs.StringVar(&in.AssigneeTrigger, "assignee_trigger", lookup("INPUT_ASSIGNEE_TRIGGER"), "")

	fs.StringVar(&in.GuidelinesPath, "guidelines_path", envOrDefault(lookup, "INPUT_GUIDELINES_PATH", "CLAUDE.md"), "")
	fs.BoolVar(&in.IncludePatch, "include_patch", envBool(lookup, "INPUT_INCLUDE_PATCH", true), "")
	fs.BoolVar(&in.IncludeFileContent, "include_file_contents", envBool(lookup, "INPUT_INCLUDE_FILE_CONTENTS", false), "")
	fs.StringVar(&repoGlobsRaw, "include_repo_globs", lookup("INPUT_INCLUDE_REPO_GLOBS"), "")

	fs.IntVar(&in.MaxFileBytes, "max_file_bytes", envInt(lookup, "INPUT_MAX_FILE_BYTES", 200000), "")
	fs.IntVar(&in.MaxChangedFiles, "max_changed_files", envInt(lookup, "INPUT_MAX_CHANGED_FILES", 200), "")

	fs.StringVar(&in.ToolMode, "tool_mode", envOrDefault(lookup, "INPUT_TOOL_MODE", "mock"), "")
	fs.StringVar(&in.ToolURL, "tool_url", lookup("INPUT_TOOL_URL"), "")
	fs.StringVar(&in.ToolMethod, "tool_method", envOrDefault(lookup, "INPUT_TOOL_METHOD", "POST"), "")
	fs.StringVar(&headersJSON, "tool_headers_json", envOrDefault(lookup, "INPUT_TOOL_HEADERS_JSON", "{}"), "")
	fs.StringVar(&in.ToolToken, "tool_token", lookup("INPUT_TOOL_TOKEN"), "")
	fs.StringVar(&in.ToolCmd, "tool_cmd", lookup("INPUT_TOOL_CMD"), "")
	fs.StringVar(&toolArgsJSON, "tool_args_json", envOrDefault(lookup, "INPUT_TOOL_ARGS_JSON", "[]"), "")
	fs.StringVar(&in.WorkingDir, "working_directory", lookup("INPUT_WORKING_DIRECTORY"), "")

	fs.StringVar(&in.OutputMode, "output_mode", envOrDefault(lookup, "INPUT_OUTPUT_MODE", "review+summary"), "")
	fs.IntVar(&in.MaxComments, "max_comments", envInt(lookup, "INPUT_MAX_COMMENTS", 30), "")
	fs.StringVar(&in.GitHubToken, "github_token", lookup("INPUT_GITHUB_TOKEN"), "")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if repoGlobsRaw != "" {
		for _, part := range strings.Split(repoGlobsRaw, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				in.IncludeRepoGlobs = append(in.IncludeRepoGlobs, part)
			}
		}
	}

	if headersJSON != "" {
		if err := json.Unmarshal([]byte(headersJSON), &in.ToolHeaders); err != nil {
			return nil, fmt.Errorf("parse tool_headers_json: %w", err)
		}
	}
	if toolArgsJSON != "" {
		if err := json.Unmarshal([]byte(toolArgsJSON), &in.ToolArgs); err != nil {
			return nil, fmt.Errorf("parse tool_args_json: %w", err)
		}
	}

	return in, nil
}

func envOrDefault(lookup func(string) string, key, fallback string) string {
	if v := lookup(key); v != "" {
		return v
	}
	return fallback
}

func envBool(lookup func(string) string, key string, fallback bool) bool {
	if v := lookup(key); v != "" {
		switch strings.ToLower(v) {
		case "true", "1", "yes", "y":
			return true
		case "false", "0", "no", "n":
			return false
		}
	}
	return fallback
}

func envInt(lookup func(string) string, key string, fallback int) int {
	if v := lookup(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
