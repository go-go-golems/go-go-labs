# SQLite State-Machine Agent – Quick-Start & Reference

## Overview

State machine agents are autonomous programs that follow predefined workflows stored in a SQLite database. Each agent knows its current state, available transitions, and the actions to perform when moving between states. This system enables agents to operate independently while maintaining consistent behavior through declarative workflow definitions.

**Who should read this**: This guide targets autonomous agents (human or programmatic) that need to execute a pre-defined workflow stored in SQLite.

## Prerequisites

Before starting, ensure you have:

- SQLite 3.41+ installed (`sqlite3` command available)
- Access to the `research_agents.db` database file
- A loaded workflow (ResearchPaperScoutAgent is pre-installed)
- Database connection capability (CLI, Go database/sql, Python sqlite3, etc.)

**Database setup**: If you need to recreate the database, run:
```bash
sqlite3 research_agents.db < init_database.sql
```

## Database Schema Overview

The system uses four core tables that work together:

```
workflow ─┬─< workflow_state ─┬─< workflow_transition
          │                   └─< state guidelines
          └─< agent_instance
          
agent_next_actions (view) ──> Single source of truth for agents
```

**Key relationships**:
- `workflow` defines the state machine
- `workflow_state` contains your action guidelines  
- `workflow_transition` defines legal moves between states
- `agent_instance` tracks your current position
- `agent_next_actions` view shows current state + available transitions

## Core Concepts

**Workflow**: A complete state machine definition with states and transitions  
**State**: A specific point in the workflow with guidelines for what the agent should do  
**Transition**: A named event that moves the agent from one state to another with specific actions  
**Agent Instance**: A running agent tied to a workflow and currently in a specific state

## Agent Execution Loop

Follow this five-step cycle until you reach the 'End' state:

### Step 1: Register Yourself in the Database

Create your agent instance in the system:

```sql
INSERT INTO agent_instance (workflow_id, agent_name, current_state_id) 
VALUES (
    (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'),
    'Agent-[YOUR-ID]',
    (SELECT state_id FROM workflow_state WHERE name = 'Start')
)
RETURNING agent_id;
```

### Step 2: Fetch Context via agent_next_actions

Query your current state and available actions:

```sql
SELECT * FROM agent_next_actions WHERE agent_id = ?;
```

The `agent_next_actions` view returns:
- `current_state`: Your current position in the workflow
- `state_guideline`: Detailed instructions for what to do in this state
- `next_transition`: Available transition name (NULL if terminal state)
- `next_state`: Destination state for each transition
- `transition_description`: What the transition does

### Step 3: Execute State Guideline

Follow the `state_guideline` precisely to perform the work required in your current state. These guidelines tell you exactly what to do and what to prepare for the next transition.

### Step 4: Commit Transition (Atomic Update)

When ready to move to the next state, update your position atomically:

```sql
BEGIN IMMEDIATE;
UPDATE agent_instance 
SET current_state_id = (
    SELECT to_state_id FROM workflow_transition 
    WHERE name = ? AND from_state_id = current_state_id
)
WHERE agent_id = ?;
COMMIT;
```

### Step 5: Repeat or Finish

Return to Step 2 and continue the cycle until `next_transition` is NULL (terminal state reached).

## Schema Reference (Cheat-Sheet)

### Essential Tables

**agent_instance**: Your runtime state
- `agent_id`: Your unique identifier
- `current_state_id`: Where you are in the workflow
- `updated_at`: Automatically updated on state changes

**agent_next_actions** (VIEW): Your command center
- Query this view to see current state guidelines and available transitions
- Contains all information needed to decide your next action
- Automatically filtered by agent_id - your single source of truth

**agent_state_log**: Audit trail of all state changes
- Automatically populated by database triggers
- Tracks every transition with timestamps and transition names
- Use for debugging, monitoring, and workflow analysis

**agent_transition_history** (VIEW): Human-readable audit log
- Shows complete transition history with state names
- Useful for analyzing agent behavior and workflow patterns

### Example Workflow States

The ResearchPaperScoutAgent workflow includes these key states:

- **Start**: Initialize and prepare for research session
- **LoadQueries**: Validate research queries from configuration
- **SearchWeb**: Execute searches across academic databases
- **ExtractMetadata**: Parse paper identifiers, titles, abstracts
- **Deduplicate**: Compare against existing archive
- **SummariseAndScore**: Generate summaries and relevance scores
- **AskUser**: Present findings for human review
- **SaveAccepted**: Archive approved papers
- **Report**: Generate session summary
- **End**: Clean up and terminate

## Error Handling & Recovery

### Database-Level Errors
- **Foreign key violations**: Invalid transition attempted - check workflow definition
- **Constraint failures**: Duplicate agent names or missing references

### Logical Errors
- **No available transitions**: You've reached terminal state or workflow misconfiguration
- **Invalid state updates**: Database will reject transitions not defined in workflow_transition table

### Concurrency Issues
- **Multiple processes updating same agent**: Use `BEGIN IMMEDIATE` to prevent conflicts
- **Deadlocks**: Implement exponential backoff and retry logic

### Recovery Patterns
1. **Retry with backoff**: For transient database connection issues
2. **Log and escalate**: For workflow logic errors requiring human intervention
3. **Reset to known state**: If agent becomes corrupted, restart from 'Start' state

## Best Practices

1. **Query agent_next_actions first** before taking any action
2. **Follow state guidelines precisely** - they contain specific work requirements
3. **Wrap state execution + transition in one transaction** where possible
4. **Automatic logging**: All state transitions are logged automatically via database triggers
5. **Handle errors gracefully** - some states may have error recovery transitions
6. **Use parameterized queries** to prevent SQL injection

## Quick Reference Commands

```sql
-- Check your current status and available actions
SELECT * FROM agent_next_actions WHERE agent_id = ?;

-- Move to next state by transition name (atomic)
BEGIN IMMEDIATE;
UPDATE agent_instance SET current_state_id = (
    SELECT to_state_id FROM workflow_transition 
    WHERE name = ? AND from_state_id = current_state_id
) WHERE agent_id = ?;
COMMIT;

-- Check if you're in terminal state
SELECT COUNT(*) FROM workflow_transition 
WHERE from_state_id = (SELECT current_state_id FROM agent_instance WHERE agent_id = ?);

-- Reset agent to start state (recovery)
UPDATE agent_instance SET current_state_id = (
    SELECT state_id FROM workflow_state WHERE name = 'Start'
) WHERE agent_id = ?;

-- View your complete transition history
SELECT from_state, to_state, transition_name, logged_at 
FROM agent_transition_history 
WHERE agent_id = ? 
ORDER BY logged_at;

-- Count transitions by state (workflow analysis)
SELECT to_state, COUNT(*) as visit_count
FROM agent_transition_history 
WHERE agent_id = ? 
GROUP BY to_state 
ORDER BY visit_count DESC;
```

## Appendix: Example Go Agent Loop

```go
func runAgent(db *sql.DB, agentID int) error {
    for {
        // Step 2: Fetch current context
        var state, guideline, transition, description string
        err := db.QueryRow(`
            SELECT current_state, state_guideline, 
                   COALESCE(next_transition, ''), COALESCE(transition_description, '')
            FROM agent_next_actions WHERE agent_id = ?`, agentID).
            Scan(&state, &guideline, &transition, &description)
        
        if err != nil || transition == "" {
            break // Terminal state or error
        }
        
        // Step 3: Execute state work based on guideline
        if err := executeStateWork(state, guideline); err != nil {
            return err
        }
        
        // Step 4: Commit transition atomically
        _, err = db.Exec(`
            BEGIN IMMEDIATE;
            UPDATE agent_instance SET current_state_id = (
                SELECT to_state_id FROM workflow_transition 
                WHERE name = ? AND from_state_id = current_state_id
            ) WHERE agent_id = ?;
            COMMIT;`, transition, agentID)
        
        if err != nil {
            return err
        }
    }
    return nil
}
```

This system provides complete autonomy while ensuring predictable, auditable behavior through declarative workflow definitions.
