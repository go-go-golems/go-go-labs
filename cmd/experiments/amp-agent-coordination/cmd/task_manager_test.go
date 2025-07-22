package cmd

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
)

func TestTILAndNotesIntegration(t *testing.T) {
	// Create temporary database
	tmpDB := "/tmp/test_amp_coordination.db"
	defer os.Remove(tmpDB)

	logger := zerolog.New(os.Stdout)
	tm, err := NewTaskManager(tmpDB, logger)
	if err != nil {
		t.Fatalf("Failed to create task manager: %v", err)
	}
	defer tm.Close()

	// Create a project
	project, err := tm.CreateProject("Test Project", "A test project", "Test guidelines", nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create an agent
	agent, err := tm.CreateAgent("Test Agent", nil)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Create a task
	task, err := tm.CreateTask("Test Task", "A test task", nil, project.ID)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Test TIL creation
	til, err := tm.CreateTIL(project.ID, &task.ID, agent.ID, "Learned about SQLite", "SQLite is great for prototyping")
	if err != nil {
		t.Fatalf("Failed to create TIL: %v", err)
	}

	if til.Title != "Learned about SQLite" {
		t.Errorf("Expected TIL title 'Learned about SQLite', got '%s'", til.Title)
	}

	// Create TIL without task (project-level)
	projectTIL, err := tm.CreateTIL(project.ID, nil, agent.ID, "Project Insight", "This project teaches agent coordination")
	if err != nil {
		t.Fatalf("Failed to create project TIL: %v", err)
	}

	// Test TIL listing
	allTILs, err := tm.ListTILs(&project.ID, nil, nil)
	if err != nil {
		t.Fatalf("Failed to list TILs: %v", err)
	}

	if len(allTILs) != 2 {
		t.Errorf("Expected 2 TILs, got %d", len(allTILs))
	}

	// Test task-specific TIL listing
	taskTILs, err := tm.ListTILs(&project.ID, &task.ID, nil)
	if err != nil {
		t.Fatalf("Failed to list task TILs: %v", err)
	}

	if len(taskTILs) != 1 {
		t.Errorf("Expected 1 task TIL, got %d", len(taskTILs))
	}

	// Test project-level TIL listing (empty string for task ID)
	emptyTaskID := ""
	projectOnlyTILs, err := tm.ListTILs(&project.ID, &emptyTaskID, nil)
	if err != nil {
		t.Fatalf("Failed to list project-only TILs: %v", err)
	}

	if len(projectOnlyTILs) != 1 {
		t.Errorf("Expected 1 project-only TIL, got %d", len(projectOnlyTILs))
	}

	// Test Note creation
	note1, err := tm.CreateNote(task.ID, agent.ID, "First note about this task")
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	note2, err := tm.CreateNote(task.ID, agent.ID, "Second note with more details")
	if err != nil {
		t.Fatalf("Failed to create second note: %v", err)
	}

	// Test Note listing
	notes, err := tm.ListNotes(&task.ID, nil)
	if err != nil {
		t.Fatalf("Failed to list notes: %v", err)
	}

	if len(notes) != 2 {
		t.Errorf("Expected 2 notes, got %d", len(notes))
	}

	// Verify notes are ordered by creation time (ASC)
	if notes[0].Content != "First note about this task" {
		t.Errorf("Expected first note content 'First note about this task', got '%s'", notes[0].Content)
	}

	// Test agent-specific note listing
	agentNotes, err := tm.ListNotes(nil, &agent.ID)
	if err != nil {
		t.Fatalf("Failed to list agent notes: %v", err)
	}

	if len(agentNotes) != 2 {
		t.Errorf("Expected 2 agent notes, got %d", len(agentNotes))
	}

	// Test GetTaskWithNotes
	taskWithNotes, err := tm.GetTaskWithNotes(task.ID)
	if err != nil {
		t.Fatalf("Failed to get task with notes: %v", err)
	}

	if taskWithNotes.Task.ID != task.ID {
		t.Errorf("Expected task ID '%s', got '%s'", task.ID, taskWithNotes.Task.ID)
	}

	if len(taskWithNotes.Notes) != 2 {
		t.Errorf("Expected 2 notes in task with notes, got %d", len(taskWithNotes.Notes))
	}

	t.Logf("Created TIL: %+v", til)
	t.Logf("Created project TIL: %+v", projectTIL)
	t.Logf("Created notes: %+v, %+v", note1, note2)
	t.Logf("Task with notes: %+v", taskWithNotes)
}
