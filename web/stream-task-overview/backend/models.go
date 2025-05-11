package main

import (
	"time"
)

// StreamInfo represents stream metadata
type StreamInfo struct {
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	StartTime   time.Time `json:"startTime" db:"start_time"`
	Language    string    `json:"language" db:"language"`
	GithubRepo  string    `json:"githubRepo" db:"github_repo"`
	ViewerCount int       `json:"viewerCount" db:"viewer_count"`
}

// StepInfo represents all task steps
type StepInfo struct {
	Completed []string `json:"completed"`
	Active    string   `json:"active"`
	Upcoming  []string `json:"upcoming"`
}

// Stream represents the complete stream state
type Stream struct {
	Info  StreamInfo `json:"info"`
	Steps StepInfo   `json:"steps"`
}