package jobreports

import (
	"time"
)

// JobStatus represents the status of a job.
type JobStatus string

const (
	StatusCompleted JobStatus = "COMPLETED"
	// Add other statuses as needed
)

// Resource represents a resource with a name and value.
type Resource struct {
	Name  string
	Value float64
}

// Job represents a job report entry with its details.
type Job struct {
	ID                 string
	User               string
	Account            string
	Partition          string
	Status             JobStatus
	StartTime          time.Time
	WallTime           time.Duration
	RunTime            time.Duration
	CPUs               int
	RAM                float64 // in GB
	GPUs               int
	PendingTime        time.Duration
	CPUEfficiency      float64 // in percentage
	RAMEfficiency      float64 // in percentage
	WallTimeEfficiency float64 // in percentage
}

// ReportData holds the parsed data from the job report.
type ReportData struct {
	Jobs        []*Job
	TotalJobs   int
	LastUpdated time.Time
}
