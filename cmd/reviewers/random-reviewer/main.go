package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"time"
)

type prContext struct {
	Number int `json:"number"`
}

type reviewResult struct {
	SummaryMarkdown string `json:"summary_markdown"`
	ReviewDecision  string `json:"review_decision"`
	ReviewBody      string `json:"review_body"`
	IssueComment    string `json:"issue_comment"`
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
	body := "Random command reviewer: please run `" + cmd + "` and share the output."

	result := reviewResult{
		SummaryMarkdown: summary,
		ReviewDecision:  "comment",
		ReviewBody:      body,
		IssueComment:    body,
	}

	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		panic(err)
	}
}
