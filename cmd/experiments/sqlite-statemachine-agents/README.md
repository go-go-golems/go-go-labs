# SQLite State Machine Agents

A system for creating autonomous agents that follow predefined workflows stored in SQLite with automatic audit logging.

## Quick Start

1. **Create the database:**
   ```bash
   sqlite3 research_agents.db < sql/init_database.sql
   ```

2. **Read the agent guidelines:**
   - [Agent Guidelines](docs/AGENT_GUIDELINES.md) - Complete guide for agents

3. **Start using:**
   - Database includes pre-loaded ResearchPaperScoutAgent workflow
   - All state transitions are automatically logged
   - Use `agent_next_actions` view for agent operations

## Directory Structure

```
├── README.md                    # This file
├── research_agents.db          # SQLite database (created)
├── sql/
│   ├── init_database.sql       # Complete database setup with logging
│   └── add_logging.sql         # Add logging to existing database
└── docs/
    ├── AGENT_GUIDELINES.md     # Complete agent operating manual
    ├── sqlite-schema.md        # Original schema documentation
    └── searcher-sqlite.md      # Research workflow example
```

## Features

- **Declarative Workflows**: Define state machines in database tables
- **Automatic Logging**: All state transitions tracked with timestamps
- **Agent-Friendly Views**: Single query shows current state + available actions
- **Audit Trail**: Complete history of agent behavior for analysis
- **Concurrent Safe**: Atomic transitions with proper locking

## Example Usage

```sql
-- Check your status and options
SELECT * FROM agent_next_actions WHERE agent_id = 1;

-- Execute state transition
BEGIN IMMEDIATE;
UPDATE agent_instance SET current_state_id = (
    SELECT to_state_id FROM workflow_transition 
    WHERE name = 'DailyCron' AND from_state_id = current_state_id
) WHERE agent_id = 1;
COMMIT;

-- View your history
SELECT from_state, to_state, transition_name, logged_at 
FROM agent_transition_history WHERE agent_id = 1;
```

## Workflows Included

- **ResearchPaperScoutAgent**: Automated scientific literature discovery workflow
  - 10 states: Start → LoadQueries → SearchWeb → ExtractMetadata → Deduplicate → SummariseAndScore → AskUser → SaveAccepted → Report → End
  - 13 transitions covering all workflow paths
  - Enhanced guidelines for scientific research

See [docs/AGENT_GUIDELINES.md](docs/AGENT_GUIDELINES.md) for complete details.
