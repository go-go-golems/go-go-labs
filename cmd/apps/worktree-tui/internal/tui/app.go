package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/go-go-golems/go-go-labs/cmd/apps/worktree-tui/internal/config"
	"github.com/go-go-golems/go-go-labs/cmd/apps/worktree-tui/internal/tui/screens"
)

// Screen represents different TUI screens
type Screen int

const (
	SelectionScreen Screen = iota
	ConfigScreen
	ProgressScreen
	CompletionScreen
)

// App represents the main TUI application
type App struct {
	config  *config.Config
	screen  Screen
	width   int
	height  int
	
	// Screen models
	selection  *screens.SelectionModel
	config_    *screens.ConfigModel
	progress   *screens.ProgressModel
	completion *screens.CompletionModel
	
	// Shared state
	selectedRepos []config.RepositorySelection
	workspaceReq  *config.WorkspaceRequest
	
	// Key bindings
	keys keyMap
}

type keyMap struct {
	Quit   key.Binding
	Back   key.Binding
	Help   key.Binding
}

var defaultKeys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

// NewApp creates a new TUI application
func NewApp(cfg *config.Config) *App {
	app := &App{
		config: cfg,
		screen: SelectionScreen,
		keys:   defaultKeys,
	}
	
	// Initialize screens
	app.selection = screens.NewSelectionModel(cfg)
	
	return app
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	return a.selection.Init()
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		
		// Update all screen models with new size
		if a.selection != nil {
			a.selection.SetSize(msg.Width, msg.Height)
		}
		if a.config_ != nil {
			a.config_.SetSize(msg.Width, msg.Height)
		}
		if a.progress != nil {
			a.progress.SetSize(msg.Width, msg.Height)
		}
		if a.completion != nil {
			a.completion.SetSize(msg.Width, msg.Height)
		}
		
		return a, nil
		
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit
		case key.Matches(msg, a.keys.Back):
			return a.handleBack()
		}
		
	case screens.NavigateToConfigMsg:
		return a.navigateToConfig(msg)
		
	case screens.NavigateToProgressMsg:
		return a.navigateToProgress(msg)
		
	case screens.NavigateToCompletionMsg:
		return a.navigateToCompletion(msg)
		
	case screens.QuitMsg:
		return a, tea.Quit
	}
	
	// Route to current screen
	return a.updateCurrentScreen(msg)
}

// View implements tea.Model
func (a *App) View() string {
	switch a.screen {
	case SelectionScreen:
		if a.selection != nil {
			return a.selection.View()
		}
	case ConfigScreen:
		if a.config_ != nil {
			return a.config_.View()
		}
	case ProgressScreen:
		if a.progress != nil {
			return a.progress.View()
		}
	case CompletionScreen:
		if a.completion != nil {
			return a.completion.View()
		}
	}
	
	return "Loading..."
}

func (a *App) updateCurrentScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch a.screen {
	case SelectionScreen:
		if a.selection != nil {
			var model tea.Model
			model, cmd = a.selection.Update(msg)
			a.selection = model.(*screens.SelectionModel)
		}
	case ConfigScreen:
		if a.config_ != nil {
			var model tea.Model
			model, cmd = a.config_.Update(msg)
			a.config_ = model.(*screens.ConfigModel)
		}
	case ProgressScreen:
		if a.progress != nil {
			var model tea.Model
			model, cmd = a.progress.Update(msg)
			a.progress = model.(*screens.ProgressModel)
		}
	case CompletionScreen:
		if a.completion != nil {
			var model tea.Model
			model, cmd = a.completion.Update(msg)
			a.completion = model.(*screens.CompletionModel)
		}
	}
	
	return a, cmd
}

func (a *App) handleBack() (tea.Model, tea.Cmd) {
	switch a.screen {
	case ConfigScreen:
		a.screen = SelectionScreen
		return a, nil
	case ProgressScreen:
		// Can't go back from progress screen
		return a, nil
	case CompletionScreen:
		a.screen = SelectionScreen
		// Reset state for new workspace creation
		a.selectedRepos = nil
		a.workspaceReq = nil
		a.selection.Reset()
		return a, nil
	default:
		return a, tea.Quit
	}
}

func (a *App) navigateToConfig(msg screens.NavigateToConfigMsg) (tea.Model, tea.Cmd) {
	a.selectedRepos = msg.SelectedRepos
	a.screen = ConfigScreen
	
	if a.config_ == nil {
		a.config_ = screens.NewConfigModel(a.config, a.selectedRepos)
		a.config_.SetSize(a.width, a.height)
	} else {
		a.config_.SetSelectedRepos(a.selectedRepos)
	}
	
	return a, a.config_.Init()
}

func (a *App) navigateToProgress(msg screens.NavigateToProgressMsg) (tea.Model, tea.Cmd) {
	a.workspaceReq = msg.WorkspaceRequest
	a.screen = ProgressScreen
	
	if a.progress == nil {
		a.progress = screens.NewProgressModel(a.workspaceReq)
		a.progress.SetSize(a.width, a.height)
	} else {
		a.progress.SetWorkspaceRequest(a.workspaceReq)
	}
	
	return a, a.progress.Init()
}

func (a *App) navigateToCompletion(msg screens.NavigateToCompletionMsg) (tea.Model, tea.Cmd) {
	a.screen = CompletionScreen
	
	if a.completion == nil {
		a.completion = screens.NewCompletionModel(a.workspaceReq, msg.Success, msg.Error)
		a.completion.SetSize(a.width, a.height)
	} else {
		a.completion.SetResult(a.workspaceReq, msg.Success, msg.Error)
	}
	
	return a, a.completion.Init()
}