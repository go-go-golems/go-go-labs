package screens

import "github.com/go-go-golems/go-go-labs/cmd/apps/worktree-tui/internal/config"

// Navigation messages between screens

type NavigateToConfigMsg struct {
	SelectedRepos []config.RepositorySelection
}

type NavigateToProgressMsg struct {
	WorkspaceRequest *config.WorkspaceRequest
}

type NavigateToCompletionMsg struct {
	Success bool
	Error   error
}

type QuitMsg struct{}

// Progress update messages

type ProgressUpdateMsg struct {
	Step        int
	Total       int
	CurrentTask string
	LogEntry    string
}

type ProgressCompleteMsg struct {
	Success bool
	Error   error
}