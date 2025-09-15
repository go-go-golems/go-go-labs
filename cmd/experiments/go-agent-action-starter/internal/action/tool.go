package action

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

// ReviewTool is any component capable of turning PR context into review output.
type ReviewTool interface {
	Review(ctx context.Context, pr *PRContext) (*ReviewResult, error)
}

// HTTPTool posts the PR context to an external service and expects ReviewResult JSON back.
type HTTPTool struct {
	Client  *http.Client
	URL     string
	Method  string
	Headers map[string]string
	Token   string
}

func (t *HTTPTool) Review(ctx context.Context, pr *PRContext) (*ReviewResult, error) {
	if t.Client == nil {
		t.Client = http.DefaultClient
	}
	method := strings.ToUpper(strings.TrimSpace(t.Method))
	if method == "" {
		method = http.MethodPost
	}
	payload, err := json.Marshal(pr)
	if err != nil {
		return nil, fmt.Errorf("marshal context: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, t.URL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range t.Headers {
		req.Header.Set(k, v)
	}
	if t.Token != "" {
		req.Header.Set("Authorization", "Bearer "+t.Token)
	}

	resp, err := t.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tool HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result ReviewResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse tool response: %w", err)
	}
	return &result, nil
}

// CommandTool executes a local process, piping context JSON to stdin and reading ReviewResult JSON from stdout.
type CommandTool struct {
	Command string
	Args    []string
	Dir     string
	Runner  func(ctx context.Context, name string, args ...string) *exec.Cmd
}

func (c *CommandTool) Review(ctx context.Context, pr *PRContext) (*ReviewResult, error) {
	if c.Command == "" {
		return nil, fmt.Errorf("tool command is required")
	}
	run := c.Runner
	if run == nil {
		run = exec.CommandContext
	}
	cmd := run(ctx, c.Command, c.Args...)
	if c.Dir != "" {
		cmd.Dir = c.Dir
	}

	payload, err := json.Marshal(pr)
	if err != nil {
		return nil, err
	}

	cmd.Stdin = bytes.NewReader(payload)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tool command failed: %v\nstderr: %s", err, stderr.String())
	}

	var result ReviewResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("parse tool stdout: %v\nstdout: %s", err, stdout.String())
	}
	return &result, nil
}

// MockTool is a deterministic in-process reviewer used for local development and tests.
type MockTool struct{}

func (MockTool) Review(_ context.Context, pr *PRContext) (*ReviewResult, error) {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("### Mock review for #%d\n", pr.Number))
	builder.WriteString(fmt.Sprintf("- %d changed file(s)\n", len(pr.ChangedFiles)))
	if len(pr.Labels) > 0 {
		builder.WriteString(fmt.Sprintf("- Labels: %s\n", strings.Join(pr.Labels, ", ")))
	}
	if pr.GuidelinesB64 != "" {
		builder.WriteString("- Guidelines attached\n")
	}

	comments := make([]ReviewComment, 0, len(pr.ChangedFiles))
	for _, file := range pr.ChangedFiles {
		if strings.Contains(file.Patch, "fmt.Print") {
			comments = append(comments, ReviewComment{
				Path: file.Path,
				Body: "Mock LLM: consider removing debug prints before merging.",
				Line: 1,
				Side: "RIGHT",
			})
		}
	}
	if len(comments) == 0 && len(pr.ChangedFiles) > 0 {
		comments = append(comments, ReviewComment{
			Path: pr.ChangedFiles[0].Path,
			Body: "Mock LLM: file reviewed automatically; no blocking issues detected.",
			Line: 1,
			Side: "RIGHT",
		})
	}

	return &ReviewResult{
		SummaryMarkdown: builder.String(),
		Comments:        comments,
		ReviewDecision:  "comment",
		ReviewBody:      "Automated mock review",
	}, nil
}
