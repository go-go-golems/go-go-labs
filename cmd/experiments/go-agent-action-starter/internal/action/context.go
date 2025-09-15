package action

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v66/github"
)

// PullRequestClient exposes the GitHub operations required to build context.
type PullRequestClient interface {
	GetPullRequest(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error)
	ListPullRequestFiles(ctx context.Context, owner, repo string, number int, opts *github.ListOptions) ([]*github.CommitFile, *github.Response, error)
}

// FileLoader mirrors os.ReadFile so tests can stub filesystem access.
type FileLoader func(string) ([]byte, error)

// CollectPRContext gathers PR metadata, file diffs, and repo context.
func CollectPRContext(ctx context.Context, gh PullRequestClient, in *Inputs, env RuntimeEnv, readFile FileLoader) (*PRContext, error) {
	if readFile == nil {
		readFile = os.ReadFile
	}

	owner, repo := splitRepo(env.Repository)
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("invalid repository value %q", env.Repository)
	}

	number, triggerText, err := determinePRNumber(env, readFile)
	if err != nil {
		return nil, err
	}

	pr, err := gh.GetPullRequest(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("github get pull request: %w", err)
	}

	labels := make([]string, 0, len(pr.Labels))
	for _, l := range pr.Labels {
		if l != nil && l.Name != nil {
			labels = append(labels, *l.Name)
		}
	}

	assignees := make([]string, 0, len(pr.Assignees))
	for _, a := range pr.Assignees {
		if a != nil {
			assignees = append(assignees, a.GetLogin())
		}
	}

	changed, err := listChangedFiles(ctx, gh, owner, repo, number, in, env.Workspace, readFile)
	if err != nil {
		return nil, err
	}

	extras := loadExtraFiles(env.Workspace, in.IncludeRepoGlobs, in.MaxFileBytes, readFile)

	guidelines := ""
	if in.GuidelinesPath != "" {
		guidelines = readWorkspaceFileB64(env.Workspace, in.GuidelinesPath, in.MaxFileBytes, readFile)
	}

	return &PRContext{
		Owner:         owner,
		Repo:          repo,
		Number:        number,
		Title:         value(pr.Title),
		Body:          value(pr.Body),
		BaseRef:       pr.GetBase().GetRef(),
		HeadRef:       pr.GetHead().GetRef(),
		HeadSHA:       pr.GetHead().GetSHA(),
		UserLogin:     pr.GetUser().GetLogin(),
		Labels:        labels,
		Assignees:     assignees,
		ChangedFiles:  changed,
		GuidelinesB64: guidelines,
		ExtraFiles:    extras,
		TriggeredBy:   env.Actor,
		EventName:     env.EventName,
		TriggerText:   triggerText,
		RunID:         env.RunID,
	}, nil
}

func determinePRNumber(env RuntimeEnv, readFile FileLoader) (int, string, error) {
	var payload struct {
		Action string `json:"action"`
		Issue  *struct {
			Number      int       `json:"number"`
			PullRequest *struct{} `json:"pull_request,omitempty"`
		} `json:"issue"`
		Comment *struct {
			Body string `json:"body"`
		} `json:"comment"`
		PullRequest *struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
			Body   string `json:"body"`
		} `json:"pull_request"`
	}
	if env.EventPath != "" {
		if data, err := readFile(env.EventPath); err == nil {
			_ = json.Unmarshal(data, &payload)
		}
	}

	switch env.EventName {
	case "issue_comment", "pull_request_review_comment":
		if payload.Issue != nil && payload.Issue.PullRequest != nil {
			body := ""
			if payload.Comment != nil {
				body = payload.Comment.Body
			}
			return payload.Issue.Number, body, nil
		}
	case "pull_request":
		if payload.PullRequest != nil {
			return payload.PullRequest.Number, "", nil
		}
	case "workflow_dispatch", "repository_dispatch":
		// Manual runs can pass a PR number via tool args in future.
	}

	return 0, "", fmt.Errorf("unable to determine pull request number for event %q", env.EventName)
}

func listChangedFiles(ctx context.Context, gh PullRequestClient, owner, repo string, number int, in *Inputs, workspace string, readFile FileLoader) ([]ChangedFile, error) {
	var all []*github.CommitFile
	opt := &github.ListOptions{PerPage: 100}
	for {
		files, resp, err := gh.ListPullRequestFiles(ctx, owner, repo, number, opt)
		if err != nil {
			return nil, fmt.Errorf("github list files: %w", err)
		}
		all = append(all, files...)
		if resp == nil || resp.NextPage == 0 || len(all) >= in.MaxChangedFiles {
			break
		}
		opt.Page = resp.NextPage
	}

	if len(all) > in.MaxChangedFiles {
		all = all[:in.MaxChangedFiles]
	}

	changed := make([]ChangedFile, 0, len(all))
	for _, f := range all {
		cf := ChangedFile{
			Path:      value(f.Filename),
			Status:    value(f.Status),
			Additions: f.GetAdditions(),
			Deletions: f.GetDeletions(),
			BlobURL:   value(f.BlobURL),
			RawURL:    value(f.RawURL),
		}
		if in.IncludePatch {
			cf.Patch = value(f.Patch)
		}
		if in.IncludeFileContent && cf.Status != "removed" {
			cf.ContentsB64 = readWorkspaceFileB64(workspace, cf.Path, in.MaxFileBytes, readFile)
		}
		changed = append(changed, cf)
	}
	return changed, nil
}

func loadExtraFiles(workspace string, globs []string, limit int, readFile FileLoader) []RepoFile {
	if len(globs) == 0 {
		return nil
	}
	var files []RepoFile
	for _, pattern := range globs {
		if pattern == "" {
			continue
		}
		matches, _ := filepath.Glob(filepath.Join(workspace, pattern))
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}
			data, err := readFile(match)
			if err != nil {
				continue
			}
			if limit > 0 && len(data) > limit {
				data = data[:limit]
			}
			rel := strings.TrimPrefix(match, workspace+string(filepath.Separator))
			files = append(files, RepoFile{
				Path:        rel,
				ContentsB64: base64.StdEncoding.EncodeToString(data),
			})
		}
	}
	return files
}

func readWorkspaceFileB64(workspace, rel string, limit int, readFile FileLoader) string {
	if rel == "" {
		return ""
	}
	path := filepath.Join(workspace, rel)
	data, err := readFile(path)
	if err != nil {
		return ""
	}
	if limit > 0 && len(data) > limit {
		data = data[:limit]
	}
	return base64.StdEncoding.EncodeToString(data)
}

func value(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func splitRepo(full string) (string, string) {
	parts := strings.Split(full, "/")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}
