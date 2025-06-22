package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	logLevel    string
	dslFilePath string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "prompt-renderer",
		Short: "Interactive prompt template builder",
		Long: `A terminal-based application for converting YAML prompt templates 
into clipboard-ready prompts through an interactive configuration interface.`,
		RunE: runApp,
	}
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().StringVar(&dslFilePath, "dsl", "", "Path to DSL file (defaults to auto-discovery)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runApp(cmd *cobra.Command, args []string) error {
	// Setup logging
	setupLogging()

	log.Info().Msg("Starting prompt renderer application")

	// Load DSL file
	var dslFile *DSLFile
	var err error

	if dslFilePath != "" {
		log.Info().Str("path", dslFilePath).Msg("Loading DSL file from specified path")
		dslFile, err = ParseDSLFile(dslFilePath)
	} else {
		log.Info().Msg("Auto-discovering DSL file")
		dslFile, err = LoadDefaultDSLFile()
	}

	if err != nil {
		return errors.Wrap(err, "❌ Failed to load DSL file. Please check your YAML syntax and ensure the file exists.")
	}

	log.Info().Int("templates", len(dslFile.Templates)).Msg("DSL file loaded successfully")

	// Initialize components
	renderer := NewPromptRenderer(dslFile)
	clipboard := NewClipboardManager()
	persistence, err := NewPersistenceManager()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize persistence manager")
		persistence = nil
	}

	// Create main application model
	app := NewAppModel(dslFile, renderer, clipboard, persistence)

	// Start the Bubble Tea program
	program := tea.NewProgram(app, tea.WithAltScreen())

	log.Info().Msg("Starting TUI application")

	if _, err := program.Run(); err != nil {
		return errors.Wrap(err, "failed to run TUI application")
	}

	log.Info().Msg("Application exited successfully")
	return nil
}

func setupLogging() {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// AppModel represents the main application model
type AppModel struct {
	state       AppState
	dslFile     *DSLFile
	renderer    *PromptRenderer
	clipboard   *ClipboardManager
	persistence *PersistenceManager

	listModel   *TemplateListModel
	configModel *TemplateConfigModel

	width    int
	height   int
	quitting bool
}

// NewAppModel creates a new application model
func NewAppModel(dslFile *DSLFile, renderer *PromptRenderer, clipboard *ClipboardManager, persistence *PersistenceManager) *AppModel {
	return &AppModel{
		state:       StateTemplateList,
		dslFile:     dslFile,
		renderer:    renderer,
		clipboard:   clipboard,
		persistence: persistence,
		listModel:   NewTemplateListModel(dslFile.Templates),
	}
}

// Init implements tea.Model
func (m *AppModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if m.listModel != nil {
			m.listModel.width = msg.Width
			m.listModel.height = msg.Height
		}
		if m.configModel != nil {
			m.configModel.SetSize(msg.Width, msg.Height)
		}

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

	case SelectTemplateMsg:
		log.Info().Str("template", msg.Template.ID).Msg("Template selected")
		m.state = StateTemplateConfig
		m.configModel = NewTemplateConfigModel(&msg.Template, m.renderer)
		m.configModel.SetSize(m.width, m.height)

		// Try to load previous state for this template
		if m.persistence != nil {
			if savedState, err := m.persistence.LoadCurrentState(); err == nil && savedState != nil && savedState.TemplateID == msg.Template.ID {
				log.Info().Msg("Loaded previous state for template")
				// We need to update the state manager directly since selection is internal
				// For now, let's create a new model with the saved state
				m.configModel = NewTemplateConfigModel(&msg.Template, m.renderer)
				m.configModel.SetSize(m.width, m.height)
				// TODO: Add method to restore saved state
			}
		}

	case GoBackMsg:
		log.Info().Msg("Going back to template list")
		m.state = StateTemplateList
		m.configModel = nil

	case CopyPromptMsg:
		log.Info().Msg("Copying prompt to clipboard")
		if err := m.clipboard.CopyToClipboard(msg.Prompt); err != nil {
			log.Error().Err(err).Msg("Failed to copy to clipboard")
		}
		return m, func() tea.Msg {
			return CopyDoneMsg{}
		}

	case SaveSelectionMsg:
		log.Info().Msg("Saving selection state")
		if m.persistence != nil {
			if err := m.persistence.SaveCurrentState(msg.Selection); err != nil {
				log.Error().Err(err).Msg("Failed to save current state")
			}
			if err := m.persistence.SaveToHistory(msg.Selection); err != nil {
				log.Error().Err(err).Msg("Failed to save to history")
			}
		}

	case ShowHelpMsg:
		// TODO: Implement help screen
		log.Info().Msg("Help requested")
	}

	// Delegate to appropriate model
	switch m.state {
	case StateTemplateList:
		if m.listModel != nil {
			var cmd tea.Cmd
			model, cmd := m.listModel.Update(msg)
			m.listModel = model.(*TemplateListModel)
			return m, cmd
		}
	case StateTemplateConfig:
		if m.configModel != nil {
			var cmd tea.Cmd
			model, cmd := m.configModel.Update(msg)
			m.configModel = model.(*TemplateConfigModel)
			return m, cmd
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *AppModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	switch m.state {
	case StateTemplateList:
		if m.listModel != nil {
			return m.listModel.View()
		}
	case StateTemplateConfig:
		if m.configModel != nil {
			return m.configModel.View()
		}
	}

	return "Loading..."
}

// Auto-save functionality
func (m *AppModel) startAutoSave() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return AutoSaveMsg{}
	})
}

// AutoSaveMsg triggers an auto-save
type AutoSaveMsg struct{}
