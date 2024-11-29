package utils

import "time"

// TextractJob represents a Textract processing job
type TextractJob struct {
	JobID       string     `json:"job_id" dynamodbav:"JobID"`
	DocumentKey string     `json:"document_key" dynamodbav:"DocumentKey"`
	Status      string     `json:"status" dynamodbav:"Status"`
	SubmittedAt time.Time  `json:"submitted_at" dynamodbav:"SubmittedAt"`
	CompletedAt *time.Time `json:"completed_at,omitempty" dynamodbav:"CompletedAt,omitempty"`
	TextractID  string     `json:"textract_id" dynamodbav:"TextractID"`
	ResultKey   string     `json:"result_key" dynamodbav:"ResultKey"`
	Error       string     `json:"error,omitempty" dynamodbav:"Error,omitempty"`
}
