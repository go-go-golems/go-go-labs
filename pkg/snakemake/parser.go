package snakemake

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Parser handles the parsing logic using tokens provided by the Tokenizer.
type Parser struct {
	tokenizer   *Tokenizer
	logData     LogData
	currentJob  *Job
	currentDate time.Time
	debug       bool
	logger      *log.Logger
}

// NewParser initializes and returns a new Parser.
func NewParser(tokenizer *Tokenizer, debug bool) *Parser {
	return &Parser{
		tokenizer: tokenizer,
		logData: LogData{
			Rules: make(map[string]*Rule),
		},
		currentJob: nil,
		debug:      debug,
		logger:     log.New(os.Stdout, "Parser: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// ParseLog parses the Snakemake log and returns structured LogData.
func (p *Parser) ParseLog() (LogData, error) {
	p.debugLog("Starting log parsing")
	for {
		token, data, err := p.tokenizer.NextToken()
		if err != nil {
			return LogData{}, err
		}

		if token.Type == TokenEOF {
			break
		}

		p.debugLog("Parsing token: %+v, Data hash: %s", token, data.ToHash())

		switch token.Type {
		case TokenDate:
			p.handleDate(data.(DateData))
		case TokenJobStart:
			p.handleJobStart(data.(JobStartData))
		case TokenJobEnd:
			p.handleJobEnd(data.(JobEndData))
		case TokenJobID:
			p.handleJobID(data.(JobIDData))
		case TokenWildcards:
			p.handleWildcards(data.(WildcardsData))
		case TokenResources:
			p.handleResources(data.(ResourcesData))
		case TokenJobSubmitted:
			p.handleJobSubmitted(data.(JobSubmittedData))
		case TokenJobStats:
			p.handleJobStats(data.(JobStatsData))
		case TokenConfigInfo, TokenDAGBuilding, TokenJobSelection:
			// These tokens don't require special handling, but we acknowledge them
			p.debugLog("Acknowledged token: %v", token.Type)
		case TokenInput:
			p.handleInput(data.(InputData))
		case TokenOutput:
			p.handleOutput(data.(OutputData))
		case TokenReason:
			p.handleReason(data.(ReasonData))
		case TokenThreads:
			p.handleThreads(data.(ThreadsData))
		case TokenGenericPair:
			p.handleGenericPair(data.(GenericPairData))
		case TokenScannerError:
			p.handleScannerError(token.Content)
			return p.logData, nil
		}
	}

	p.finalizeLogData()
	return p.logData, nil
}

// debugLog logs a debug message if debug mode is enabled.
func (p *Parser) debugLog(format string, v ...interface{}) {
	if p.debug {
		p.logger.Printf(format, v...)
	}
}

// handleDate processes a date token.
func (p *Parser) handleDate(data DateData) {
	p.debugLog("Handling date: %s", data.Time)
	p.currentDate = data.Time
	p.logData.LastUpdated = data.Time
}

// handleJobStart processes a job start token.
func (p *Parser) handleJobStart(data JobStartData) {
	p.debugLog("Handling job start: %s", data.RuleName)
	p.currentJob = &Job{
		Rule:      data.RuleName,
		Status:    StatusInProgress,
		StartTime: p.currentDate,
		Details:   make(map[string]string), // Initialize the Details map
	}
	p.logData.Jobs = append(p.logData.Jobs, p.currentJob)
}

// handleJobEnd processes a job end token.
func (p *Parser) handleJobEnd(data JobEndData) {
	p.debugLog("Handling job end: JobID %d", data.JobID)
	for _, job := range p.logData.Jobs {
		if job.ID == fmt.Sprintf("%d", data.JobID) {
			job.Status = StatusCompleted
			job.EndTime = p.currentDate
			job.Duration = job.EndTime.Sub(job.StartTime)
			return
		}
	}
}

// handleJobID processes a job ID token.
func (p *Parser) handleJobID(data JobIDData) {
	p.debugLog("Handling job ID: %d", data.ID)
	if p.currentJob != nil {
		p.currentJob.ID = fmt.Sprintf("%d", data.ID)
	}
}

// handleWildcards processes a wildcards token.
func (p *Parser) handleWildcards(data WildcardsData) {
	p.debugLog("Handling wildcards: %v", data.Wildcards)
	if p.currentJob != nil {
		p.currentJob.Wildcards = data.Wildcards
	}
}

// handleResources processes a resources token.
func (p *Parser) handleResources(data ResourcesData) {
	p.debugLog("Handling resources: %v", data.Resources)
	if p.currentJob != nil {
		for name, value := range data.Resources {
			p.currentJob.Resources = append(p.currentJob.Resources, Resource{Name: name, Value: value})
		}
	}
}

// handleJobSubmitted processes a job submitted token.
func (p *Parser) handleJobSubmitted(data JobSubmittedData) {
	p.debugLog("Handling job submitted: JobID %d, ExternalID %s", data.JobID, data.ExternalID)
	for _, job := range p.logData.Jobs {
		if job.ID == fmt.Sprintf("%d", data.JobID) {
			job.ExternalID = data.ExternalID
			return
		}
	}
}

// handleJobStats processes a job stats token.
func (p *Parser) handleJobStats(data JobStatsData) {
	p.debugLog("Handling job stats: %v", data.Stats)
	p.logData.JobStats = data.Stats
}

// handleInput processes an input token.
func (p *Parser) handleInput(data InputData) {
	if p.currentJob != nil {
		p.currentJob.Input = data.Inputs
	}
}

// handleOutput processes an output token.
func (p *Parser) handleOutput(data OutputData) {
	if p.currentJob != nil {
		p.currentJob.Output = data.Outputs
	}
}

// handleReason processes a reason token.
func (p *Parser) handleReason(data ReasonData) {
	if p.currentJob != nil {
		p.currentJob.Reason = data.Reason
	}
}

// handleThreads processes a threads token.
func (p *Parser) handleThreads(data ThreadsData) {
	if p.currentJob != nil {
		p.currentJob.Threads = data.Threads
	}
}

// handleGenericPair processes a generic pair token.
func (p *Parser) handleGenericPair(data GenericPairData) {
	if p.currentJob != nil {
		p.currentJob.Details[data.Field] = data.Value
	}
}

// finalizeLogData populates the Rules map and calculates job counts.
func (p *Parser) finalizeLogData() {
	p.debugLog("Finalizing log data")
	for _, job := range p.logData.Jobs {
		rule, exists := p.logData.Rules[job.Rule]
		if !exists {
			rule = &Rule{
				Name:      job.Rule,
				Jobs:      []*Job{},
				Resources: []Resource{},
			}
			p.logData.Rules[job.Rule] = rule
		}

		rule.Jobs = append(rule.Jobs, job)
		rule.Resources = mergeResources(rule.Resources, job.Resources)
	}

	p.logData.TotalJobs = len(p.logData.Jobs)
	for _, job := range p.logData.Jobs {
		if job.Status == StatusCompleted {
			p.logData.Completed++
		} else {
			p.logData.InProgress++
		}
	}

	p.debugLog("Parsing complete - Total Jobs: %d, Completed: %d, In Progress: %d, Jobs: %v",
		p.logData.TotalJobs, p.logData.Completed, p.logData.InProgress, len(p.logData.Jobs))
	p.debugLog("Total Rules: %d", len(p.logData.Rules))
	for ruleName, rule := range p.logData.Rules {
		p.debugLog("Rule: %s, Jobs: %d, Resources: %d",
			ruleName, len(rule.Jobs), len(rule.Resources))
	}
}

// Add this method to the Parser struct
func (p *Parser) handleScannerError(errorMessage string) {
	p.debugLog("Warning: Scanner error encountered: %s", errorMessage)
	if p.currentJob != nil {
		p.currentJob.ScannerError = errorMessage
	} else {
		// If there's no current job, create a new one to store the error
		p.currentJob = &Job{
			Rule:         "Unknown",
			Status:       StatusInProgress,
			ScannerError: errorMessage,
		}
		p.logData.Jobs = append(p.logData.Jobs, p.currentJob)
	}
}
