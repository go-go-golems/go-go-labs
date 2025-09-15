package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/example/agent-action/internal/action"
)

func main() {
	ctx := context.Background()

	inputs, err := action.ParseInputs(os.Args[1:], os.Getenv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "input error: %v\n", err)
		os.Exit(1)
	}

	env := action.LoadRuntimeEnv(os.Getenv)

	token := inputs.GitHubToken
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	gh := action.NewGitHubClient(ctx, token)

	tool, err := buildTool(inputs, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tool setup error: %v\n", err)
		os.Exit(1)
	}

	runner := &action.Runner{
		Inputs:     inputs,
		Env:        env,
		GitHub:     gh,
		Tool:       tool,
		FileLoader: os.ReadFile,
	}

	if err := runner.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "run error: %v\n", err)
		os.Exit(1)
	}
}

func buildTool(in *action.Inputs, env action.RuntimeEnv) (action.ReviewTool, error) {
	mode := strings.ToLower(strings.TrimSpace(in.ToolMode))
	switch mode {
	case "", "mock":
		return action.MockTool{}, nil
	case "http":
		if in.ToolURL == "" {
			return nil, fmt.Errorf("tool_url is required for http mode")
		}
		return &action.HTTPTool{
			Client:  http.DefaultClient,
			URL:     in.ToolURL,
			Method:  in.ToolMethod,
			Headers: in.ToolHeaders,
			Token:   in.ToolToken,
		}, nil
	case "cmd":
		workingDir := in.WorkingDir
		if workingDir == "" {
			workingDir = env.Workspace
		}
		return &action.CommandTool{
			Command: in.ToolCmd,
			Args:    in.ToolArgs,
			Dir:     workingDir,
		}, nil
	default:
		return nil, fmt.Errorf("unknown tool_mode %q", in.ToolMode)
	}
}
