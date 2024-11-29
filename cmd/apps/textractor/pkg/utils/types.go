package utils

import "time"

type TextractJob struct {
	JobID       string     `json:"jobId"`
	DocumentKey string     `json:"documentKey"`
	Status      string     `json:"status"`
	SubmittedAt time.Time  `json:"submittedAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	Error       string     `json:"error,omitempty"`
}
