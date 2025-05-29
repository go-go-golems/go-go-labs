# Agent Fleet API Implementations Documentation

## Overview

The Agent Fleet system consists of three main components that communicate via REST API and Server-Sent Events (SSE):

1. **Backend Server** (`cmd/apps/agent-fleet/backend/`) - Central API server and data management
2. **Mock Agent** (`cmd/apps/agent-fleet/mock-agent/`) - Simulated coding agents for testing
3. **Mobile App** (`cmd/apps/agent-fleet/mobile/`) - React Native monitoring interface

## Component Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Mobile App    │◄───│  Backend Server │◄───│   Mock Agent    │
│  (React Native) │    │    (Go/Chi)     │    │     (Go)        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
    UI Components          REST API + SSE         Agent Simulation
    Redux Store           SQLite Database         Scenario Engine
    RTK Query             HTTP Handlers           API Client
```

## 1. Backend Server (`cmd/apps/agent-fleet/backend/`)

### Entry Point
- **File**: `main.go`
- **Purpose**: Server initialization, routing, middleware setup
- **Key Components**:
  - Chi router setup
  - CORS configuration
  - Authentication middleware (optional with `--disable-auth`)
  - Database initialization
  - SSE manager setup

### API Routes Structure (`main.go:136-183`)

```go
r.Route("/v1", func(r chi.Router) {
    // Authentication middleware
    if !viper.GetBool("disable-auth") {
        r.Use(auth.BearerTokenMiddleware)
    }
    
    // Agents API
    r.Route("/agents", func(r chi.Router) {
        r.Get("/", h.ListAgents)                          // GET /v1/agents
        r.Post("/", h.CreateAgent)                        // POST /v1/agents
        r.Get("/{agentID}", h.GetAgent)                   // GET /v1/agents/{id}
        r.Patch("/{agentID}", h.UpdateAgent)              // PATCH /v1/agents/{id}
        r.Delete("/{agentID}", h.DeleteAgent)             // DELETE /v1/agents/{id}
        
        // Agent Events
        r.Get("/{agentID}/events", h.ListAgentEvents)     // GET /v1/agents/{id}/events
        r.Post("/{agentID}/events", h.CreateAgentEvent)   // POST /v1/agents/{id}/events
        
        // Agent Todos
        r.Get("/{agentID}/todos", h.ListAgentTodos)       // GET /v1/agents/{id}/todos
        r.Post("/{agentID}/todos", h.CreateAgentTodo)     // POST /v1/agents/{id}/todos
        r.Patch("/{agentID}/todos/{todoID}", h.UpdateAgentTodo)     // PATCH /v1/agents/{id}/todos/{todoId}
        r.Delete("/{agentID}/todos/{todoID}", h.DeleteAgentTodo)    // DELETE /v1/agents/{id}/todos/{todoId}
        
        // Agent Commands
        r.Get("/{agentID}/commands", h.ListAgentCommands)   // GET /v1/agents/{id}/commands
        r.Post("/{agentID}/commands", h.CreateAgentCommand) // POST /v1/agents/{id}/commands
        r.Patch("/{agentID}/commands/{commandID}", h.UpdateAgentCommand) // PATCH /v1/agents/{id}/commands/{cmdId}
    })
    
    // Tasks API
    r.Route("/tasks", func(r chi.Router) {
        r.Get("/", h.ListTasks)                           // GET /v1/tasks
        r.Post("/", h.CreateTask)                         // POST /v1/tasks
        r.Get("/{taskID}", h.GetTask)                     // GET /v1/tasks/{id}
        r.Patch("/{taskID}", h.UpdateTask)                // PATCH /v1/tasks/{id}
        r.Delete("/{taskID}", h.DeleteTask)               // DELETE /v1/tasks/{id}
    })
    
    // Fleet Operations
    r.Route("/fleet", func(r chi.Router) {
        r.Get("/status", h.GetFleetStatus)                // GET /v1/fleet/status
        r.Get("/recent-updates", h.GetRecentUpdates)      // GET /v1/fleet/recent-updates
    })
    
    // Server-Sent Events
    r.Get("/stream", h.SSEHandler)                        // GET /v1/stream
})
```

### Key Handler Files

#### 1. Agent Handlers (`handlers/agents.go`)
- **Purpose**: CRUD operations for agents
- **Key Methods**:
  - `ListAgents()` - Paginated agent listing with filtering
  - `CreateAgent()` - Agent registration
  - `GetAgent()` - Single agent details
  - `UpdateAgent()` - Agent status/data updates
  - `DeleteAgent()` - Agent removal

#### 2. Event Handlers (`handlers/events.go`)
- **Purpose**: Agent event management
- **Key Methods**:
  - `CreateAgentEvent()` - Creates events and broadcasts SSE notifications
  - `ListAgentEvents()` - Event history for agents
- **Enhanced Error Handling**: Creates specific SSE events for error/warning types

#### 3. Command Handlers (`handlers/commands.go`)
- **Purpose**: Command/feedback system between users and agents
- **Key Methods**:
  - `CreateAgentCommand()` - Send commands/feedback to agents
  - `ListAgentCommands()` - Command history
  - `UpdateAgentCommand()` - Mark commands as acknowledged/completed

#### 4. SSE Manager (`handlers/sse.go`)
- **Purpose**: Real-time event broadcasting
- **Key SSE Events**:
  - `agent_status_changed` - Agent status updates
  - `agent_event_created` - General agent events
  - `agent_question_posted` - Agent questions requiring feedback
  - `agent_warning_posted` - Agent warnings
  - `agent_error_posted` - Agent errors (NEW)
  - `agent_progress_updated` - Progress updates
  - `agent_step_updated` - Current step changes
  - `todo_updated` - Todo item changes
  - `task_assigned` - Task assignments
  - `command_received` - New commands for agents

### Database Layer
- **Technology**: SQLite with custom Go models
- **Location**: `database/` directory
- **Models**: Agents, Events, Commands, Todos, Tasks

## 2. Mock Agent (`cmd/apps/agent-fleet/mock-agent/`)

### Entry Point
- **File**: `main.go`
- **Purpose**: Agent simulation with configurable behavior
- **Configuration Options**:
  - `--server` - Backend server URL
  - `--token` - Authentication token
  - `--name` - Agent name (auto-generated if not provided)
  - `--worktree` - Simulated workspace path
  - `--randomized` - Enable randomized behavior
  - `--tick-interval` - Agent activity frequency
  - `--command-check-interval` - Command polling frequency

### Agent Simulation Engine

#### 1. Agent Core (`agent/agent.go`)
- **Purpose**: Main agent lifecycle and behavior
- **Key Methods**:
  - `Run(ctx)` - Main agent loop
  - `registerAgent()` - Initial registration with backend
  - `tick()` - Periodic agent activity
  - `checkForCommands()` - Poll for user commands
  - `processCommand()` - Handle feedback/instructions

#### 2. API Client (`client/client.go`)
- **Purpose**: HTTP client for backend communication
- **Key Methods**:
  - `RegisterAgent()` - POST /v1/agents
  - `UpdateAgentStatus()` - PATCH /v1/agents/{id}
  - `CreateEvent()` - POST /v1/agents/{id}/events
  - `CreateTodo()` - POST /v1/agents/{id}/todos
  - `UpdateTodo()` - PATCH /v1/agents/{id}/todos/{id}
  - `GetCommands()` - GET /v1/agents/{id}/commands
  - `UpdateCommand()` - PATCH /v1/agents/{id}/commands/{id}

#### 3. Scenario Engine (`scenarios/scenarios.go`)
- **Purpose**: Realistic agent behavior simulation
- **Scenarios Available**:
  - Code Review with feedback requests
  - Bug Fixing with error conditions
  - Feature Implementation with step tracking
  - Testing with progress updates
  - Infrastructure/CI-CD with warnings
- **Behavior Configuration**:
  - Question probability (currently 0.05 = 5%)
  - Error probability
  - Warning probability
  - Step duration variability

### Agent Behavior Flow

```
1. Agent starts → Register with backend
2. Periodic tick → Execute scenario step
3. Update progress → Send events to backend
4. Check commands → Process user feedback
5. Generate questions/errors → Trigger SSE events
6. Update todos → Track progress
7. Repeat until scenario complete
```

## 3. Mobile App (`cmd/apps/agent-fleet/mobile/`)

### Entry Point
- **File**: `App.tsx`
- **Purpose**: React Native app with Redux store setup

### Architecture Overview

#### 1. API Layer (`services/api.ts`)
- **Technology**: RTK Query
- **Base URL**: Configurable (default: `http://192.168.0.79:8080/v1/`)
- **Authentication**: Bearer token (currently hardcoded for demo)
- **Key Endpoints**:
  - `getAgents` - GET /v1/agents
  - `getAgent` - GET /v1/agents/{id}
  - `getAgentEvents` - GET /v1/agents/{id}/events
  - `getAgentTodos` - GET /v1/agents/{id}/todos
  - `sendCommand` - POST /v1/agents/{id}/commands
  - `getFleetStatus` - GET /v1/fleet/status
  - `getRecentUpdates` - GET /v1/fleet/recent-updates

#### 2. Real-time Updates (`services/sse.ts`)
- **Purpose**: SSE connection management with Redux integration
- **Features**:
  - Automatic reconnection on failure
  - Event-based cache invalidation
  - Notification creation for questions/warnings/errors
- **Event Handling**:
  - Invalidates RTK Query caches on relevant events
  - Creates UI notifications for user interaction needs
  - Comprehensive debug logging

#### 3. State Management (`store/`)
- **Technology**: Redux Toolkit
- **Slices**:
  - `api` (RTK Query) - Server data caching
  - `ui` - UI state, notifications, modal visibility

#### 4. UI Components (`components/`)

##### AgentCard (`components/AgentCard.tsx`)
- **Features**:
  - Status indicators with color coding
  - Current step and last 3 recent steps display
  - Question/warning/error boxes with prominence
  - Quick response button for questions
  - Progress tracking

##### NotificationBanner (`components/NotificationBanner.tsx`)
- **Purpose**: Prominent display of urgent agent notifications
- **Features**:
  - Top-of-screen positioning
  - Color-coded by notification type
  - Tap-to-respond for questions
  - Dismissible interface

##### FeedbackModal (`components/FeedbackModal.tsx`)
- **Purpose**: Full-screen interface for responding to agent questions
- **Features**:
  - Agent context display
  - Response type selection (feedback vs instruction)
  - Rich text input
  - Command sending integration

#### 5. Screens (`screens/`)

##### FleetScreen (`screens/FleetScreen.tsx`)
- **Purpose**: Main agent monitoring interface
- **Features**:
  - Real-time agent list (filters out finished agents)
  - Fleet statistics
  - Pull-to-refresh
  - SSE connection management
  - Notification banner integration

## Data Flow Architecture

### 1. Agent Registration & Updates

```
Mock Agent → POST /v1/agents → Backend → Database
Mock Agent → PATCH /v1/agents/{id} → Backend → SSE: agent_status_changed → Mobile App
```

### 2. Event Creation & Broadcasting

```
Mock Agent → POST /v1/agents/{id}/events → Backend → Database
                                                  ↓
                                        SSE: agent_event_created → Mobile App
                                                  ↓
                                   (if error/warning type)
                                                  ↓
                              SSE: agent_error_posted/agent_warning_posted → Mobile App
                                                  ↓
                                         Notification Creation → UI Display
```

### 3. Question/Feedback Flow

```
Mock Agent → Question Event → Backend → SSE: agent_question_posted → Mobile App
                                                                          ↓
                                                               Notification Banner
                                                                          ↓
                                                               User Taps Response
                                                                          ↓
                                                                 Feedback Modal
                                                                          ↓
Mobile App → POST /v1/agents/{id}/commands → Backend → Database
                                                  ↓
                                        SSE: command_received → Mock Agent
                                                  ↓
                                           Agent Processes Command
```

### 4. Real-time Synchronization

```
Any Change → Backend → SSE Event → Mobile App → RTK Query Cache Invalidation → UI Update
```

## API Event Types & Data Structures

### SSE Events
| Event Type | Data Fields | Triggered By | Consumed By |
|------------|-------------|--------------|-------------|
| `agent_status_changed` | agent_id, old_status, new_status, agent | Agent updates | Mobile UI |
| `agent_event_created` | agent_id, event | Event creation | Mobile UI |
| `agent_question_posted` | agent_id, question, agent | Question events | Mobile notifications |
| `agent_warning_posted` | agent_id, warning, agent | Warning events | Mobile notifications |
| `agent_error_posted` | agent_id, error, agent | Error events | Mobile notifications |
| `agent_progress_updated` | agent_id, progress, files_changed, lines_added, lines_removed | Progress updates | Mobile UI |
| `agent_step_updated` | agent_id, current_step, recent_steps | Step changes | Mobile UI |
| `todo_updated` | agent_id, todo, action | Todo changes | Mobile UI |
| `command_received` | agent_id, command | Command creation | Mock Agent |

### Agent Data Structure
```typescript
interface Agent {
  id: string;
  name: string;
  status: 'active' | 'idle' | 'waiting_feedback' | 'error' | 'finished' | 'warning';
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
```

## Security & Configuration

### Authentication
- **Backend**: Bearer token middleware (optional with `--disable-auth`)
- **Mock Agent**: Token passed via `--token` flag
- **Mobile App**: Hardcoded demo token (configurable in `api.ts`)

### CORS Configuration
- **Allowed Origins**: `*` (development setting)
- **Allowed Methods**: GET, POST, PATCH, DELETE, OPTIONS
- **Allowed Headers**: `*`

### Environment Configuration
- **Backend**: Configurable via flags, environment variables, or YAML config
- **Mock Agent**: CLI flags with YAML config support
- **Mobile App**: Compile-time configuration in source files

## Debugging & Monitoring

### Logging
- **Backend**: Zerolog with configurable levels
- **Mock Agent**: Zerolog with console output
- **Mobile App**: Console logging for SSE events and API calls

### Development Tools
- **Backend**: `--dev` flag for enhanced console logging
- **Mock Agent**: `--randomized` flag for varied behavior
- **Mobile App**: React Native debugging, Redux DevTools support

This architecture provides a comprehensive, real-time agent monitoring and control system with proper separation of concerns and robust error handling.
