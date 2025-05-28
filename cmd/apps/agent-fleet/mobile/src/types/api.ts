export type AgentStatus =
  | "active"
  | "idle"
  | "waiting_feedback"
  | "error"
  | "finished"
  | "warning";

export interface AgentStep {
  id: string;
  text: string;
  status: "current" | "completed" | "failed";
  timestamp: string;
}

export interface Agent {
  id: string;
  name: string;
  status: AgentStatus;
  current_task: string;
  current_step: string | null;
  recent_steps: AgentStep[];
  worktree: string;
  files_changed: number;
  lines_added: number;
  lines_removed: number;
  last_commit: string;
  progress: number;
  pending_question: string | null;
  warning_message: string | null;
  error_message: string | null;
  created_at: string;
  updated_at: string;
}

export type EventType =
  | "start"
  | "commit"
  | "question"
  | "success"
  | "error"
  | "info"
  | "command";

export interface Event {
  id: string;
  agent_id: string;
  type: EventType;
  message: string;
  metadata: Record<string, any>;
  timestamp: string;
}

export interface TodoItem {
  id: string;
  agent_id: string;
  text: string;
  completed: boolean;
  current: boolean;
  order: number;
  created_at: string;
  completed_at: string | null;
}

export type TaskStatus =
  | "pending"
  | "assigned"
  | "in_progress"
  | "completed"
  | "failed";
export type TaskPriority = "low" | "medium" | "high" | "urgent";

export interface Task {
  id: string;
  title: string;
  description: string;
  assigned_agent_id: string | null;
  status: TaskStatus;
  priority: TaskPriority;
  created_at: string;
  assigned_at: string | null;
  completed_at: string | null;
}

export type CommandType = "instruction" | "feedback" | "question";
export type CommandStatus = "sent" | "acknowledged" | "completed";

export interface Command {
  id: string;
  agent_id: string;
  content: string;
  type: CommandType;
  response: string | null;
  status: CommandStatus;
  sent_at: string;
  responded_at: string | null;
}

export interface FleetStatus {
  total_agents: number;
  active_agents: number;
  pending_tasks: number;
  agents_needing_feedback: number;
  total_files_changed: number;
  total_commits_today: number;
}

export interface ApiResponse<T> {
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: Record<string, any>;
  };
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface AgentsResponse extends PaginatedResponse<Agent> {
  agents: Agent[];
}

export interface EventsResponse extends PaginatedResponse<Event> {
  events: Event[];
}

export interface TasksResponse extends PaginatedResponse<Task> {
  tasks: Task[];
}

export interface TodosResponse {
  todos: TodoItem[];
}

export interface CommandsResponse {
  commands: Command[];
}

export interface RecentUpdatesResponse {
  updates: Event[];
}

// SSE Event Types
export type SSEEventType =
  | "agent_status_changed"
  | "agent_event_created"
  | "agent_question_posted"
  | "agent_progress_updated"
  | "agent_step_updated"
  | "agent_warning_posted"
  | "agent_error_posted"
  | "todo_updated"
  | "task_assigned"
  | "command_received";

export interface SSEEvent {
  event: SSEEventType;
  data: Record<string, any>;
}
