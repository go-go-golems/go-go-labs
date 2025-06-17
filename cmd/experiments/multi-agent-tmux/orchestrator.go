package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/GianlucaP106/gotmux/gotmux"
	"github.com/rs/zerolog/log"
)

// Task represents a task to be executed by an agent
type Task struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	AgentType   string `json:"agent_type"` // "research", "analysis", "writing", "review"
}

// Orchestrator manages multiple agents and their tmux output
type Orchestrator struct {
	tmux         *gotmux.Tmux
	session      *gotmux.Session
	agents       map[string]Agent
	agentPanes   map[string]*gotmux.Pane
	agentSockets map[string]string
	statusPane   *gotmux.Pane
	statusSocket string
	sessionName  string
	socketServer *SocketServer
	mu           sync.RWMutex
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(sessionName string) (*Orchestrator, error) {
	tmux, err := gotmux.DefaultTmux()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tmux: %w", err)
	}

	// Create socket server
	socketServer, err := NewSocketServer()
	if err != nil {
		return nil, fmt.Errorf("failed to create socket server: %w", err)
	}

	return &Orchestrator{
		tmux:         tmux,
		agents:       make(map[string]Agent),
		agentPanes:   make(map[string]*gotmux.Pane),
		agentSockets: make(map[string]string),
		sessionName:  sessionName,
		socketServer: socketServer,
	}, nil
}

// Initialize sets up the tmux session and panes for the orchestrator
func (o *Orchestrator) Initialize(ctx context.Context) error {
	log.Info().Str("session", o.sessionName).Str("socketDir", o.socketServer.GetSocketDir()).Msg("Initializing orchestrator session")

	// Create the main session
	session, err := o.tmux.NewSession(&gotmux.SessionOptions{
		Name: o.sessionName,
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	o.session = session

	// Get the main window
	mainWindow, err := session.NewWindow(nil)
	if err != nil {
		return fmt.Errorf("failed to create main window: %w", err)
	}

	// Initialize agents first
	o.registerAgents()

	// Setup panes and sockets for agents and status
	err = o.setupPanesAndSockets(mainWindow)
	if err != nil {
		return fmt.Errorf("failed to setup panes and sockets: %w", err)
	}

	// Give TUI processes a moment to start up and connect
	time.Sleep(2 * time.Second)

	o.sendStatusMessage("🚀 Multi-Agent Orchestrator Initialized")
	o.sendStatusMessage(fmt.Sprintf("📊 Session: %s", o.sessionName))
	o.sendStatusMessage(fmt.Sprintf("🤖 Agents: %d registered", len(o.agents)))
	o.sendStatusMessage(fmt.Sprintf("📂 Sockets: %s", o.socketServer.GetSocketDir()))

	return nil
}

// registerAgents creates and registers all available agents
func (o *Orchestrator) registerAgents() {
	agents := []Agent{
		NewResearchAgent("research-001"),
		NewAnalysisAgent("analysis-001"),
		NewWritingAgent("writing-001"),
		NewReviewAgent("review-001"),
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	for _, agent := range agents {
		o.agents[agent.ID()] = agent
		log.Debug().Str("agent_id", agent.ID()).Str("agent_name", agent.Name()).Msg("Registered agent")
	}
}

// setupPanesAndSockets creates tmux panes and sockets for each agent
func (o *Orchestrator) setupPanesAndSockets(mainWindow *gotmux.Window) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Get the initial pane (will be used for status)
	statusPane, err := mainWindow.GetPaneByIndex(0)
	if err != nil {
		return fmt.Errorf("failed to get status pane: %w", err)
	}
	o.statusPane = statusPane

	// Create status socket
	statusSocket, err := o.socketServer.CreateStatusSocket()
	if err != nil {
		return fmt.Errorf("failed to create status socket: %w", err)
	}
	o.statusSocket = statusSocket

	// Setup status pane to run TUI
	tuiCmd := fmt.Sprintf("cd %s && go run . tui --socket %s", ".", statusSocket)
	err = statusPane.SendKeys(tuiCmd)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to setup status pane TUI")
	} else {
		// Send Enter to execute the command
		err = statusPane.SendKeys("\n")
		if err != nil {
			log.Warn().Err(err).Msg("Failed to send Enter to status pane")
		}
	}

	// Create panes for each agent by splitting
	agentIDs := make([]string, 0, len(o.agents))
	for agentID := range o.agents {
		agentIDs = append(agentIDs, agentID)
	}

	// -----------------------------------------------------------------------------
	// Layout strategy:
	// 1. Split the initial status pane horizontally, creating a *right* column that
	//    will host all agent panes stacked vertically.
	// 2. The status pane (index 0) therefore spans the full height on the *left*.
	// 3. All agent panes live in the *right* column, one below the other.
	// -----------------------------------------------------------------------------

	// 1. Create the right column for agents by horizontally splitting the status
	//    pane.
	if err := statusPane.SplitWindow(&gotmux.SplitWindowOptions{
		SplitDirection: gotmux.PaneSplitDirectionHorizontal, // "-h" – vertical bar
	}); err != nil {
		return fmt.Errorf("failed to create right column for agents: %w", err)
	}

	// The newly-created pane (index 1) is the top pane in the right column.
	rightPane, err := mainWindow.GetPaneByIndex(1)
	if err != nil {
		return fmt.Errorf("failed to get right pane after horizontal split: %w", err)
	}

	currentPane := rightPane
	currentPaneIndex := 1 // keep track of the tmux pane index we are filling

	for i, agentID := range agentIDs {
		agent := o.agents[agentID]

		// Create socket for this agent
		socketPath, err := o.socketServer.CreateAgentSocket(agentID)
		if err != nil {
			return fmt.Errorf("failed to create socket for agent %s: %w", agentID, err)
		}
		o.agentSockets[agentID] = socketPath

		// Initialize agent display
		if err := o.socketServer.InitializeAgent(agentID, agent.Name(), agent.Role()); err != nil {
			log.Warn().Err(err).Str("agent", agentID).Msg("Failed to initialize agent display")
		}

		// Send initial ready message
		readyMsg := NewAgentUpdateMessage(agentID, agent.Name(), agent.Role(), "Ready for tasks", "status")
		o.socketServer.SendToAgent(agentID, readyMsg)

		// Store the pane reference **before** potentially splitting it for the next
		// agent.
		o.agentPanes[agentID] = currentPane

		// Setup the pane to run TUI for this agent
		tuiCmd := fmt.Sprintf("cd %s && go run . tui --socket %s", ".", socketPath)
		if err := currentPane.SendKeys(tuiCmd); err != nil {
			log.Warn().Err(err).Str("agent", agentID).Msg("Failed to setup agent pane TUI")
		} else {
			// Send Enter to execute the command
			if err := currentPane.SendKeys("\n"); err != nil {
				log.Warn().Err(err).Str("agent", agentID).Msg("Failed to send Enter to agent pane")
			}
		}

		log.Info().
			Str("agent_id", agentID).
			Str("agent_name", agent.Name()).
			Str("socket_path", socketPath).
			Msg("Setup agent pane and socket")

		// If there is another agent to place, split the current agent pane *vertically*
		// to create a new pane *below* it.
		if i < len(agentIDs)-1 {
			if err := currentPane.SplitWindow(&gotmux.SplitWindowOptions{
				SplitDirection: gotmux.PaneSplitDirectionVertical, // "-v" – horizontal bar
			}); err != nil {
				return fmt.Errorf("failed to split pane for next agent after %s: %w", agentID, err)
			}

			currentPaneIndex++
			newPane, err := mainWindow.GetPaneByIndex(currentPaneIndex)
			if err != nil {
				return fmt.Errorf("failed to get new pane for agent after %s: %w", agentID, err)
			}
			currentPane = newPane
		}
	}

	return nil
}

// ExecuteTasks executes multiple tasks concurrently across agents
func (o *Orchestrator) ExecuteTasks(ctx context.Context, tasks []Task) error {
	if len(tasks) == 0 {
		return fmt.Errorf("no tasks provided")
	}

	o.sendStatusMessage(fmt.Sprintf("🎯 Starting execution of %d tasks", len(tasks)))

	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))

	for _, task := range tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()

			agent, exists := o.agents[o.findAgentByType(t.AgentType)]
			if !exists {
				errChan <- fmt.Errorf("no agent found for type: %s", t.AgentType)
				return
			}

			o.sendStatusMessage(fmt.Sprintf("▶️  Starting task %s on %s", t.ID, agent.Name()))

			// Create output channel for this agent
			output := make(chan AgentMessage, 100)

			// Start goroutine to handle agent output
			go o.handleAgentOutput(agent.ID(), output)

			// Execute the task
			if err := agent.Execute(ctx, t.Description, output); err != nil {
				errChan <- fmt.Errorf("agent %s failed: %w", agent.ID(), err)
				return
			}

			o.sendStatusMessage(fmt.Sprintf("✅ Task %s completed by %s", t.ID, agent.Name()))
		}(task)
	}

	// Wait for all tasks to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		o.sendStatusMessage(fmt.Sprintf("❌ %d tasks failed", len(errors)))
		for _, err := range errors {
			o.sendStatusMessage(fmt.Sprintf("Error: %v", err))
		}
		return fmt.Errorf("some tasks failed")
	}

	o.sendStatusMessage("🎉 All tasks completed successfully!")
	return nil
}

// findAgentByType finds an agent ID by its type
func (o *Orchestrator) findAgentByType(agentType string) string {
	o.mu.RLock()
	defer o.mu.RUnlock()

	typeMap := map[string]string{
		"research": "research-001",
		"analysis": "analysis-001",
		"writing":  "writing-001",
		"review":   "review-001",
	}

	if agentID, exists := typeMap[agentType]; exists {
		return agentID
	}

	// Fallback: return first agent if type not found
	for id := range o.agents {
		return id
	}

	return ""
}

// handleAgentOutput processes messages from an agent and displays them in tmux
func (o *Orchestrator) handleAgentOutput(agentID string, output <-chan AgentMessage) {
	agent, exists := o.agents[agentID]
	if !exists {
		log.Error().Str("agent_id", agentID).Msg("No agent found")
		return
	}

	for msg := range output {
		// Create socket message
		socketMsg := NewAgentUpdateMessage(
			agentID,
			agent.Name(),
			agent.Role(),
			msg.Content,
			msg.Type,
		)

		// Send to socket
		err := o.socketServer.SendToAgent(agentID, socketMsg)
		if err != nil {
			log.Error().Err(err).Str("agent_id", agentID).Msg("Failed to send message to socket")
		}

		// Also log for debugging
		log.Debug().
			Str("agent_id", msg.AgentID).
			Str("type", msg.Type).
			Str("content", msg.Content).
			Msg("Agent message")
	}
}

// sendStatusMessage sends a message to the orchestrator status socket
func (o *Orchestrator) sendStatusMessage(message string) {
	if o.socketServer == nil {
		log.Warn().Msg("Socket server not initialized")
		return
	}

	socketMsg := NewStatusUpdateMessage(message)

	err := o.socketServer.SendToStatus(socketMsg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send status message")
	}

	// Also log the status message
	log.Info().Str("status", message).Msg("Orchestrator status")
}

// Shutdown gracefully shuts down the orchestrator
func (o *Orchestrator) Shutdown() error {
	o.sendStatusMessage("🔄 Shutting down orchestrator...")

	// Give a moment for final messages to be sent
	time.Sleep(1 * time.Second)

	o.sendStatusMessage("👋 Orchestrator shutdown complete")

	// Give a moment for the final message to be sent
	time.Sleep(500 * time.Millisecond)

	// Shutdown socket server
	if o.socketServer != nil {
		o.socketServer.Shutdown()
	}

	log.Info().Msg("Orchestrator shutdown complete")
	return nil
}

// GetSessionInfo returns information about the tmux session
func (o *Orchestrator) GetSessionInfo() (string, error) {
	if o.session == nil {
		return "", fmt.Errorf("session not initialized")
	}

	return fmt.Sprintf("Session: %s - Use 'tmux attach -t %s' to connect", o.sessionName, o.sessionName), nil
}
