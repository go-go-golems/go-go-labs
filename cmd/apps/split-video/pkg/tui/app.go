package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/go-go-golems/go-go-labs/cmd/apps/split-video/pkg/config"
	"github.com/go-go-golems/go-go-labs/cmd/apps/split-video/pkg/video"
)

// App represents the main TUI application
type App struct {
	config *config.Config
}

// NewApp creates a new TUI application
func NewApp(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

// Run starts the TUI application
func (a *App) Run() error {
	model := initialModel(a.config)
	p := tea.NewProgram(model, tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}
	
	return nil
}

// Model represents the state of the TUI
type Model struct {
	config       *config.Config
	fileInput    string
	outputDir    string
	splitMode    config.SplitMode
	segments     int
	overlap      time.Duration
	intervals    string
	duration     time.Duration
	extractAudio bool
	audioFormat  string
	
	// UI state
	activeField  int
	processing   bool
	message      string
	messageStyle lipgloss.Style
	inputMode    bool
	inputBuffer  string
	
	// Status
	videoInfo    *video.VideoInfo
}

// Field constants for navigation
const (
	FieldFileInput = iota
	FieldOutputDir
	FieldSplitMode
	FieldSegments
	FieldDuration
	FieldOverlap
	FieldIntervals
	FieldExtractAudio
	FieldAudioFormat
	FieldProcess
	FieldQuit
	FieldCount
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#0000AA")).
		Bold(true).
		Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#0000AA")).
		Bold(true)

	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#AAAAAA")).
		Padding(0, 1)

	activeBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFFF00")).
		Padding(0, 1)

	fieldLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Width(18)

	fieldValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Background(lipgloss.Color("#000080")).
		Padding(0, 1).
		Width(25)

	activeFieldStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#FFFF00")).
		Padding(0, 1).
		Width(25)

	buttonStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#00FF00")).
		Bold(true).
		Padding(0, 2)

	activeButtonStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#FF0000")).
		Bold(true).
		Padding(0, 2)

	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Background(lipgloss.Color("#000080")).
		Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#FF0000")).
		Bold(true).
		Padding(0, 1)

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#00FF00")).
		Bold(true).
		Padding(0, 1)
)

func initialModel(cfg *config.Config) Model {
	model := Model{
		config:       cfg,
		fileInput:    cfg.InputFile, // Use input file from config if provided
		outputDir:    ".",
		splitMode:    config.SplitModeEqual,
		segments:     5,
		overlap:      0,
		duration:     10 * time.Minute,
		extractAudio: false,
		audioFormat:  "mp3",
		intervals:    "10m,20m,30m",
		activeField:  0,
		messageStyle: statusStyle,
	}
	
	// If input file was provided, try to get video info
	if model.fileInput != "" {
		if info, err := video.GetVideoInfo(model.fileInput); err == nil {
			model.videoInfo = info
		}
	}
	
	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.processing {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
			return m, nil
		}

		if m.inputMode {
			return m.handleInputMode(msg)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.activeField > 0 {
				m.activeField--
			}
		case "down", "j":
			if m.activeField < FieldCount-1 {
				m.activeField++
			}
		case "tab":
			m.activeField = (m.activeField + 1) % FieldCount
		case "shift+tab":
			m.activeField = (m.activeField - 1 + FieldCount) % FieldCount
		case "enter", " ":
			return m.handleFieldAction()
		case "left", "h":
			return m.handleFieldDecrement()
		case "right", "l":
			return m.handleFieldIncrement()
		case "f1":
			m.activeField = FieldFileInput
		case "f2":
			m.activeField = FieldSplitMode
		case "f3":
			m.activeField = FieldProcess
		case "f4":
			return m, tea.Quit
		}

	case processMsg:
		m.processing = false
		if msg.err != nil {
			m.message = fmt.Sprintf("Error: %v", msg.err)
			m.messageStyle = errorStyle
		} else {
			m.message = "Processing completed successfully!"
			m.messageStyle = successStyle
		}
		return m, nil
	}
	
	return m, nil
}

func (m Model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.inputMode = false
		switch m.activeField {
		case FieldFileInput:
			m.fileInput = m.inputBuffer
			// Try to get video info
			if m.fileInput != "" {
				if info, err := video.GetVideoInfo(m.fileInput); err == nil {
					m.videoInfo = info
				}
			}
		case FieldOutputDir:
			m.outputDir = m.inputBuffer
		case FieldIntervals:
			m.intervals = m.inputBuffer
		}
		m.inputBuffer = ""
	case "escape":
		m.inputMode = false
		m.inputBuffer = ""
	case "backspace":
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.inputBuffer += msg.String()
		}
	}
	return m, nil
}

func (m Model) handleFieldAction() (tea.Model, tea.Cmd) {
	switch m.activeField {
	case FieldFileInput, FieldOutputDir, FieldIntervals:
		m.inputMode = true
		switch m.activeField {
		case FieldFileInput:
			m.inputBuffer = m.fileInput
		case FieldOutputDir:
			m.inputBuffer = m.outputDir
		case FieldIntervals:
			m.inputBuffer = m.intervals
		}
	case FieldProcess:
		if m.fileInput == "" {
			m.message = "Please select a video file first"
			m.messageStyle = errorStyle
		} else {
			return m, m.startProcessing()
		}
	case FieldQuit:
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleFieldIncrement() (tea.Model, tea.Cmd) {
	switch m.activeField {
	case FieldSplitMode:
		if m.splitMode < config.SplitModeDuration {
			m.splitMode++
		}
	case FieldSegments:
		if m.segments < 20 {
			m.segments++
		}
	case FieldDuration:
		m.duration += time.Minute
	case FieldOverlap:
		m.overlap += 30 * time.Second
	case FieldExtractAudio:
		m.extractAudio = !m.extractAudio
	case FieldAudioFormat:
		formats := []string{"mp3", "wav", "aac", "flac"}
		for i, f := range formats {
			if f == m.audioFormat && i < len(formats)-1 {
				m.audioFormat = formats[i+1]
				break
			}
		}
	}
	return m, nil
}

func (m Model) handleFieldDecrement() (tea.Model, tea.Cmd) {
	switch m.activeField {
	case FieldSplitMode:
		if m.splitMode > 0 {
			m.splitMode--
		}
	case FieldSegments:
		if m.segments > 2 {
			m.segments--
		}
	case FieldDuration:
		if m.duration > time.Minute {
			m.duration -= time.Minute
		}
	case FieldOverlap:
		if m.overlap > 0 {
			m.overlap -= 30 * time.Second
		}
	case FieldExtractAudio:
		m.extractAudio = !m.extractAudio
	case FieldAudioFormat:
		formats := []string{"mp3", "wav", "aac", "flac"}
		for i, f := range formats {
			if f == m.audioFormat && i > 0 {
				m.audioFormat = formats[i-1]
				break
			}
		}
	}
	return m, nil
}



func (m Model) startProcessing() tea.Cmd {
	m.processing = true
	m.message = "Processing..."
	m.messageStyle = statusStyle
	
	return func() tea.Msg {
		// Create config for processing
		cfg := &config.Config{
			InputFile:    m.fileInput,
			OutputDir:    m.outputDir,
			ExtractAudio: m.extractAudio,
			AudioFormat:  m.audioFormat,
		}
		
		var err error
		
		switch m.splitMode {
		case config.SplitModeEqual:
			cfg.Segments = m.segments
			cfg.Overlap = m.overlap
			err = video.SplitEqual(cfg)
		case config.SplitModeTime:
			// Parse intervals string
			intervalStrs := strings.Split(m.intervals, ",")
			cfg.Intervals = make([]string, 0, len(intervalStrs))
			for _, interval := range intervalStrs {
				cfg.Intervals = append(cfg.Intervals, strings.TrimSpace(interval))
			}
			err = video.SplitByTime(cfg)
		case config.SplitModeDuration:
			cfg.SegmentDuration = m.duration
			cfg.Overlap = m.overlap
			err = video.SplitByDuration(cfg)
		}
		
		return processMsg{err: err}
	}
}

type processMsg struct {
	err error
}

func (m Model) View() string {
	var content strings.Builder
	
	// Title bar
	title := titleStyle.Render(" SPLIT-VIDEO v1.0 - Professional Video Splitting Tool ")
	content.WriteString(title + "\n\n")
	
	// Video info section
	if m.videoInfo != nil {
		duration := m.videoInfo.Duration.Truncate(time.Second)
		videoInfoText := fmt.Sprintf("Video: %dx%d @ %d kbps | Duration: %v", 
			m.videoInfo.Width, m.videoInfo.Height, m.videoInfo.Bitrate/1000, duration)
		content.WriteString(statusStyle.Render(videoInfoText) + "\n\n")
	}
	
	// Main form layout
	leftColumn := m.renderLeftColumn()
	rightColumn := m.renderRightColumn()
	
	// Create two-column layout
	leftBox := boxStyle.Render(leftColumn)
	rightBox := boxStyle.Render(rightColumn)
	
	if m.activeField >= FieldFileInput && m.activeField <= FieldAudioFormat {
		if m.activeField <= FieldIntervals {
			leftBox = activeBoxStyle.Render(leftColumn)
		} else {
			rightBox = activeBoxStyle.Render(rightColumn)
		}
	}
	
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, "  ", rightBox)
	content.WriteString(columns + "\n\n")
	
	// Action buttons
	content.WriteString(m.renderButtons() + "\n\n")
	
	// Function keys help
	help := "F1:File F2:Mode F3:Process F4:Quit | ↑↓:Navigate ←→:Change ⏎:Edit/Action"
	content.WriteString(statusStyle.Render(help) + "\n")
	
	// Status/message line
	if m.message != "" {
		content.WriteString("\n" + m.messageStyle.Render(m.message))
	}
	
	if m.processing {
		content.WriteString("\n" + statusStyle.Render("█ Processing video... Please wait █"))
	}
	
	return content.String()
}

func (m Model) renderLeftColumn() string {
	var content strings.Builder
	content.WriteString(headerStyle.Render(" INPUT & OUTPUT ") + "\n\n")
	
	// File input
	fileValue := m.fileInput
	if m.inputMode && m.activeField == FieldFileInput {
		fileValue = m.inputBuffer + "█"
	}
	if fileValue == "" {
		fileValue = "<none selected>"
	}
	
	style := fieldValueStyle
	if m.activeField == FieldFileInput {
		style = activeFieldStyle
	}
	content.WriteString(fieldLabelStyle.Render("Input File:") + " " + style.Render(fileValue) + "\n")
	
	// Output directory
	outputValue := m.outputDir
	if m.inputMode && m.activeField == FieldOutputDir {
		outputValue = m.inputBuffer + "█"
	}
	
	style = fieldValueStyle
	if m.activeField == FieldOutputDir {
		style = activeFieldStyle
	}
	content.WriteString(fieldLabelStyle.Render("Output Dir:") + " " + style.Render(outputValue) + "\n\n")
	
	// Split mode
	modes := []string{"Equal Segments", "Time Intervals", "Duration Based"}
	modeValue := modes[m.splitMode]
	
	style = fieldValueStyle
	if m.activeField == FieldSplitMode {
		style = activeFieldStyle
	}
	content.WriteString(fieldLabelStyle.Render("Split Mode:") + " " + style.Render(modeValue) + "\n")
	
	// Mode-specific fields
	switch m.splitMode {
	case config.SplitModeEqual:
		segmentValue := fmt.Sprintf("%d", m.segments)
		style = fieldValueStyle
		if m.activeField == FieldSegments {
			style = activeFieldStyle
		}
		content.WriteString(fieldLabelStyle.Render("Segments:") + " " + style.Render(segmentValue) + "\n")
		
	case config.SplitModeTime:
		intervalValue := m.intervals
		if m.inputMode && m.activeField == FieldIntervals {
			intervalValue = m.inputBuffer + "█"
		}
		style = fieldValueStyle
		if m.activeField == FieldIntervals {
			style = activeFieldStyle
		}
		content.WriteString(fieldLabelStyle.Render("Intervals:") + " " + style.Render(intervalValue) + "\n")
		
	case config.SplitModeDuration:
		durationValue := m.duration.String()
		style = fieldValueStyle
		if m.activeField == FieldDuration {
			style = activeFieldStyle
		}
		content.WriteString(fieldLabelStyle.Render("Duration:") + " " + style.Render(durationValue) + "\n")
	}
	
	// Overlap (for equal and duration modes)
	if m.splitMode != config.SplitModeTime {
		overlapValue := m.overlap.String()
		style = fieldValueStyle
		if m.activeField == FieldOverlap {
			style = activeFieldStyle
		}
		content.WriteString(fieldLabelStyle.Render("Overlap:") + " " + style.Render(overlapValue) + "\n")
	}
	
	return content.String()
}

func (m Model) renderRightColumn() string {
	var content strings.Builder
	content.WriteString(headerStyle.Render(" AUDIO OPTIONS ") + "\n\n")
	
	// Extract audio
	audioValue := "No"
	if m.extractAudio {
		audioValue = "Yes"
	}
	
	style := fieldValueStyle
	if m.activeField == FieldExtractAudio {
		style = activeFieldStyle
	}
	content.WriteString(fieldLabelStyle.Render("Extract Audio:") + " " + style.Render(audioValue) + "\n")
	
	// Audio format (only if extracting)
	if m.extractAudio {
		formatValue := strings.ToUpper(m.audioFormat)
		style = fieldValueStyle
		if m.activeField == FieldAudioFormat {
			style = activeFieldStyle
		}
		content.WriteString(fieldLabelStyle.Render("Audio Format:") + " " + style.Render(formatValue) + "\n")
	}
	
	content.WriteString("\n")
	content.WriteString(headerStyle.Render(" PREVIEW ") + "\n\n")
	
	// Show preview of what will be created
	if m.fileInput != "" {
		switch m.splitMode {
		case config.SplitModeEqual:
			content.WriteString(fmt.Sprintf("Will create %d segments\n", m.segments))
			if m.overlap > 0 {
				content.WriteString(fmt.Sprintf("with %v overlap\n", m.overlap))
			}
		case config.SplitModeTime:
			intervals := strings.Split(m.intervals, ",")
			content.WriteString(fmt.Sprintf("Will create %d parts\n", len(intervals)+1))
			content.WriteString(fmt.Sprintf("at intervals: %s\n", m.intervals))
		case config.SplitModeDuration:
			if m.videoInfo != nil {
				segments := int(m.videoInfo.Duration/m.duration) + 1
				content.WriteString(fmt.Sprintf("Will create ~%d chunks\n", segments))
				content.WriteString(fmt.Sprintf("of %v each\n", m.duration))
			}
		}
		
		if m.extractAudio {
			content.WriteString(fmt.Sprintf("+ Audio files (%s)\n", m.audioFormat))
		}
	} else {
		content.WriteString("Select input file to preview\n")
	}
	
	return content.String()
}

func (m Model) renderButtons() string {
	var buttons []string
	
	// Process button
	processStyle := buttonStyle
	if m.activeField == FieldProcess {
		processStyle = activeButtonStyle
	}
	buttons = append(buttons, processStyle.Render("▶ PROCESS"))
	
	// Quit button  
	quitStyle := buttonStyle
	if m.activeField == FieldQuit {
		quitStyle = activeButtonStyle
	}
	buttons = append(buttons, quitStyle.Render("✖ QUIT"))
	
	return lipgloss.JoinHorizontal(lipgloss.Center, buttons...)
}


