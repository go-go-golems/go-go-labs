package action

import (
	"context"
	"fmt"
	"io"
	"os"
)

type githubService interface {
	PullRequestClient
	ReviewPublisher
}

// Runner coordinates the full lifecycle of the action.
type Runner struct {
	Inputs     *Inputs
	Env        RuntimeEnv
	GitHub     githubService
	Tool       ReviewTool
	FileLoader FileLoader
	Publisher  *Publisher
	Logger     io.Writer
}

// Run executes the review flow: gather context, gate on triggers, call the tool, publish outputs.
func (r *Runner) Run(ctx context.Context) error {
	if r.Inputs == nil {
		return fmt.Errorf("inputs required")
	}
	if r.GitHub == nil {
		return fmt.Errorf("github client required")
	}
	if r.Tool == nil {
		return fmt.Errorf("review tool required")
	}
	if r.FileLoader == nil {
		r.FileLoader = os.ReadFile
	}
	if r.Logger == nil {
		r.Logger = os.Stdout
	}
	if r.Publisher == nil {
		r.Publisher = NewPublisher(r.GitHub, r.Inputs, r.Env)
	}

	prc, err := CollectPRContext(ctx, r.GitHub, r.Inputs, r.Env, r.FileLoader)
	if err != nil {
		return err
	}

	if !ShouldTrigger(r.Inputs, prc) {
		fmt.Fprintf(r.Logger, "no review triggers matched; skipping for PR #%d\n", prc.Number)
		return nil
	}

	result, err := r.Tool.Review(ctx, prc)
	if err != nil {
		return fmt.Errorf("review tool: %w", err)
	}

	if err := r.Publisher.Publish(ctx, prc, result); err != nil {
		return err
	}

	modes := parseOutputModes(r.Inputs.OutputMode)
	if result.SummaryMarkdown != "" && !modes["stdout"] {
		fmt.Fprintln(r.Publisher.Stdout, result.SummaryMarkdown)
	}

	return nil
}
