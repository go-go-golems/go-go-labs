package snakemake

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// TokenType defines the type of token identified by the tokenizer.
type TokenType string

const (
	TokenUnknown      TokenType = "UNKNOWN"
	TokenDate         TokenType = "DATE"
	TokenJobStart     TokenType = "JOB_START"
	TokenJobEnd       TokenType = "JOB_END"
	TokenJobID        TokenType = "JOB_ID"
	TokenWildcards    TokenType = "WILDCARDS"
	TokenResources    TokenType = "RESOURCES"
	TokenJobSubmitted TokenType = "JOB_SUBMITTED"
	TokenConfigInfo   TokenType = "CONFIG_INFO"
	TokenDAGBuilding  TokenType = "DAG_BUILDING"
	TokenJobStats     TokenType = "JOB_STATS"
	TokenJobSelection TokenType = "JOB_SELECTION"
	TokenEOF          TokenType = "EOF"
	TokenInput        TokenType = "INPUT"
	TokenOutput       TokenType = "OUTPUT"
	TokenReason       TokenType = "REASON"
	TokenThreads      TokenType = "THREADS"
	TokenGenericPair  TokenType = "GENERIC_PAIR"
	TokenScannerError TokenType = "SCANNER_ERROR"
)

// Token represents a single token identified by the tokenizer.
type Token struct {
	Type    TokenType
	Content string
}

// TokenData is an interface for structured token data
type TokenData interface {
	ToHash() map[string]interface{}
}

// DateData represents the parsed date information
type DateData struct {
	Time time.Time
}

func (d DateData) ToHash() map[string]interface{} {
	return map[string]interface{}{
		"time": d.Time,
	}
}

// JobStartData represents the parsed job start information
type JobStartData struct {
	RuleName string
}

func (j JobStartData) ToHash() map[string]interface{} {
	return map[string]interface{}{
		"ruleName": j.RuleName,
	}
}

// JobEndData represents the parsed job end information
type JobEndData struct {
	JobID int
}

func (j JobEndData) ToHash() map[string]interface{} {
	return map[string]interface{}{
		"jobID": j.JobID,
	}
}

// JobIDData represents the parsed job ID information
type JobIDData struct {
	ID int
}

func (j JobIDData) ToHash() map[string]interface{} {
	return map[string]interface{}{
		"id": j.ID,
	}
}

// WildcardsData represents the parsed wildcards information
type WildcardsData struct {
	Wildcards map[string]string
}

func (w WildcardsData) ToHash() map[string]interface{} {
	return map[string]interface{}{
		"wildcards": w.Wildcards,
	}
}

// ResourcesData represents the parsed resources information
type ResourcesData struct {
	Resources map[string]string
}

func (r ResourcesData) ToHash() map[string]interface{} {
	return map[string]interface{}{
		"resources": r.Resources,
	}
}

// JobSubmittedData represents the parsed job submitted information
type JobSubmittedData struct {
	JobID      int
	ExternalID string
}

func (j JobSubmittedData) ToHash() map[string]interface{} {
	return map[string]interface{}{
		"jobID":      j.JobID,
		"externalID": j.ExternalID,
	}
}

// JobStatsData represents the parsed job stats information
type JobStatsData struct {
	Stats map[string]int
}

func (j JobStatsData) ToHash() map[string]interface{} {
	return map[string]interface{}{
		"stats": j.Stats,
	}
}

// EmptyData represents an empty data structure for tokens that don't carry additional information
type EmptyData struct{}

func (e EmptyData) ToHash() map[string]interface{} {
	return map[string]interface{}{}
}

// InputData represents the parsed input information
type InputData struct {
	Inputs []string
}

func (i InputData) ToHash() map[string]interface{} {
	return map[string]interface{}{
		"inputs": i.Inputs,
	}
}

// OutputData represents the parsed output information
type OutputData struct {
	Outputs []string
}

func (o OutputData) ToHash() map[string]interface{} {
	return map[string]interface{}{
		"outputs": o.Outputs,
	}
}

// ReasonData represents the parsed reason information
type ReasonData struct {
	Reason string
}

func (r ReasonData) ToHash() map[string]interface{} {
	return map[string]interface{}{
		"reason": r.Reason,
	}
}

// ThreadsData represents the parsed threads information
type ThreadsData struct {
	Threads int
}

func (t ThreadsData) ToHash() map[string]interface{} {
	return map[string]interface{}{
		"threads": t.Threads,
	}
}

// GenericPairData represents the parsed generic pair information
type GenericPairData struct {
	Field string
	Value string
}

func (g GenericPairData) ToHash() map[string]interface{} {
	return map[string]interface{}{
		"field": g.Field,
		"value": g.Value,
	}
}

// Tokenizer is responsible for breaking the log into tokens.
type Tokenizer struct {
	scanner *bufio.Scanner
	file    *os.File
	buffer  *Token
	debug   bool
	logger  *log.Logger
}

// NewTokenizer initializes and returns a new Tokenizer.
func NewTokenizer(filename string, debug bool) (*Tokenizer, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	filteredReader := &lineLimitReader{
		reader:     bufio.NewReader(file),
		maxLineLen: 500,
	}

	return &Tokenizer{
		scanner: bufio.NewScanner(filteredReader),
		file:    file,
		buffer:  nil,
		debug:   debug,
		logger:  log.New(os.Stdout, "Tokenizer: ", log.Ldate|log.Ltime|log.Lshortfile),
	}, nil
}

// lineLimitReader is a custom io.Reader that truncates long lines
type lineLimitReader struct {
	reader     *bufio.Reader
	maxLineLen int
}

// Read implements the io.Reader interface
func (r *lineLimitReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	var readBytes int
	for readBytes < len(p) {
		line, isPrefix, err := r.reader.ReadLine()
		if err != nil {
			if err == io.EOF && readBytes > 0 {
				return readBytes, nil
			}
			return readBytes, err
		}

		if len(line) > r.maxLineLen {
			line = append(line[:r.maxLineLen-3], []byte("...")...)
			isPrefix = false
		}

		copyLen := copy(p[readBytes:], line)
		readBytes += copyLen

		if !isPrefix {
			if readBytes < len(p) {
				p[readBytes] = '\n'
				readBytes++
			}
			break
		}
	}

	return readBytes, nil
}

// Close closes the file associated with the Tokenizer.
func (t *Tokenizer) Close() error {
	return t.file.Close()
}

// debugLog logs a debug message if debug mode is enabled.
func (t *Tokenizer) debugLog(format string, v ...interface{}) {
	if t.debug {
		t.logger.Printf(format, v...)
	}
}

// NextToken returns the next token from the log.
func (t *Tokenizer) NextToken() (Token, TokenData, error) {
	if t.buffer != nil {
		token := *t.buffer
		t.buffer = nil
		t.debugLog("Returning buffered token: %+v", token)
		return token, nil, nil
	}

	for t.scanner.Scan() {
		line := t.scanner.Text()
		trimmed := strings.TrimSpace(line)
		t.debugLog("Processing line: %.80s", line)

		switch {
		// Example: [Mon Oct 14 10:13:42 2024]
		case t.isDateLine(trimmed):
			t.debugLog("Identified TokenDate")
			dateToken := Token{Type: TokenDate, Content: line}
			parsedTime, _ := time.Parse("[Mon Jan 2 15:04:05 2006]", line)
			return dateToken, DateData{Time: parsedTime}, nil

		// Example: rule rsem:
		case strings.HasPrefix(trimmed, "rule "):
			t.debugLog("Identified TokenJobStart")
			ruleName := strings.TrimPrefix(trimmed, "rule ")
			ruleName = strings.TrimSuffix(ruleName, ":")
			return Token{Type: TokenJobStart, Content: line}, JobStartData{RuleName: ruleName}, nil

		// Example: Finished job 433.
		case strings.HasPrefix(trimmed, "Finished job"):
			t.debugLog("Identified TokenJobEnd")
			parts := strings.Fields(trimmed)
			jobID, _ := strconv.Atoi(strings.TrimSuffix(parts[2], "."))
			return Token{Type: TokenJobEnd, Content: line}, JobEndData{JobID: jobID}, nil

		// Example: jobid: 433
		case strings.HasPrefix(trimmed, "jobid:"):
			t.debugLog("Identified TokenJobID")
			parts := strings.Fields(trimmed)
			jobID, _ := strconv.Atoi(parts[1])
			return Token{Type: TokenJobID, Content: line}, JobIDData{ID: jobID}, nil

		// Example: wildcards: name=S3
		case strings.HasPrefix(trimmed, "wildcards:"):
			t.debugLog("Identified TokenWildcards")
			wildcards := make(map[string]string)
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				for _, pair := range strings.Split(parts[1], ",") {
					kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
					if len(kv) == 2 {
						wildcards[kv[0]] = kv[1]
					}
				}
			}
			return Token{Type: TokenWildcards, Content: line}, WildcardsData{Wildcards: wildcards}, nil

		// Example: resources: mem_mb=13648, mem_mib=13016, disk_mb=13648, disk_mib=13016, tmpdir=<TBD>
		case strings.HasPrefix(trimmed, "resources:"):
			t.debugLog("Identified TokenResources")
			resources := make(map[string]string)
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				for _, pair := range strings.Split(parts[1], ",") {
					kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
					if len(kv) == 2 {
						resources[kv[0]] = kv[1]
					}
				}
			}
			return Token{Type: TokenResources, Content: line}, ResourcesData{Resources: resources}, nil

		// Example: Submitted job 433 with external jobid 'Submitted batch job 49368050'.
		case strings.HasPrefix(trimmed, "Submitted job"):
			t.debugLog("Identified TokenJobSubmitted")
			parts := strings.Fields(trimmed)
			jobID, _ := strconv.Atoi(parts[2])
			externalID := strings.Trim(parts[len(parts)-1], "'.")
			return Token{Type: TokenJobSubmitted, Content: line}, JobSubmittedData{JobID: jobID, ExternalID: externalID}, nil

		// Example: Config file config.json is extended by additional config specified via the command line.
		case strings.HasPrefix(trimmed, "Config file"):
			t.debugLog("Identified TokenConfigInfo")
			return Token{Type: TokenConfigInfo, Content: line}, EmptyData{}, nil

		// Example: Building DAG of jobs...
		case strings.HasPrefix(trimmed, "Building DAG"):
			t.debugLog("Identified TokenDAGBuilding")
			return Token{Type: TokenDAGBuilding, Content: line}, EmptyData{}, nil

		// Example: Job stats:
		case strings.HasPrefix(trimmed, "Job stats:"):
			t.debugLog("Identified TokenJobStats")
			jobStats := t.parseJobStats()
			return Token{Type: TokenJobStats, Content: line}, JobStatsData{Stats: jobStats}, nil

		// Example: Select jobs to execute...
		case strings.HasPrefix(trimmed, "Select jobs to execute"):
			t.debugLog("Identified TokenJobSelection")
			return Token{Type: TokenJobSelection, Content: line}, EmptyData{}, nil

		case strings.HasPrefix(trimmed, "input:"):
			t.debugLog("Identified TokenInput")
			inputs := strings.Split(strings.TrimPrefix(trimmed, "input:"), ",")
			for i, input := range inputs {
				inputs[i] = strings.TrimSpace(input)
			}
			return Token{Type: TokenInput, Content: line}, InputData{Inputs: inputs}, nil

		case strings.HasPrefix(trimmed, "output:"):
			t.debugLog("Identified TokenOutput")
			outputs := strings.Split(strings.TrimPrefix(trimmed, "output:"), ",")
			for i, output := range outputs {
				outputs[i] = strings.TrimSpace(output)
			}
			return Token{Type: TokenOutput, Content: line}, OutputData{Outputs: outputs}, nil

		case strings.HasPrefix(trimmed, "reason:"):
			t.debugLog("Identified TokenReason")
			reason := strings.TrimSpace(strings.TrimPrefix(trimmed, "reason:"))
			return Token{Type: TokenReason, Content: line}, ReasonData{Reason: reason}, nil

		case strings.HasPrefix(trimmed, "threads:"):
			t.debugLog("Identified TokenThreads")
			threadsStr := strings.TrimSpace(strings.TrimPrefix(trimmed, "threads:"))
			threads, _ := strconv.Atoi(threadsStr)
			return Token{Type: TokenThreads, Content: line}, ThreadsData{Threads: threads}, nil

		case strings.Contains(trimmed, ":"):
			t.debugLog("Identified TokenGenericPair")
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				field := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				return Token{Type: TokenGenericPair, Content: line}, GenericPairData{Field: field, Value: value}, nil
			}

		default:
			t.debugLog("Unrecognized line type")
		}
	}

	if err := t.scanner.Err(); err != nil {
		t.debugLog("Scanner error: %v", err)
		return Token{Type: TokenScannerError, Content: err.Error()}, EmptyData{}, nil
	}

	t.debugLog("Reached end of file")
	return Token{Type: TokenEOF, Content: ""}, EmptyData{}, nil
}

// isDateLine determines if a line is a date line.
func (t *Tokenizer) isDateLine(line string) bool {
	_, err := time.Parse("[Mon Jan 2 15:04:05 2006]", line)
	return err == nil
}

// isJobDetailLine determines if a line is part of job details.
func (t *Tokenizer) isJobDetailLine(line string) bool {
	return strings.HasPrefix(line, "input:") ||
		strings.HasPrefix(line, "output:") ||
		strings.HasPrefix(line, "reason:") ||
		strings.HasPrefix(line, "threads:") ||
		strings.Contains(line, ":") // This catches any "FIELD: VALUE" pair
}

// parseJobStats parses the job stats section
func (t *Tokenizer) parseJobStats() map[string]int {
	jobStats := make(map[string]int)
	for t.scanner.Scan() {
		line := strings.TrimSpace(t.scanner.Text())
		if line == "" {
			break
		}
		parts := strings.Fields(line)
		if len(parts) == 2 {
			jobStats[parts[0]] = parseInt(parts[1])
		}
	}
	return jobStats
}

// parseInt converts a string to an integer, returning 0 if conversion fails
func parseInt(s string) int {
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return value
}