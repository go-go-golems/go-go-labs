package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"time"
)

type prContext struct {
	Number       int           `json:"number"`
	ChangedFiles []changedFile `json:"changed_files"`
}

type changedFile struct {
	Path string `json:"path"`
}

type reviewResult struct {
	SummaryMarkdown string          `json:"summary_markdown"`
	Comments        []reviewComment `json:"comments"`
	ReviewDecision  string          `json:"review_decision"`
	ReviewBody      string          `json:"review_body"`
}

type reviewComment struct {
	Path string `json:"path"`
	Body string `json:"body"`
	Line int    `json:"line,omitempty"`
	Side string `json:"side,omitempty"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var ctx prContext
	if err := json.NewDecoder(os.Stdin).Decode(&ctx); err != nil {
		panic(err)
	}

	commands := []string{
		"go test ./...",
		"go vet ./...",
		"golangci-lint run",
	}

	cmd := commands[rand.Intn(len(commands))]

	summary := "### Random review\n- suggested command: `" + cmd + "`"

	result := reviewResult{
		SummaryMarkdown: summary,
		ReviewDecision:  "comment",
		ReviewBody:      "Random command reviewer",
	}

	if len(ctx.ChangedFiles) > 0 {
		result.Comments = []reviewComment{
			{
				Path: ctx.ChangedFiles[0].Path,
				Body: "Please run `" + cmd + "` and share the output.",
				Line: 1,
				Side: "RIGHT",
			},
		}
	}

	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		panic(err)
	}
}
