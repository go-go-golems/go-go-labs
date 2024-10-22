package snakemake

import (
	"time"
	"github.com/go-go-golems/go-go-labs/pkg/jobreports"
)

// Resource represents a resource with a name and value.
type Resource struct {
	Name  string
	Value string
}

// JobStatus represents the status of a job.
type JobStatus string

const (
	StatusInProgress JobStatus = "In Progress"
	StatusCompleted  JobStatus = "Completed"
)

// Job represents a Snakemake job with its details.
type Job struct {
	ID                 string
	Rule               string
	StartTime          time.Time
	EndTime            time.Time
	Duration           time.Duration
	Status             JobStatus
	Input              []string
	Output             []string
	Reason             string
	Threads            int
	Details            map[string]string
	Resources          []Resource
	Wildcards          map[string]string
	ExternalID         string
	ScannerError       string
	JobReport          *jobreports.Job
}

// Rule represents a Snakemake rule with its associated jobs and resources.
type Rule struct {
	Name      string
	Jobs      []*Job
	Resources []Resource
}

// LogData holds the parsed data from the Snakemake log.
type LogData struct {
	Rules       map[string]*Rule
	Jobs        []*Job
	FullLog     string
	TotalJobs   int
	Completed   int
	InProgress  int
	LastUpdated time.Time
	JobStats    map[string]int
}
