# Agent Fleet Mobile App - Development Specification

## Project Overview

### Purpose
A mobile-first application for monitoring and controlling a fleet of coding agents. Users can view agent status, send commands, track progress, and manage tasks through an intuitive interface.

### Target Platform
- **Primary**: Mobile web (responsive)
- **Secondary**: Native mobile apps (future consideration)
- **Orientation**: Portrait-first

### Core Features
- Real-time agent fleet monitoring
- Individual agent detailed views
- Command/feedback interface
- Task management
- Event history tracking
- Todo list progression
- Live status updates

---

## Technical Requirements

### Frontend Stack
- **Framework**: expo + rtk-toolkit
- **Icons**: Emoji (no external icon libraries)
- **API Communication**: REST API + Server-Sent Events

---

## Architecture

### State Architecture
```javascript
// Global App State
{
  agents: Agent[],
  selectedAgent: Agent | null,
  activeTab: 'fleet' | 'updates' | 'tasks',
  recentUpdates: Event[],
  tasks: Task[],
  connectionStatus: 'connected' | 'disconnected' | 'reconnecting',
  loading: boolean,
  error: string | null
}

// Agent Detail State
{
  showLogs: boolean,
  command: string,
  events: Event[],
  todos: TodoItem[]
}
```

---

## Screen Specifications

### 1. Fleet View (`/fleet`)

#### Layout
- **Header**: Fixed top bar with app title and sync button
- **Content**: Scrollable agent cards grid
- **Navigation**: Fixed bottom navigation

#### Agent Card Specifications
```
┌─────────────────────────────────────┐
│ ● Status    Agent Name    [Badge]   │ ← Header (48px height)
│                           Status    │
├─────────────────────────────────────┤
│ Current task description            │ ← Task (32px height)
├─────────────────────────────────────┤
│ [Question box if pending]           │ ← Optional (variable)
├─────────────────────────────────────┤
│ 🌿 branch      📝 2m ago           │ ← Metadata (24px)
│ 📁 12 files   +347   -89          │ ← Stats (24px)
├─────────────────────────────────────┤
│ [████████████░░░░] 75%             │ ← Progress (16px)
└─────────────────────────────────────┘
```

#### Interactions
- **Tap Agent Card**: Open agent detail modal
- **Pull to Refresh**: Refresh agent data
- **Sync All Button**: Trigger fleet sync

#### Visual States
- **Normal**: Gray border (`border-gray-700`)
- **Needs Feedback**: Orange border (`border-orange-500`) with pulsing animation
- **Hover/Active**: Lighter border (`border-gray-600`)

### 2. Updates View (`/updates`)

#### Layout
```
┌─────────────────────────────────────┐
│ Recent Updates                      │ ← Header
├─────────────────────────────────────┤
│ Agent Name           2m ago         │ ← Update item header
│ ● Success message text              │ ← Status + message
├─────────────────────────────────────┤
│ Agent Name           5m ago         │
│ ● Info message text                 │
└─────────────────────────────────────┘
```

#### Update Item Colors
- **Success**: `text-green-400`
- **Warning**: `text-yellow-400`
- **Error**: `text-red-400`
- **Info**: `text-blue-400`

### 3. Tasks View (`/tasks`)

#### Layout
- **Task Input**: Textarea with submit button
- **Queue Section**: List of pending tasks
- **Empty State**: Centered message when no tasks

#### Task Input Specifications
```
┌─────────────────────────────────────┐
│ Submit Task                         │ ← Header
├─────────────────────────────────────┤
│ ┌─────────────────────────────────┐ │
│ │ Describe the task...            │ │ ← Textarea (4 rows)
│ │                                 │ │
│ │                                 │ │
│ └─────────────────────────────────┘ │
│ Task will be distributed... [🚀 Submit] │ ← Footer
└─────────────────────────────────────┘
```

### 4. Agent Detail Modal

#### Layout Sections (scrollable)
1. **Header**: Agent name with status indicator and close button
2. **Command Interface**: Input field for sending commands/feedback
3. **Task Info**: Current task description
4. **Stats Grid**: Worktree and progress
5. **Todo List**: Interactive checklist
6. **Event History**: Chronological event list
7. **Debug Logs**: Collapsible technical logs

#### Command Interface States
- **Feedback Mode**: Orange styling when agent has pending question
- **Command Mode**: Blue styling for active agents
- **Disabled**: Gray styling for idle agents

#### Todo List Specifications
```
┌─────────────────────────────────────┐
│ 📋 Todo List                        │ ← Header
├─────────────────────────────────────┤
│ ✓ Completed task (strikethrough)    │ ← Completed
│ ● Current task (highlighted)        │ ← Current
│ ○ Pending task                      │ ← Pending
└─────────────────────────────────────┘
```

---

# Agent Fleet Mobile App - User Experience Overview

## What It Is

The Agent Fleet app is a mobile interface for managing multiple AI coding agents working on software projects. Think of it as a mission control center where a developer can monitor, guide, and coordinate a team of AI assistants, each working on different parts of a codebase.

The app addresses a fundamental challenge: as AI agents become more capable, they still need human guidance for complex decisions, creative choices, and when they encounter ambiguous situations. However, these requests for help come at unpredictable times, and developers need to stay responsive while remaining mobile and productive.

## Core User Experience

### The Mental Model
Users think of their AI agents as semi-autonomous team members. Each agent has its own personality, strengths, and current responsibilities. Like managing a distributed team, users need awareness of what everyone is doing, the ability to provide guidance when asked, and confidence that work is progressing smoothly.

### Primary User Flow
The typical interaction starts with the user checking their "fleet" - a dashboard showing all active agents, their current tasks, and most importantly, which ones need attention. Agents requiring feedback are immediately obvious through visual indicators and brief question previews.

When an agent has a question, the user taps into a detailed view where they can see the full context: what the agent is working on, what specific guidance they need, the agent's progress through their todo list, and recent activity. The user provides direction through a simple chat-like interface, and the agent continues working.

The experience feels conversational but asynchronous - users can catch up on what happened while they were away, provide guidance when needed, and assign new work as priorities change.

## Three Core Spaces

**Fleet View** serves as the primary dashboard. Users scan their agents' status at a glance, similar to checking a team's Slack statuses. The emphasis is on immediate awareness - which agents are working productively, which need help, and what progress is being made. This view optimizes for quick triage and decision-making.

**Updates Feed** provides the narrative thread of what's been happening. Users catch up on completed work, see patterns of progress, and identify any emerging issues. This satisfies the human need to understand the story of the work, not just the current state.

**Task Assignment** offers a simple way to add new work to the queue. Users describe what they need in natural language, and the system handles distribution to appropriate agents. This maintains the feeling that agents are capable team members who can interpret and execute on higher-level direction.

## Key Interaction Principles

### Immediate Awareness
The app makes agent status immediately apparent. Users shouldn't have to dig for critical information - questions requiring feedback, progress updates, and potential issues surface prominently. The interface uses visual hierarchy and selective attention-grabbing to guide focus to what matters most.

### Contextual Communication
When users need to interact with an agent, they see the full picture: current task, progress through subtasks, recent activity, and the specific question or issue. This context helps users provide better guidance and makes the interaction feel more collaborative than transactional.

### Graceful Interruption
Since agent questions come at unpredictable times, the interface accommodates interruption-driven workflows. Users can quickly assess what's needed, provide guidance, and return to their other work. The app remembers context and makes it easy to dive back in later.

### Progressive Detail
Information flows from general to specific. The fleet view shows what users need for triage. Agent details reveal the full context for decision-making. Debug logs and technical details are available but don't clutter the primary experience.

## Design Philosophy

### Calm Technology
The interface stays out of the way until attention is needed. Agents working smoothly fade into the background, while those requiring input become prominent. This prevents notification fatigue while ensuring important requests don't get missed.

### Trust Through Transparency
Users can see what agents are thinking through their todo lists, understand their progress through metrics and activity feeds, and review their decision-making through event logs. This transparency builds confidence in agent autonomy while providing oversight.

### Mobile-First Collaboration
The experience acknowledges that meaningful development management can happen away from a desk. Touch interactions, quick scanning, and essential-information-first design enable effective agent coordination from anywhere.

### Human-AI Partnership
Rather than positioning AI agents as tools to be operated, the interface treats them as collaborators to be guided. The communication feels more like directing team members than programming software - natural language, conversational feedback, and acknowledgment of agent initiative.

## Emotional Experience

Users should feel **in control but not burdened**. They have oversight and influence over their agent team without micromanagement. The interface builds **confidence through awareness** - users know what's happening and trust that they'll be informed when input is needed.

The experience should reduce anxiety around AI unpredictability by providing transparency and control, while celebrating the productivity gains of having capable AI assistants handling routine development work.

**Success feels like**: Checking your phone during a coffee break, seeing your agents have made solid progress on their tasks, quickly answering one thoughtful question, and knowing your development work is moving forward even when you're not actively coding.

This app transforms AI agent management from a technical challenge into an intuitive, mobile-friendly collaboration experience that fits naturally into a developer's workflow.