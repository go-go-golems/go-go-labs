package models

import (
	"time"
)

// StreamInfo represents the stream information data structure
type StreamInfo struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"startTime"`
	Language    string    `json:"language"`
	GithubRepo  string    `json:"githubRepo"`
	ViewerCount int       `json:"viewerCount"`
}

// Step represents a task step in the stream
type Step struct {
	ID        int    `json:"id"`
	Content   string `json:"content"`
	Status    string `json:"status"` // "completed", "active", or "upcoming"
	CreatedAt time.Time `json:"createdAt"`
}
