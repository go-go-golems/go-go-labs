-- Enhanced SQLite Schema for State Machine Agents
-- Includes workflow descriptions and enhanced guidelines for scientific research

-- 1.1  One record per state-machine definition (enhanced with description)
CREATE TABLE workflow (
    workflow_id   INTEGER PRIMARY KEY,
    name          TEXT UNIQUE NOT NULL,
    description   TEXT,                         -- Enhanced: workflow description
    created_at    TEXT DEFAULT CURRENT_TIMESTAMP
);

-- 1.2  The possible states inside a workflow
CREATE TABLE workflow_state (
    state_id      INTEGER PRIMARY KEY,
    workflow_id   INTEGER NOT NULL REFERENCES workflow(workflow_id)
                  ON DELETE CASCADE,
    name          TEXT NOT NULL,
    guideline     TEXT,                     -- Enhanced reflection text for agents
    UNIQUE (workflow_id, name)
);

-- 1.3  Transitions (events / actions) between states
CREATE TABLE workflow_transition (
    transition_id INTEGER PRIMARY KEY,
    workflow_id   INTEGER NOT NULL REFERENCES workflow(workflow_id)
                  ON DELETE CASCADE,
    name          TEXT NOT NULL,            -- e.g. `SendIntro`
    from_state_id INTEGER NOT NULL REFERENCES workflow_state(state_id)
                  ON DELETE CASCADE,
    to_state_id   INTEGER NOT NULL REFERENCES workflow_state(state_id)
                  ON DELETE CASCADE,
    description   TEXT,                     -- Enhanced action descriptions
    UNIQUE (workflow_id, name, from_state_id)
);

-- 2. Runtime table for active agents
CREATE TABLE agent_instance (
    agent_id      INTEGER PRIMARY KEY,
    workflow_id   INTEGER NOT NULL REFERENCES workflow(workflow_id)
                  ON DELETE CASCADE,
    agent_name    TEXT,                     -- optional human-friendly tag
    current_state_id INTEGER NOT NULL REFERENCES workflow_state(state_id),
    updated_at    TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Keep `updated_at` fresh whenever state changes
CREATE TRIGGER agent_state_touch
AFTER UPDATE OF current_state_id ON agent_instance
BEGIN
    UPDATE agent_instance
    SET updated_at = CURRENT_TIMESTAMP
    WHERE agent_id = NEW.agent_id;
END;

-- 3. Agent-centric helper view
CREATE VIEW agent_next_actions AS
SELECT
    a.agent_id,
    a.agent_name,
    w.name                 AS workflow,
    w.description          AS workflow_description,
    cs.name                AS current_state,
    cs.guideline           AS state_guideline,
    t.name                 AS next_transition,
    ns.name                AS next_state,
    t.description          AS transition_description
FROM agent_instance           AS a
JOIN workflow_state           AS cs  ON cs.state_id = a.current_state_id
JOIN workflow                 AS w   ON w.workflow_id = a.workflow_id
LEFT JOIN workflow_transition AS t   ON t.from_state_id = cs.state_id
LEFT JOIN workflow_state      AS ns  ON ns.state_id = t.to_state_id
ORDER BY a.agent_id, cs.name, t.name;
