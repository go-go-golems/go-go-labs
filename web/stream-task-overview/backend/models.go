package main

import (
	"time"
)

// StreamInfo represents stream metadata
type StreamInfo struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"startTime"`
	Language    string    `json:"language"`
	GithubRepo  string    `json:"githubRepo"`
	ViewerCount int       `json:"viewerCount"`
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