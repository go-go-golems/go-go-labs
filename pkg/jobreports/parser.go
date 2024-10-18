package jobreports

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// ParseJobReport parses job report data from a Reader and returns a ReportData struct.
func ParseJobReport(r io.Reader) (*ReportData, error) {
	scanner := bufio.NewScanner(r)
	reportData := &ReportData{
		Jobs:        make([]*Job, 0),
		LastUpdated: time.Now(),
	}

	// Skip the header line
	if scanner.Scan() {
		// Optionally, you could validate the header here
	}

	// Skip the separator line
	if scanner.Scan() {
		// Optionally, you could validate the separator here
	}

	for scanner.Scan() {
		line := scanner.Text()
		job, err := parseJobLine(line)
		if err != nil {
			return nil, fmt.Errorf("error parsing job line: %w", err)
		}
		if job == nil {
			continue
		}
		reportData.Jobs = append(reportData.Jobs, job)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	reportData.TotalJobs = len(reportData.Jobs)
	return reportData, nil
}

// parseJobLine parses a single line of job data and returns a Job struct.
func parseJobLine(line string) (*Job, error) {
	fields := strings.Fields(line)
	if len(fields) != 13 {
		return nil, nil
	}

	job := &Job{}

	var err error
	job.ID = fields[0]
	job.User = fields[1]
	job.Account = fields[2]
	job.Partition = fields[3]
	job.Status = JobStatus(fields[4])

	job.StartTime, err = time.Parse("2006-01-02", fields[5])
	if err != nil {
		return nil, fmt.Errorf("error parsing start time: %w", err)
	}

	job.WallTime, err = parseDuration(fields[6])
	if err != nil {
		return nil, fmt.Errorf("error parsing wall time: %w", err)
	}

	job.RunTime, err = parseDuration(fields[7])
	if err != nil {
		return nil, fmt.Errorf("error parsing run time: %w", err)
	}

	resources := strings.Split(fields[8], ",")
	if len(resources) != 3 {
		return nil, fmt.Errorf("invalid resource field: %s", fields[8])
	}

	job.CPUs, err = strconv.Atoi(resources[0])
	if err != nil {
		return nil, fmt.Errorf("error parsing CPUs: %w", err)
	}

	job.RAM, err = strconv.ParseFloat(strings.TrimSuffix(resources[1], "GB"), 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing RAM: %w", err)
	}

	job.GPUs, err = strconv.Atoi(resources[2])
	if err != nil {
		return nil, fmt.Errorf("error parsing GPUs: %w", err)
	}

	job.PendingTime, err = parseDuration(fields[9])
	if err != nil {
		return nil, fmt.Errorf("error parsing pending time: %w", err)
	}

	job.CPUEfficiency, err = parsePercentage(fields[10])
	if err != nil {
		return nil, fmt.Errorf("error parsing CPU efficiency: %w", err)
	}

	job.RAMEfficiency, err = parsePercentage(fields[11])
	if err != nil {
		return nil, fmt.Errorf("error parsing RAM efficiency: %w", err)
	}

	job.WallTimeEfficiency, err = parsePercentage(fields[12])
	if err != nil {
		return nil, fmt.Errorf("error parsing wall time efficiency: %w", err)
	}

	return job, nil
}

// parseDuration parses a duration string in the format "X.Y" and returns a time.Duration.
func parseDuration(s string) (time.Duration, error) {
	hours, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(hours * float64(time.Hour)), nil
}

// parsePercentage parses a percentage string and returns a float64.
func parsePercentage(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// ParseJobReportFile parses a job report file and returns a ReportData struct.
func ParseJobReportFile(filename string) (*ReportData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	return ParseJobReport(file)
}
