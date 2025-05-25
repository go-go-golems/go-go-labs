package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/rlimit"
	"github.com/spf13/cobra"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target amd64 -cflags "-I/usr/include/x86_64-linux-gnu -mllvm -bpf-stack-size=8192" sniffwrites sniff_writes.c

type Event struct {
	Pid      uint32
	Fd       int32
	Comm     [16]int8
	Filename [64]int8
	Type     uint32 // 0 = open, 1 = read, 2 = write, 3 = close
}

type Config struct {
	Directory    string
	OutputFormat string
	Operations   []string
	ProcessFilter string
	Duration     time.Duration
	Verbose      bool
	ShowFd       bool
	OutputFile   string
	Debug        bool
}

type EventOutput struct {
	Timestamp string `json:"timestamp"`
	Pid       uint32 `json:"pid"`
	Process   string `json:"process"`
	Operation string `json:"operation"`
	Filename  string `json:"filename"`
	Fd        int32  `json:"fd,omitempty"`
}

var config Config

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "sniff-writes",
	Short: "Monitor file operations using eBPF",
	Long: `sniff-writes monitors file operations (open, read, write, close) on specified 
directories using eBPF tracepoints. This is a Go port of a bpftrace script that 
provides more flexible filtering and output options.

Examples:
  # Monitor default directory with plain output
  sudo sniff-writes monitor

  # Monitor specific directory with JSON output for 30 seconds
  sudo sniff-writes monitor -d /var/log -f json -t 30s

  # Monitor only read/write operations with table format
  sudo sniff-writes monitor -o read,write -f table --show-fd

  # Filter by process name
  sudo sniff-writes monitor -p nginx -v`,
}

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Start monitoring file operations",
	Long:  "Start monitoring file operations on the specified directory using eBPF tracepoints.",
	RunE:  runMonitor,
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(monitorCmd)
	
	// Add flags to monitor command
	monitorCmd.Flags().StringVarP(&config.Directory, "directory", "d", "cmd/n8n-cli", "Directory to monitor")
	monitorCmd.Flags().StringVarP(&config.OutputFormat, "format", "f", "plain", "Output format: plain, json, table")
	monitorCmd.Flags().StringSliceVarP(&config.Operations, "operations", "o", []string{"open", "read", "write", "close"}, "Operations to monitor")
	monitorCmd.Flags().StringVarP(&config.ProcessFilter, "process", "p", "", "Filter by process name (substring match)")
	monitorCmd.Flags().DurationVarP(&config.Duration, "duration", "t", 0, "Duration to run (0 = infinite)")
	monitorCmd.Flags().BoolVarP(&config.Verbose, "verbose", "v", false, "Verbose output")
	monitorCmd.Flags().BoolVar(&config.ShowFd, "show-fd", false, "Show file descriptor numbers")
	monitorCmd.Flags().StringVar(&config.OutputFile, "output", "", "Output to file instead of stdout")
	monitorCmd.Flags().BoolVar(&config.Debug, "debug", false, "Debug mode - show all events regardless of filters")
}

func runMonitor(cmd *cobra.Command, args []string) error {
	// Check for root privileges
	if os.Geteuid() != 0 {
		return fmt.Errorf("this program requires root privileges to load eBPF programs")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Set up duration-based cancellation if specified
	if config.Duration > 0 {
		durationCtx, durationCancel := context.WithTimeout(ctx, config.Duration)
		defer durationCancel()
		ctx = durationCtx
	}

	// Remove memory limit for eBPF
	if err := rlimit.RemoveMemlock(); err != nil {
		return fmt.Errorf("failed to remove memlock limit: %w", err)
	}

	// Load pre-compiled programs and maps into the kernel
	spec, err := loadSniffwrites()
	if err != nil {
		return fmt.Errorf("failed to load eBPF spec: %w", err)
	}

	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		return fmt.Errorf("failed to create eBPF collection: %w", err)
	}
	defer coll.Close()

	// Attach tracepoints based on selected operations
	links, err := attachTracepoints(coll)
	if err != nil {
		return fmt.Errorf("failed to attach tracepoints: %w", err)
	}
	defer func() {
		for _, l := range links {
			l.Close()
		}
	}()

	// Open perf event reader
	rd, err := perf.NewReader(coll.Maps["events"], os.Getpagesize())
	if err != nil {
		return fmt.Errorf("failed to create perf reader: %w", err)
	}
	defer rd.Close()

	if config.Verbose {
		fmt.Printf("Monitoring %s operations on directory '%s'...\n", 
			strings.Join(config.Operations, ", "), config.Directory)
		if config.ProcessFilter != "" {
			fmt.Printf("Filtering processes containing: %s\n", config.ProcessFilter)
		}
		if config.Duration > 0 {
			fmt.Printf("Running for: %s\n", config.Duration)
		}
	}

	// Set up output writer
	outputWriter := os.Stdout
	if config.OutputFile != "" {
		file, err := os.Create(config.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		outputWriter = file
	}

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		rd.Close()
	}()

	// Process events
	return processEvents(ctx, rd, outputWriter)
}

func attachTracepoints(coll *ebpf.Collection) ([]link.Link, error) {
	links := make([]link.Link, 0)
	
	operationMap := map[string]bool{}
	for _, op := range config.Operations {
		operationMap[op] = true
	}

	if operationMap["open"] {
		l, err := link.Tracepoint("syscalls", "sys_enter_openat", coll.Programs["trace_openat_enter"], nil)
		if err != nil {
			return nil, err
		}
		links = append(links, l)

		l, err = link.Tracepoint("syscalls", "sys_exit_openat", coll.Programs["trace_openat_exit"], nil)
		if err != nil {
			return nil, err
		}
		links = append(links, l)
	}

	if operationMap["read"] {
		l, err := link.Tracepoint("syscalls", "sys_enter_read", coll.Programs["trace_read_enter"], nil)
		if err != nil {
			return nil, err
		}
		links = append(links, l)
	}

	if operationMap["write"] {
		l, err := link.Tracepoint("syscalls", "sys_enter_write", coll.Programs["trace_write_enter"], nil)
		if err != nil {
			return nil, err
		}
		links = append(links, l)
	}

	if operationMap["close"] {
		l, err := link.Tracepoint("syscalls", "sys_enter_close", coll.Programs["trace_close_enter"], nil)
		if err != nil {
			return nil, err
		}
		links = append(links, l)
	}

	return links, nil
}

func processEvents(ctx context.Context, rd *perf.Reader, outputWriter *os.File) error {
	var event Event
	
	// Print table header if using table format
	if config.OutputFormat == "table" {
		printTableHeader(outputWriter)
	}
	
	for {
		record, err := rd.Read()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			if config.Verbose {
				log.Printf("reading from perf event reader: %s", err)
			}
			continue
		}

		if record.LostSamples != 0 {
			if config.Verbose {
				log.Printf("lost %d samples", record.LostSamples)
			}
			continue
		}

		if err := parseEvent(record.RawSample, &event); err != nil {
			if config.Verbose {
				log.Printf("parsing event: %s", err)
			}
			continue
		}

		if config.Debug || shouldProcessEvent(&event) {
			if config.Debug {
				filename := cString(event.Filename[:])
				comm := cString(event.Comm[:])
				fmt.Printf("[DEBUG] PID=%d FD=%d COMM=%s FILE=%s TYPE=%d\n", 
					event.Pid, event.Fd, comm, filename, event.Type)
			}
			if !config.Debug || shouldProcessEvent(&event) {
				outputEvent(&event, outputWriter)
			}
		}
	}
}

func parseEvent(data []byte, event *Event) error {
	expectedSize := int(unsafe.Sizeof(*event))
	if len(data) < expectedSize {
		return fmt.Errorf("data too short: got %d bytes, expected %d bytes", len(data), expectedSize)
	}
	
	*event = *(*Event)(unsafe.Pointer(&data[0]))
	return nil
}

func shouldProcessEvent(event *Event) bool {
	filename := cString(event.Filename[:])
	comm := cString(event.Comm[:])
	
	// Check process filter first (more efficient)
	if config.ProcessFilter != "" && !strings.Contains(comm, config.ProcessFilter) {
		return false
	}
	
	// If we don't have a filename, try to resolve it from /proc
	if filename == "" {
		resolvedFilename := resolveFilenameFromFd(event.Pid, event.Fd)
		if resolvedFilename != "" {
			filename = resolvedFilename
			// Update the event with resolved filename
			for i := 0; i < len(event.Filename) && i < len(resolvedFilename); i++ {
				event.Filename[i] = int8(resolvedFilename[i])
			}
		}
	}
	
	// Skip if we still don't have a filename and it's not a close event
	if filename == "" && event.Type != 3 {
		return false
	}
	
	// Check directory filter - for close events without filename, we can't filter
	if filename != "" {
		// Get current working directory for relative path comparison
		cwd, _ := os.Getwd()
		
		// Convert both paths to absolute for proper comparison
		absFilename := filename
		if !filepath.IsAbs(filename) {
			absFilename = filepath.Join(cwd, filename)
		}
		absFilename = filepath.Clean(absFilename)
		
		absTargetDir := config.Directory
		if !filepath.IsAbs(config.Directory) {
			absTargetDir = filepath.Join(cwd, config.Directory)
		}
		absTargetDir = filepath.Clean(absTargetDir)
		
		// Check if file is within the target directory
		relPath, err := filepath.Rel(absTargetDir, absFilename)
		if err != nil || strings.HasPrefix(relPath, "..") {
			return false
		}
	}
	
	return true
}

func resolveFilenameFromFd(pid uint32, fd int32) string {
	// Try to resolve filename from /proc/PID/fd/FD
	procPath := fmt.Sprintf("/proc/%d/fd/%d", pid, fd)
	
	// Use readlink to get the actual file path
	if link, err := os.Readlink(procPath); err == nil {
		// Clean the path to make it more readable
		if abs, err := filepath.Abs(link); err == nil {
			return abs
		}
		return link
	}
	
	return ""
}

func outputEvent(event *Event, writer *os.File) {
	comm := cString(event.Comm[:])
	filename := cString(event.Filename[:])
	
	// Format filename to be relative and user-friendly
	displayFilename := formatFilename(filename)
	
	eventOutput := EventOutput{
		Timestamp: time.Now().Format(time.RFC3339),
		Pid:       event.Pid,
		Process:   comm,
		Filename:  displayFilename,
	}
	
	switch event.Type {
	case 0:
		eventOutput.Operation = "open"
	case 1:
		eventOutput.Operation = "read"
		if config.ShowFd {
			eventOutput.Fd = event.Fd
		}
	case 2:
		eventOutput.Operation = "write"
		if config.ShowFd {
			eventOutput.Fd = event.Fd
		}
	case 3:
		eventOutput.Operation = "close"
		if config.ShowFd {
			eventOutput.Fd = event.Fd
		}
	default:
		return
	}
	
	switch config.OutputFormat {
	case "json":
		outputJSON(eventOutput, writer)
	case "table":
		outputTable(eventOutput, writer)
	default:
		outputPlain(eventOutput, writer)
	}
}

func outputPlain(event EventOutput, writer *os.File) {
	fdInfo := ""
	if config.ShowFd && event.Fd != 0 {
		fdInfo = fmt.Sprintf(" (fd: %d)", event.Fd)
	}
	
	switch event.Operation {
	case "open":
		fmt.Fprintf(writer, "[%s] Process %s (PID %d) opening file: %s\n", 
			event.Timestamp, event.Process, event.Pid, event.Filename)
	case "read":
		fmt.Fprintf(writer, "[%s] Process %s (PID %d) reading from file: %s%s\n", 
			event.Timestamp, event.Process, event.Pid, event.Filename, fdInfo)
	case "write":
		fmt.Fprintf(writer, "[%s] Process %s (PID %d) writing to file: %s%s\n", 
			event.Timestamp, event.Process, event.Pid, event.Filename, fdInfo)
	case "close":
		fmt.Fprintf(writer, "[%s] Process %s (PID %d) closing file descriptor%s\n", 
			event.Timestamp, event.Process, event.Pid, fdInfo)
	}
}

func outputJSON(event EventOutput, writer *os.File) {
	data, err := json.Marshal(event)
	if err != nil {
		if config.Verbose {
			log.Printf("failed to marshal JSON: %v", err)
		}
		return
	}
	fmt.Fprintf(writer, "%s\n", data)
}

func printTableHeader(writer *os.File) {
	if config.ShowFd {
		fmt.Fprintf(writer, "%-8s %-12s %-8s %-8s %-8s %s\n", 
			"TIME", "PROCESS", "PID", "OPERATION", "FD", "FILENAME")
		fmt.Fprintf(writer, "%-8s %-12s %-8s %-8s %-8s %s\n", 
			"--------", "------------", "--------", "--------", "--------", "--------")
	} else {
		fmt.Fprintf(writer, "%-8s %-12s %-8s %-8s %s\n", 
			"TIME", "PROCESS", "PID", "OPERATION", "FILENAME")
		fmt.Fprintf(writer, "%-8s %-12s %-8s %-8s %s\n", 
			"--------", "------------", "--------", "--------", "--------")
	}
}

func outputTable(event EventOutput, writer *os.File) {
	// Truncate timestamp to show only time part for better readability
	timestamp := event.Timestamp
	if len(timestamp) > 19 {
		timestamp = timestamp[11:19] // Extract HH:MM:SS part
	}
	
	fdCol := ""
	if config.ShowFd && event.Fd != 0 {
		fdCol = fmt.Sprintf("%d", event.Fd)
	}
	
	// Truncate filename if too long for better table formatting
	filename := event.Filename
	if len(filename) > 50 {
		filename = "..." + filename[len(filename)-47:]
	}
	
	if config.ShowFd {
		fmt.Fprintf(writer, "%-8s %-12s %-8d %-8s %-8s %s\n", 
			timestamp, event.Process, event.Pid, event.Operation, fdCol, filename)
	} else {
		fmt.Fprintf(writer, "%-8s %-12s %-8d %-8s %s\n", 
			timestamp, event.Process, event.Pid, event.Operation, filename)
	}
}

func formatFilename(filename string) string {
	if filename == "" {
		return ""
	}
	
	cwd, err := os.Getwd()
	if err != nil {
		return filename
	}
	
	// Try to make path relative to current directory
	if relPath, err := filepath.Rel(cwd, filename); err == nil && !strings.HasPrefix(relPath, "..") {
		return relPath
	}
	
	// If not under current directory, try relative to target directory
	if !filepath.IsAbs(config.Directory) {
		absTargetDir := filepath.Join(cwd, config.Directory)
		if relPath, err := filepath.Rel(absTargetDir, filename); err == nil && !strings.HasPrefix(relPath, "..") {
			return filepath.Join(config.Directory, relPath)
		}
	} else {
		if relPath, err := filepath.Rel(config.Directory, filename); err == nil && !strings.HasPrefix(relPath, "..") {
			return relPath
		}
	}
	
	// If path is too long, elide the beginning (keep the important end part)
	const maxLen = 60
	if len(filename) > maxLen {
		return "..." + filename[len(filename)-maxLen+3:]
	}
	
	return filename
}

func cString(data []int8) string {
	var buf []byte
	for _, b := range data {
		if b == 0 {
			break
		}
		buf = append(buf, byte(b))
	}
	return string(buf)
}