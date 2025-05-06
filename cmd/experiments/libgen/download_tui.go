package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Messages ---

type progressMsg float64 // Represents percentage from 0.0 to 1.0, or -1 for indeterminate
type statusUpdateMsg string
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type successMsg struct{ path string }

type startDownloadMsg struct{}

// --- Styles ---

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
)

// --- Model ---

type downloadModel struct {
	targetUrl  string
	outputPath string
	program    *tea.Program

	spinner  spinner.Model
	prog     progress.Model
	err      error
	done     bool
	quitting bool
	percent  float64 // 0.0 to 1.0, or -1 for indeterminate
	status   string  // Text status for the current operation
	finalMsg string  // Message to display on completion or error
	width    int     // Terminal width
	height   int     // Terminal height (unused for now)
}

func newDownloadModel(targetUrl, outputPath, initialStatus string, program *tea.Program) downloadModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Fixed width for now, updated on WindowSizeMsg
	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	return downloadModel{
		targetUrl:  targetUrl,
		outputPath: outputPath,
		program:    program,
		spinner:    s,
		prog:       prog,
		status:     initialStatus,
		percent:    -1, // Start as indeterminate
	}
}

func (m *downloadModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			return startDownloadMsg{}
		},
	)
}

func (m *downloadModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			m.finalMsg = "Download cancelled."
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height // Store if needed later
		// Update progress bar width, leave some padding
		progressWidth := m.width - 6 // Adjust padding as needed
		if progressWidth < 10 {
			progressWidth = 10 // Minimum width
		}
		m.prog.Width = progressWidth
		return m, nil

	case statusUpdateMsg:
		m.status = string(msg)

	case progressMsg:
		m.percent = float64(msg)
		if m.percent > 1.0 {
			m.percent = 1.0
		}
		// The progress bar will show its animation via FrameMsg.
		// If we directly set percent, we might use m.prog.SetPercent(m.percent) if available
		// or rely on ViewAs(m.percent).
		// For now, FrameMsg handles animation.

	case errMsg:
		m.err = msg.err
		m.finalMsg = fmt.Sprintf("Error: %v", m.err)
		return m, tea.Quit

	case successMsg:
		m.done = true
		m.finalMsg = fmt.Sprintf("Download complete! Saved to %s", msg.path)
		return m, tea.Quit

	case spinner.TickMsg:
		if !m.done && m.err == nil {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case progress.FrameMsg:
		// Ensure progress bar animation continues
		if !m.done && m.err == nil && m.percent >= 0 { // Only update if determinate
			progressModel, frameCmd := m.prog.Update(msg)
			m.prog = progressModel.(progress.Model)
			cmds = append(cmds, frameCmd)
		}

	case startDownloadMsg:
		go actualDownloadLogic(m.targetUrl, m.outputPath, m.program)
		return m, nil

	default:
		// For unknown messages, do nothing
	}

	return m, tea.Batch(cmds...)
}

func (m *downloadModel) View() string {
	if m.quitting && m.finalMsg == "" { // If quitting but finalMsg not set yet by error/success
		return "Cancelling download...\n"
	}
	if m.finalMsg != "" {
		return "\n" + m.finalMsg + "\n\n"
	}

	var content strings.Builder

	content.WriteString(fmt.Sprintf("\n %s %s\n\n", m.spinner.View(), m.status))

	if m.percent < 0 { // Indeterminate
		content.WriteString("  Calculating size or downloading (size unknown)...\n")
	} else { // Determinate
		content.WriteString(fmt.Sprintf("  %s %.0f%%\n", m.prog.ViewAs(m.percent), m.percent*100))
	}

	content.WriteString(fmt.Sprintf("\n%s\n", helpStyle("Press 'q' or Ctrl+C to cancel.")))

	return content.String()
}

// actualDownloadLogic performs the download and sends messages to the tea.Program
func actualDownloadLogic(targetUrl, outputPath string, p *tea.Program) {
	p.Send(statusUpdateMsg("Connecting..."))

	resp, err := http.Get(targetUrl)
	if err != nil {
		p.Send(errMsg{fmt.Errorf("failed to start download: %w", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		p.Send(errMsg{fmt.Errorf("bad status: %s. Response: %s", resp.Status, string(bodyBytes))})
		return
	}

	fileSize, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if fileSize <= 0 {
		p.Send(progressMsg(-1)) // Indeterminate progress
		p.Send(statusUpdateMsg(fmt.Sprintf("Downloading %s (size unknown)...", filepath.Base(outputPath))))
	} else {
		p.Send(statusUpdateMsg(fmt.Sprintf("Downloading %s (%.2f MB)...", filepath.Base(outputPath), float64(fileSize)/1024/1024)))
	}

	// Download to a temporary file first
	tmpOutputPath := outputPath + ".tmp"
	f, err := os.Create(tmpOutputPath)
	if err != nil {
		p.Send(errMsg{fmt.Errorf("failed to create temporary file: %w", err)})
		return
	}
	defer f.Close()

	processedBytes := int64(0)
	buf := make([]byte, 32*1024) // 32KB buffer for copying

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := f.Write(buf[:n])
			if writeErr != nil {
				os.Remove(tmpOutputPath) // Clean up
				p.Send(errMsg{fmt.Errorf("failed to write to file: %w", writeErr)})
				return
			}
			processedBytes += int64(n)
			if fileSize > 0 {
				percent := float64(processedBytes) / float64(fileSize)
				p.Send(progressMsg(percent))
			} else {
				// For indeterminate, spinner is already going.
				// We could send a different kind of message if needed.
				p.Send(progressMsg(-1)) // Keep sending indeterminate signal
			}
		}

		if readErr == io.EOF {
			break // Download finished
		}
		if readErr != nil {
			os.Remove(tmpOutputPath) // Clean up
			p.Send(errMsg{fmt.Errorf("failed to read from download stream: %w", readErr)})
			return
		}
		// Yield to allow other goroutines, like tea.Program's message loop
		time.Sleep(1 * time.Millisecond)
	}

	// Close file before rename
	if err := f.Close(); err != nil {
		os.Remove(tmpOutputPath)
		p.Send(errMsg{fmt.Errorf("failed to close temporary file: %w", err)})
		return
	}

	// Rename temporary file to final output path
	err = os.Rename(tmpOutputPath, outputPath)
	if err != nil {
		os.Remove(tmpOutputPath) // Attempt to clean up again
		p.Send(errMsg{fmt.Errorf("failed to rename temporary file: %w", err)})
		return
	}

	p.Send(successMsg{path: outputPath})
}

// Helper to get filepath.Base for messages
func filepathBase(path string) string {
	return filepath.Base(path)
}
