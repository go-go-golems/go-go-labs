package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type prContext struct {
	Number       int    `json:"number"`
	Body         string `json:"body"`
	TriggerText  string `json:"trigger_text"`
	EventName    string `json:"event_name"`
	ChangedFiles []struct {
		Path string `json:"path"`
	} `json:"changed_files"`
}

type reviewResult struct {
	SummaryMarkdown string          `json:"summary_markdown"`
	ReviewDecision  string          `json:"review_decision"`
	ReviewBody      string          `json:"review_body"`
	IssueComment    string          `json:"issue_comment"`
	Comments        []reviewComment `json:"comments"`
}

type reviewComment struct {
	Path    string `json:"path"`
	Body    string `json:"body"`
	Subject string `json:"subject_type,omitempty"`
}

func main() {
	var ctx prContext
	if err := json.NewDecoder(os.Stdin).Decode(&ctx); err != nil {
		panic(err)
	}

	commands := gatherCommands(ctx)
	if len(commands) == 0 {
		commands = []string{"go test ./...", "go vet ./...", "golangci-lint run"}
	}

	summary := buildSummary(commands)
	body := buildBody(commands)

	res := reviewResult{
		SummaryMarkdown: summary,
		ReviewDecision:  "comment",
		ReviewBody:      body,
	}

	if ctx.EventName == "issue_comment" {
		res.IssueComment = body
	}

	if ctx.EventName == "pull_request" && len(ctx.ChangedFiles) > 0 {
		res.Comments = buildFileComments(commands, ctx.ChangedFiles)
	}

	if err := json.NewEncoder(os.Stdout).Encode(res); err != nil {
		panic(err)
	}
}

func gatherCommands(ctx prContext) []string {
	seen := map[string]struct{}{}
	var commands []string

	add := func(list []string) {
		for _, cmd := range list {
			trimmed := strings.TrimSpace(cmd)
			if trimmed == "" {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			commands = append(commands, trimmed)
		}
	}

	add(extractCommandsFromBody(ctx.Body))
	add(extractCommandsFromBody(ctx.TriggerText))
	add(extractRunLines(ctx.Body))
	add(extractRunLines(ctx.TriggerText))

	return commands
}

func extractCommandsFromBody(text string) []string {
	const marker = "```agent"
	var cmds []string
	lower := strings.ToLower(text)
	search := 0
	for {
		idx := strings.Index(lower[search:], marker)
		if idx == -1 {
			break
		}
		idx += search
		start := idx + len(marker)
		// skip newline characters
		for start < len(text) && (text[start] == '\n' || text[start] == '\r') {
			start++
		}
		endIdx := strings.Index(lower[start:], "```")
		if endIdx == -1 {
			break
		}
		end := start + endIdx
		block := text[start:end]
		scanner := bufio.NewScanner(strings.NewReader(block))
		for scanner.Scan() {
			trimmed := strings.TrimSpace(scanner.Text())
			if trimmed != "" {
				cmds = append(cmds, trimmed)
			}
		}
		search = end + len("```")
	}
	return cmds
}

func extractRunLines(text string) []string {
	var cmds []string
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "@agent run ") {
			cmds = append(cmds, strings.TrimSpace(line[len("@agent run "):]))
		}
	}
	return cmds
}

func buildSummary(commands []string) string {
	if len(commands) == 0 {
		return "### Automated review\n- no commands requested"
	}
	lines := []string{"### Automated review"}
	for _, cmd := range commands {
		lines = append(lines, fmt.Sprintf("- run `%s`", cmd))
	}
	return strings.Join(lines, "\n")
}

func buildBody(commands []string) string {
	if len(commands) == 0 {
		return "No commands were provided for this review."
	}
	var b strings.Builder
	b.WriteString("Requested commands:\n\n")
	for _, cmd := range commands {
		b.WriteString("- `")
		b.WriteString(cmd)
		b.WriteString("`\n")
	}
	return b.String()
}

func buildFileComments(commands []string, files []struct {
	Path string `json:"path"`
}) []reviewComment {
	var comments []reviewComment
	if len(files) == 0 {
		return comments
	}
	for i, cmd := range commands {
		file := files[i%len(files)].Path
		msg := fmt.Sprintf("Please run `" + cmd + "` and update the PR with the results.")
		comments = append(comments, reviewComment{
			Path:    file,
			Body:    msg,
			Subject: "file",
		})
	}
	return comments
}
