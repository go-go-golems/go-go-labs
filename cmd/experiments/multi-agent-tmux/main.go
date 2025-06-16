package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	logLevel    string
	sessionName string
	taskCount   int
	duration    time.Duration
	interactive bool
)

func main() {
	rand.Seed(time.Now().UnixNano())

	rootCmd := &cobra.Command{
		Use:   "multi-agent-tmux",
		Short: "Multi-agent orchestrator with tmux output visualization",
		Long: `A multi-agent system orchestrator that demonstrates:
- Mock LLM agents working concurrently
- Real-time tmux output visualization for each agent
- Orchestrator coordination and status tracking
- Different agent types: Research, Analysis, Writing, Review`,
		Run: runOrchestrator,
	}

	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().StringVar(&sessionName, "session", "multi-agent", "Tmux session name")
	rootCmd.Flags().IntVar(&taskCount, "tasks", 4, "Number of tasks to execute")
	rootCmd.Flags().DurationVar(&duration, "duration", 30*time.Second, "Maximum duration for task execution")
	rootCmd.Flags().BoolVar(&interactive, "interactive", false, "Wait for user input before starting tasks")

	// Add TUI subcommand
	rootCmd.AddCommand(createTUICommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runOrchestrator(cmd *cobra.Command, args []string) {
	setupLogging()

	log.Info().
		Str("session", sessionName).
		Int("tasks", taskCount).
		Dur("duration", duration).
		Msg("Starting multi-agent orchestrator")

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info().Msg("Received shutdown signal")
		cancel()
	}()

	if err := startOrchestrator(ctx); err != nil {
		log.Fatal().Err(err).Msg("Orchestrator failed")
	}

	log.Info().Msg("Multi-agent orchestrator completed successfully")
}

func startOrchestrator(ctx context.Context) error {
	// Create orchestrator
	orchestrator, err := NewOrchestrator(sessionName)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	// Initialize the orchestrator and tmux session
	if err := orchestrator.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize orchestrator: %w", err)
	}

	// Get session info for user
	sessionInfo, err := orchestrator.GetSessionInfo()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get session info")
	} else {
		log.Info().Str("info", sessionInfo).Msg("Session created")
	}

	// Wait for user input if interactive mode
	if interactive {
		orchestrator.sendStatusMessage("⏸️  Interactive mode: Press Enter to start tasks...")
		fmt.Print("Press Enter to start executing tasks...")
		fmt.Scanln()
	}

	// Generate tasks
	tasks := generateTasks(taskCount)

	orchestrator.sendStatusMessage(fmt.Sprintf("📋 Generated %d tasks for execution", len(tasks)))

	// Log task details
	for _, task := range tasks {
		log.Info().
			Str("task_id", task.ID).
			Str("agent_type", task.AgentType).
			Str("description", task.Description).
			Msg("Generated task")
	}

	// Execute tasks
	if err := orchestrator.ExecuteTasks(ctx, tasks); err != nil {
		orchestrator.sendStatusMessage(fmt.Sprintf("❌ Task execution failed: %v", err))
		return fmt.Errorf("failed to execute tasks: %w", err)
	}

	// Keep session alive for a bit to see final results
	orchestrator.sendStatusMessage("✨ All tasks completed! Session will remain active for viewing...")

	if !interactive {
		log.Info().Msg("Keeping session alive for 30 seconds...")
		time.Sleep(30 * time.Second)
	} else {
		orchestrator.sendStatusMessage("⏸️  Interactive mode: Press Enter to shutdown...")
		fmt.Print("Press Enter to shutdown...")
		fmt.Scanln()
	}

	// Shutdown orchestrator
	return orchestrator.Shutdown()
}

func generateTasks(count int) []Task {
	taskTemplates := []struct {
		agentType   string
		description string
	}{
		{"research", "Research latest developments in distributed systems"},
		{"research", "Investigate machine learning optimization techniques"},
		{"research", "Study microservices architecture patterns"},
		{"analysis", "Analyze performance metrics from distributed system"},
		{"analysis", "Evaluate cost-benefit of cloud migration"},
		{"analysis", "Compare different database architectures"},
		{"writing", "Write technical documentation for API endpoints"},
		{"writing", "Create user guide for new features"},
		{"writing", "Draft architecture decision record"},
		{"review", "Review technical documentation for accuracy"},
		{"review", "Evaluate code quality and best practices"},
		{"review", "Assess system design proposals"},
	}

	var tasks []Task
	for i := 0; i < count; i++ {
		template := taskTemplates[rand.Intn(len(taskTemplates))]

		task := Task{
			ID:          fmt.Sprintf("task-%03d", i+1),
			Description: template.description,
			AgentType:   template.agentType,
		}

		tasks = append(tasks, task)
	}

	return tasks
}

func setupLogging() {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
