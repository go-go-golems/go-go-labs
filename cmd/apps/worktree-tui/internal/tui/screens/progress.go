package screens

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/go-go-golems/go-go-labs/cmd/apps/worktree-tui/internal/config"
	"github.com/go-go-golems/go-go-labs/cmd/apps/worktree-tui/internal/workspace"
)

var _ tea.Model = (*ProgressModel)(nil)

type ProgressModel struct {
	workspaceReq *config.WorkspaceRequest
	width        int
	height       int
	
	// UI components
	progressBar progress.Model
	
	// State
	currentStep int
	totalSteps  int
	currentTask string
	logs        []logEntry
	completed   bool
	success     bool
	err         error
	
	// Key bindings
	keys progressKeyMap
}

type progressKeyMap struct {
	Cancel key.Binding
}

var progressKeys = progressKeyMap{
	Cancel: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "cancel"),
	),
}

type logEntry struct {
	timestamp time.Time
	message   string
}

type progressTickMsg struct {
	step        int
	total       int
	currentTask string
	logMessage  string
}

func NewProgressModel(req *config.WorkspaceRequest) *ProgressModel {
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 50

	return &ProgressModel{
		workspaceReq: req,
		progressBar:  prog,
		totalSteps:   len(req.Repositories) + 2, // +2 for directory creation and go.work init
		keys:         progressKeys,
	}
}

func (m *ProgressModel) Init() tea.Cmd {
	return tea.Batch(
		m.startWorkspaceCreation(),
		m.progressBar.Init(),
	)
}

func (m *ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Cancel) && !m.completed {
			// TODO: Implement cancellation
			return m, func() tea.Msg {
				return NavigateToCompletionMsg{
					Success: false,
					Error:   fmt.Errorf("operation cancelled by user"),
				}
			}
		}

	case progressTickMsg:
		m.currentStep = msg.step
		m.totalSteps = msg.total
		m.currentTask = msg.currentTask
		
		if msg.logMessage != "" {
			m.logs = append(m.logs, logEntry{
				timestamp: time.Now(),
				message:   msg.logMessage,
			})
			
			// Keep only the last 10 log entries
			if len(m.logs) > 10 {
				m.logs = m.logs[len(m.logs)-10:]
			}
		}
		
		progress := float64(m.currentStep) / float64(m.totalSteps)
		return m, m.progressBar.SetPercent(progress)

	case ProgressCompleteMsg:
		m.completed = true
		m.success = msg.Success
		m.err = msg.Error
		
		return m, func() tea.Msg {
			return NavigateToCompletionMsg{
				Success: msg.Success,
				Error:   msg.Error,
			}
		}

	default:
		var cmd tea.Cmd
		var model tea.Model
		model, cmd = m.progressBar.Update(msg)
		if pb, ok := model.(progress.Model); ok {
			m.progressBar = pb
		}
		return m, cmd
	}

	return m, nil
}

func (m *ProgressModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true).
		Padding(1, 2)

	title := titleStyle.Render(fmt.Sprintf("Creating Workspace: %s", m.workspaceReq.Name))

	// Progress section
	progressSection := m.renderProgress()

	// Current task section
	taskSection := m.renderCurrentTask()

	// Logs section
	logsSection := m.renderLogs()

	// Status section
	statusSection := m.renderStatus()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		progressSection,
		taskSection,
		logsSection,
		statusSection,
	)
}

func (m *ProgressModel) renderProgress() string {
	sectionStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		MarginBottom(1)

	steps := make([]string, m.totalSteps)
	
	// Directory creation
	if m.currentStep > 0 {
		steps[0] = "✓ Creating workspace directory"
	} else if m.currentStep == 0 {
		steps[0] = "⟳ Creating workspace directory"
	} else {
		steps[0] = "○ Creating workspace directory"
	}

	// Repository worktrees
	for i, repo := range m.workspaceReq.Repositories {
		stepIdx := i + 1
		if m.currentStep > stepIdx {
			steps[stepIdx] = fmt.Sprintf("✓ Setting up %s worktree", repo.Name)
		} else if m.currentStep == stepIdx {
			steps[stepIdx] = fmt.Sprintf("⟳ Setting up %s worktree", repo.Name)
		} else {
			steps[stepIdx] = fmt.Sprintf("○ Setting up %s worktree", repo.Name)
		}
	}

	// Go workspace initialization
	lastIdx := len(steps) - 1
	if m.currentStep > lastIdx {
		steps[lastIdx] = "✓ Initializing go.work"
	} else if m.currentStep == lastIdx {
		steps[lastIdx] = "⟳ Initializing go.work"
	} else {
		steps[lastIdx] = "○ Initializing go.work"
	}

	progressContent := strings.Join(steps, "\n")
	progressContent += "\n\n" + m.progressBar.View()

	return sectionStyle.Render(progressContent)
}

func (m *ProgressModel) renderCurrentTask() string {
	if m.currentTask == "" {
		return ""
	}

	sectionStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		MarginBottom(1)

	taskStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true)

	content := taskStyle.Render("Current: ") + m.currentTask

	return sectionStyle.Render(content)
}

func (m *ProgressModel) renderLogs() string {
	if len(m.logs) == 0 {
		return ""
	}

	sectionStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		MarginBottom(1).
		Height(8) // Fixed height for log section

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62"))

	header := headerStyle.Render("Logs:")

	logLines := make([]string, len(m.logs))
	for i, entry := range m.logs {
		timestamp := entry.timestamp.Format("15:04:05")
		logLines[i] = fmt.Sprintf("%s │ %s", 
			lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(timestamp),
			entry.message)
	}

	content := header + "\n\n" + strings.Join(logLines, "\n")

	return sectionStyle.Render(content)
}

func (m *ProgressModel) renderStatus() string {
	statusStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color("240"))

	if m.completed {
		if m.success {
			status := lipgloss.NewStyle().
				Foreground(lipgloss.Color("34")).
				Render("✓ Workspace created successfully!")
			return statusStyle.Render(status + "\n\nPress any key to continue...")
		} else {
			status := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Render("✗ Workspace creation failed")
			if m.err != nil {
				status += "\n" + m.err.Error()
			}
			return statusStyle.Render(status + "\n\nPress any key to continue...")
		}
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	help := helpStyle.Render("ctrl+c cancel")
	progress := fmt.Sprintf("Progress: %d/%d steps", m.currentStep, m.totalSteps)

	return statusStyle.Render(progress + "\n" + help)
}

func (m *ProgressModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.progressBar.Width = width - 10
}

func (m *ProgressModel) SetWorkspaceRequest(req *config.WorkspaceRequest) {
	m.workspaceReq = req
	m.totalSteps = len(req.Repositories) + 2
	m.currentStep = 0
	m.currentTask = ""
	m.logs = nil
	m.completed = false
	m.success = false
	m.err = nil
}

func (m *ProgressModel) startWorkspaceCreation() tea.Cmd {
	return func() tea.Msg {
		manager := workspace.NewManager()
		
		// Create a context that can be cancelled
		ctx := context.Background()
		
		// Create progress channel
		progressCh := make(chan progressTickMsg, 10)
		
		// Start workspace creation in a goroutine
		go func() {
			defer close(progressCh)
			
			err := manager.CreateWorkspace(ctx, m.workspaceReq, func(step, total int, task, logMsg string) {
			select {
			case progressCh <- progressTickMsg{
			step:        step,
			total:       total,
			currentTask: task,
			logMessage:  logMsg,
			}:
			case <-ctx.Done():
			return
			}
			})
			
			// Send completion message with result
			select {
			case progressCh <- progressTickMsg{
			step:        m.totalSteps,
			total:       m.totalSteps,
			currentTask: "",
			logMessage:  fmt.Sprintf("Workspace creation completed (success: %t)", err == nil),
			}:
			case <-ctx.Done():
			return
			}
			
			// TODO: Send final completion result through proper channel
			// For now, this is a simplified implementation
			_ = err // Acknowledge the error variable is captured
		}()
		
		// Return the first progress update
		select {
		case msg := <-progressCh:
			return msg
		case <-time.After(5 * time.Second):
			return ProgressCompleteMsg{
				Success: false,
				Error:   fmt.Errorf("workspace creation timed out"),
			}
		}
	}
}