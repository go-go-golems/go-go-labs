package cmd

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

type Task struct {
	ID                     string     `json:"id"`
	ParentID               *string    `json:"parent_id"`
	Title                  string     `json:"title"`
	Description            string     `json:"description"`
	Status                 TaskStatus `json:"status"`
	AgentID                *string    `json:"agent_id"`
	ProjectID              string     `json:"project_id"`
	PreferredAgentTypeID   *string    `json:"preferred_agent_type_id"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type TaskWithAgentInfo struct {
	Task
	AgentName             *string `json:"agent_name"`
	AgentTypeName         *string `json:"agent_type_name"`
	PreferredAgentTypeName *string `json:"preferred_agent_type_name"`
}

type TaskDependency struct {
	TaskID      string    `json:"task_id"`
	DependsOnID string    `json:"depends_on_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type Agent struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	AgentTypeID *string   `json:"agent_type_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AgentType struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ProjectID   string    `json:"project_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Guidelines  string    `json:"guidelines"`
	AuthorID    *string   `json:"author_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TIL struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	TaskID    *string   `json:"task_id"`
	AgentID   string    `json:"agent_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Note struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	AgentID   string    `json:"agent_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TaskWithNotes struct {
	Task  Task   `json:"task"`
	Notes []Note `json:"notes"`
}

type AgentStats struct {
	Agent       Agent   `json:"agent"`
	Pending     int     `json:"pending"`
	InProgress  int     `json:"in_progress"`
	Completed   int     `json:"completed"`
	Failed      int     `json:"failed"`
	Total       int     `json:"total"`
	SuccessRate float64 `json:"success_rate"`
}

type TaskManager struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewTaskManager(dbPath string, logger zerolog.Logger) (*TaskManager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	tm := &TaskManager{
		db:     db,
		logger: logger,
	}

	if err := tm.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return tm, nil
}

func (tm *TaskManager) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		guidelines TEXT,
		author_id TEXT REFERENCES agents(id),
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS agent_types (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		project_id TEXT NOT NULL REFERENCES projects(id),
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS agents (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'idle',
		agent_type_id TEXT REFERENCES agent_types(id),
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		parent_id TEXT REFERENCES tasks(id),
		title TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		agent_id TEXT REFERENCES agents(id),
		project_id TEXT NOT NULL REFERENCES projects(id),
		preferred_agent_type_id TEXT REFERENCES agent_types(id),
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS task_dependencies (
		task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
		depends_on_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (task_id, depends_on_id)
	);

	CREATE TABLE IF NOT EXISTS global_kv (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		author_id TEXT REFERENCES agents(id),
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tils (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL REFERENCES projects(id),
		task_id TEXT REFERENCES tasks(id),
		agent_id TEXT NOT NULL REFERENCES agents(id),
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS notes (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL REFERENCES tasks(id),
		agent_id TEXT NOT NULL REFERENCES agents(id),
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_agent_id ON tasks(agent_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_preferred_agent_type_id ON tasks(preferred_agent_type_id);
	CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
	CREATE INDEX IF NOT EXISTS idx_agents_type_id ON agents(agent_type_id);
	CREATE INDEX IF NOT EXISTS idx_agent_types_project_id ON agent_types(project_id);
	CREATE INDEX IF NOT EXISTS idx_tils_project_id ON tils(project_id);
	CREATE INDEX IF NOT EXISTS idx_tils_task_id ON tils(task_id);
	CREATE INDEX IF NOT EXISTS idx_tils_agent_id ON tils(agent_id);
	CREATE INDEX IF NOT EXISTS idx_notes_task_id ON notes(task_id);
	CREATE INDEX IF NOT EXISTS idx_notes_agent_id ON notes(agent_id);
	`

	_, err := tm.db.Exec(schema)
	return err
}

func (tm *TaskManager) CreateTask(title, description string, parentID *string, projectID string, preferredAgentTypeID *string) (*Task, error) {
	task := &Task{
		ID:                   uuid.New().String(),
		ParentID:             parentID,
		Title:                title,
		Description:          description,
		Status:               TaskStatusPending,
		ProjectID:            projectID,
		PreferredAgentTypeID: preferredAgentTypeID,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	query := `
		INSERT INTO tasks (id, parent_id, title, description, status, project_id, preferred_agent_type_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := tm.db.Exec(query, task.ID, task.ParentID, task.Title, task.Description, task.Status, task.ProjectID, task.PreferredAgentTypeID, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	tm.logger.Debug().Str("task_id", task.ID).Str("title", task.Title).Msg("Task created")
	return task, nil
}

func (tm *TaskManager) GetTask(taskID string) (*Task, error) {
	query := `
		SELECT id, parent_id, title, description, status, agent_id, project_id, preferred_agent_type_id, created_at, updated_at
		FROM tasks WHERE id = ?
	`

	var task Task
	row := tm.db.QueryRow(query, taskID)
	err := row.Scan(&task.ID, &task.ParentID, &task.Title, &task.Description, &task.Status, &task.AgentID, &task.ProjectID, &task.PreferredAgentTypeID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

func (tm *TaskManager) ListTasks(parentID *string, status *TaskStatus, agentID *string, projectID *string, preferredAgentTypeID *string) ([]Task, error) {
	query := `
		SELECT id, parent_id, title, description, status, agent_id, project_id, preferred_agent_type_id, created_at, updated_at
		FROM tasks WHERE 1=1
	`
	var args []interface{}

	if parentID != nil {
		if *parentID == "" {
			query += " AND parent_id IS NULL"
		} else {
			query += " AND parent_id = ?"
			args = append(args, *parentID)
		}
	}

	if status != nil {
		query += " AND status = ?"
		args = append(args, *status)
	}

	if agentID != nil {
		query += " AND agent_id = ?"
		args = append(args, *agentID)
	}

	if projectID != nil {
		query += " AND project_id = ?"
		args = append(args, *projectID)
	}

	if preferredAgentTypeID != nil {
		query += " AND preferred_agent_type_id = ?"
		args = append(args, *preferredAgentTypeID)
	}

	query += " ORDER BY created_at ASC"

	rows, err := tm.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.ParentID, &task.Title, &task.Description, &task.Status, &task.AgentID, &task.ProjectID, &task.PreferredAgentTypeID, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (tm *TaskManager) ListTasksWithAgentInfo(parentID *string, status *TaskStatus, agentID *string, projectID *string, preferredAgentTypeID *string) ([]TaskWithAgentInfo, error) {
	query := `
		SELECT 
			t.id, t.parent_id, t.title, t.description, t.status, t.agent_id, t.project_id, t.preferred_agent_type_id, t.created_at, t.updated_at,
			a.name as agent_name,
			at.name as agent_type_name,
			pat.name as preferred_agent_type_name
		FROM tasks t
		LEFT JOIN agents a ON t.agent_id = a.id
		LEFT JOIN agent_types at ON a.agent_type_id = at.id
		LEFT JOIN agent_types pat ON t.preferred_agent_type_id = pat.id
		WHERE 1=1
	`
	var args []interface{}

	if parentID != nil {
		if *parentID == "" {
			query += " AND t.parent_id IS NULL"
		} else {
			query += " AND t.parent_id = ?"
			args = append(args, *parentID)
		}
	}

	if status != nil {
		query += " AND t.status = ?"
		args = append(args, *status)
	}

	if agentID != nil {
		query += " AND t.agent_id = ?"
		args = append(args, *agentID)
	}

	if projectID != nil {
		query += " AND t.project_id = ?"
		args = append(args, *projectID)
	}

	if preferredAgentTypeID != nil {
		query += " AND t.preferred_agent_type_id = ?"
		args = append(args, *preferredAgentTypeID)
	}

	query += " ORDER BY t.created_at ASC"

	rows, err := tm.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks with agent info: %w", err)
	}
	defer rows.Close()

	var tasks []TaskWithAgentInfo
	for rows.Next() {
		var task TaskWithAgentInfo
		err := rows.Scan(&task.ID, &task.ParentID, &task.Title, &task.Description, &task.Status, &task.AgentID, &task.ProjectID, &task.PreferredAgentTypeID, &task.CreatedAt, &task.UpdatedAt, &task.AgentName, &task.AgentTypeName, &task.PreferredAgentTypeName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task with agent info: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (tm *TaskManager) AssignTask(taskID, agentID string) error {
	query := `UPDATE tasks SET agent_id = ?, status = ?, updated_at = ? WHERE id = ?`

	_, err := tm.db.Exec(query, agentID, TaskStatusInProgress, time.Now(), taskID)
	if err != nil {
		return fmt.Errorf("failed to assign task: %w", err)
	}

	tm.logger.Debug().Str("task_id", taskID).Str("agent_id", agentID).Msg("Task assigned")
	return nil
}

func (tm *TaskManager) UpdateTaskStatus(taskID string, status TaskStatus) error {
	query := `UPDATE tasks SET status = ?, updated_at = ? WHERE id = ?`

	_, err := tm.db.Exec(query, status, time.Now(), taskID)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	tm.logger.Debug().Str("task_id", taskID).Str("status", string(status)).Msg("Task status updated")
	return nil
}

func (tm *TaskManager) AddDependency(taskID, dependsOnID string) error {
	if taskID == dependsOnID {
		return fmt.Errorf("task cannot depend on itself")
	}

	query := `INSERT INTO task_dependencies (task_id, depends_on_id, created_at) VALUES (?, ?, ?)`

	_, err := tm.db.Exec(query, taskID, dependsOnID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add dependency: %w", err)
	}

	tm.logger.Debug().Str("task_id", taskID).Str("depends_on_id", dependsOnID).Msg("Dependency added")
	return nil
}

func (tm *TaskManager) GetTaskDependencies(taskID string) ([]TaskDependency, error) {
	query := `
		SELECT task_id, depends_on_id, created_at
		FROM task_dependencies WHERE task_id = ?
	`

	rows, err := tm.db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}
	defer rows.Close()

	var deps []TaskDependency
	for rows.Next() {
		var dep TaskDependency
		err := rows.Scan(&dep.TaskID, &dep.DependsOnID, &dep.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		deps = append(deps, dep)
	}

	return deps, nil
}

func (tm *TaskManager) GetAvailableTasks(preferredAgentTypeID *string) ([]Task, error) {
	query := `
		SELECT t.id, t.parent_id, t.title, t.description, t.status, t.agent_id, t.project_id, t.preferred_agent_type_id, t.created_at, t.updated_at
		FROM tasks t
		WHERE t.status = 'pending'
		AND NOT EXISTS (
			SELECT 1 FROM task_dependencies td
			JOIN tasks dep ON td.depends_on_id = dep.id
			WHERE td.task_id = t.id AND dep.status != 'completed'
		)
	`
	var args []interface{}

	if preferredAgentTypeID != nil {
		query += " AND t.preferred_agent_type_id = ?"
		args = append(args, *preferredAgentTypeID)
	}

	query += " ORDER BY t.created_at ASC"

	rows, err := tm.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get available tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.ParentID, &task.Title, &task.Description, &task.Status, &task.AgentID, &task.ProjectID, &task.PreferredAgentTypeID, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (tm *TaskManager) GetAvailableTasksWithAgentInfo(preferredAgentTypeID *string) ([]TaskWithAgentInfo, error) {
	query := `
		SELECT 
			t.id, t.parent_id, t.title, t.description, t.status, t.agent_id, t.project_id, t.preferred_agent_type_id, t.created_at, t.updated_at,
			a.name as agent_name,
			at.name as agent_type_name,
			pat.name as preferred_agent_type_name
		FROM tasks t
		LEFT JOIN agents a ON t.agent_id = a.id
		LEFT JOIN agent_types at ON a.agent_type_id = at.id
		LEFT JOIN agent_types pat ON t.preferred_agent_type_id = pat.id
		WHERE t.status = 'pending'
		AND NOT EXISTS (
			SELECT 1 FROM task_dependencies td
			JOIN tasks dep ON td.depends_on_id = dep.id
			WHERE td.task_id = t.id AND dep.status != 'completed'
		)
	`
	var args []interface{}

	if preferredAgentTypeID != nil {
		query += " AND t.preferred_agent_type_id = ?"
		args = append(args, *preferredAgentTypeID)
	}

	query += " ORDER BY t.created_at ASC"

	rows, err := tm.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get available tasks with agent info: %w", err)
	}
	defer rows.Close()

	var tasks []TaskWithAgentInfo
	for rows.Next() {
		var task TaskWithAgentInfo
		err := rows.Scan(&task.ID, &task.ParentID, &task.Title, &task.Description, &task.Status, &task.AgentID, &task.ProjectID, &task.PreferredAgentTypeID, &task.CreatedAt, &task.UpdatedAt, &task.AgentName, &task.AgentTypeName, &task.PreferredAgentTypeName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan available task with agent info: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (tm *TaskManager) CreateAgent(name string, agentTypeID *string) (*Agent, error) {
	agent := &Agent{
		ID:          uuid.New().String(),
		Name:        name,
		Status:      "idle",
		AgentTypeID: agentTypeID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO agents (id, name, status, agent_type_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := tm.db.Exec(query, agent.ID, agent.Name, agent.Status, agent.AgentTypeID, agent.CreatedAt, agent.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	tm.logger.Debug().Str("agent_id", agent.ID).Str("name", agent.Name).Msg("Agent created")
	return agent, nil
}

func (tm *TaskManager) ListAgents() ([]Agent, error) {
	query := `
		SELECT id, name, status, agent_type_id, created_at, updated_at
		FROM agents ORDER BY created_at ASC
	`

	rows, err := tm.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent
		err := rows.Scan(&agent.ID, &agent.Name, &agent.Status, &agent.AgentTypeID, &agent.CreatedAt, &agent.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

// Project management methods
func (tm *TaskManager) CreateProject(name, description, guidelines string, authorID *string) (*Project, error) {
	project := &Project{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Guidelines:  guidelines,
		AuthorID:    authorID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO projects (id, name, description, guidelines, author_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := tm.db.Exec(query, project.ID, project.Name, project.Description, project.Guidelines, project.AuthorID, project.CreatedAt, project.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	tm.logger.Debug().Str("project_id", project.ID).Str("name", project.Name).Msg("Project created")
	return project, nil
}

func (tm *TaskManager) GetProject(projectID string) (*Project, error) {
	query := `
		SELECT id, name, description, guidelines, author_id, created_at, updated_at
		FROM projects WHERE id = ?
	`

	var project Project
	row := tm.db.QueryRow(query, projectID)
	err := row.Scan(&project.ID, &project.Name, &project.Description, &project.Guidelines, &project.AuthorID, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found: %s", projectID)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

func (tm *TaskManager) GetDefaultProject() (*Project, error) {
	// First check for explicitly set default
	defaultID, err := tm.GetGlobalKV("default_project")
	if err == nil && defaultID != "" {
		return tm.GetProject(defaultID)
	}

	// Fall back to latest project
	query := `
		SELECT id, name, description, guidelines, author_id, created_at, updated_at
		FROM projects ORDER BY created_at DESC LIMIT 1
	`

	var project Project
	row := tm.db.QueryRow(query)
	err = row.Scan(&project.ID, &project.Name, &project.Description, &project.Guidelines, &project.AuthorID, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no projects found")
		}
		return nil, fmt.Errorf("failed to get default project: %w", err)
	}

	return &project, nil
}

func (tm *TaskManager) ListProjects() ([]Project, error) {
	query := `
		SELECT id, name, description, guidelines, author_id, created_at, updated_at
		FROM projects ORDER BY created_at DESC
	`

	rows, err := tm.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		err := rows.Scan(&project.ID, &project.Name, &project.Description, &project.Guidelines, &project.AuthorID, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// Agent type management methods
func (tm *TaskManager) CreateAgentType(name, description, projectID string) (*AgentType, error) {
	agentType := &AgentType{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		ProjectID:   projectID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO agent_types (id, name, description, project_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := tm.db.Exec(query, agentType.ID, agentType.Name, agentType.Description, agentType.ProjectID, agentType.CreatedAt, agentType.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent type: %w", err)
	}

	tm.logger.Debug().Str("agent_type_id", agentType.ID).Str("name", agentType.Name).Msg("Agent type created")
	return agentType, nil
}

func (tm *TaskManager) ListAgentTypes(projectID *string) ([]AgentType, error) {
	query := `
		SELECT id, name, description, project_id, created_at, updated_at
		FROM agent_types WHERE 1=1
	`
	var args []interface{}

	if projectID != nil {
		query += " AND project_id = ?"
		args = append(args, *projectID)
	}

	query += " ORDER BY created_at ASC"

	rows, err := tm.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent types: %w", err)
	}
	defer rows.Close()

	var agentTypes []AgentType
	for rows.Next() {
		var agentType AgentType
		err := rows.Scan(&agentType.ID, &agentType.Name, &agentType.Description, &agentType.ProjectID, &agentType.CreatedAt, &agentType.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent type: %w", err)
		}
		agentTypes = append(agentTypes, agentType)
	}

	return agentTypes, nil
}

func (tm *TaskManager) AssignTaskToAgentType(taskID, agentTypeID string) error {
	// Find an available agent of this type
	query := `
		SELECT id FROM agents 
		WHERE agent_type_id = ? AND status = 'idle' 
		ORDER BY created_at ASC LIMIT 1
	`

	var agentID string
	err := tm.db.QueryRow(query, agentTypeID).Scan(&agentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no available agents of type %s", agentTypeID)
		}
		return fmt.Errorf("failed to find agent: %w", err)
	}

	return tm.AssignTask(taskID, agentID)
}

// Global KV methods
func (tm *TaskManager) SetGlobalKV(key, value string, authorID *string) error {
	query := `
		INSERT OR REPLACE INTO global_kv (key, value, author_id, created_at, updated_at)
		VALUES (?, ?, ?, COALESCE((SELECT created_at FROM global_kv WHERE key = ?), CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)
	`

	_, err := tm.db.Exec(query, key, value, authorID, key)
	if err != nil {
		return fmt.Errorf("failed to set global kv: %w", err)
	}

	tm.logger.Debug().Str("key", key).Str("value", value).Msg("Global KV set")
	return nil
}

func (tm *TaskManager) GetGlobalKV(key string) (string, error) {
	query := `SELECT value FROM global_kv WHERE key = ?`

	var value string
	err := tm.db.QueryRow(query, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get global kv: %w", err)
	}

	return value, nil
}

// TIL management methods
func (tm *TaskManager) CreateTIL(projectID string, taskID *string, agentID string, title string, content string) (*TIL, error) {
	til := &TIL{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		TaskID:    taskID,
		AgentID:   agentID,
		Title:     title,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO tils (id, project_id, task_id, agent_id, title, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := tm.db.Exec(query, til.ID, til.ProjectID, til.TaskID, til.AgentID, til.Title, til.Content, til.CreatedAt, til.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create TIL: %w", err)
	}

	tm.logger.Debug().Str("til_id", til.ID).Str("title", til.Title).Msg("TIL created")
	return til, nil
}

func (tm *TaskManager) ListTILs(projectID *string, taskID *string, agentID *string) ([]TIL, error) {
	query := `
		SELECT id, project_id, task_id, agent_id, title, content, created_at, updated_at
		FROM tils WHERE 1=1
	`
	var args []interface{}

	if projectID != nil {
		query += " AND project_id = ?"
		args = append(args, *projectID)
	}

	if taskID != nil {
		if *taskID == "" {
			query += " AND task_id IS NULL"
		} else {
			query += " AND task_id = ?"
			args = append(args, *taskID)
		}
	}

	if agentID != nil {
		query += " AND agent_id = ?"
		args = append(args, *agentID)
	}

	query += " ORDER BY created_at DESC"

	rows, err := tm.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list TILs: %w", err)
	}
	defer rows.Close()

	var tils []TIL
	for rows.Next() {
		var til TIL
		err := rows.Scan(&til.ID, &til.ProjectID, &til.TaskID, &til.AgentID, &til.Title, &til.Content, &til.CreatedAt, &til.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan TIL: %w", err)
		}
		tils = append(tils, til)
	}

	return tils, nil
}

// Note management methods
func (tm *TaskManager) CreateNote(taskID string, agentID string, content string) (*Note, error) {
	note := &Note{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		AgentID:   agentID,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO notes (id, task_id, agent_id, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := tm.db.Exec(query, note.ID, note.TaskID, note.AgentID, note.Content, note.CreatedAt, note.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	tm.logger.Debug().Str("note_id", note.ID).Str("task_id", note.TaskID).Msg("Note created")
	return note, nil
}

func (tm *TaskManager) ListNotes(taskID *string, agentID *string) ([]Note, error) {
	query := `
		SELECT id, task_id, agent_id, content, created_at, updated_at
		FROM notes WHERE 1=1
	`
	var args []interface{}

	if taskID != nil {
		query += " AND task_id = ?"
		args = append(args, *taskID)
	}

	if agentID != nil {
		query += " AND agent_id = ?"
		args = append(args, *agentID)
	}

	query += " ORDER BY created_at ASC"

	rows, err := tm.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.TaskID, &note.AgentID, &note.Content, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, note)
	}

	return notes, nil
}

func (tm *TaskManager) GetTaskWithNotes(taskID string) (*TaskWithNotes, error) {
	task, err := tm.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	notes, err := tm.ListNotes(&taskID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get task notes: %w", err)
	}

	return &TaskWithNotes{
		Task:  *task,
		Notes: notes,
	}, nil
}

func (tm *TaskManager) Close() error {
	return tm.db.Close()
}
