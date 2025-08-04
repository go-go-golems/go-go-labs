Below is a compact, copy-paste-ready SQL schema that lets you:

1. **Register every workflow** (state machine, states, and transitions).
2. **Track each live LLM agent** and the state it is currently in.
3. **Expose an agent-friendly view** that returns, for every agent, the reflection / guideline for its current state **plus** every legal next transition (with descriptions).

---

## 1  Core reference tables

```sql
-- 1.1  One record per state-machine definition
CREATE TABLE workflow (
    workflow_id   INTEGER PRIMARY KEY,
    name          TEXT UNIQUE NOT NULL,
    created_at    TEXT DEFAULT CURRENT_TIMESTAMP
);

-- 1.2  The possible states inside a workflow
CREATE TABLE workflow_state (
    state_id      INTEGER PRIMARY KEY,
    workflow_id   INTEGER NOT NULL REFERENCES workflow(workflow_id)
                  ON DELETE CASCADE,
    name          TEXT NOT NULL,
    guideline     TEXT,                     -- “reflect” text shown to the agent
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
    description   TEXT,                     -- tool calls, guards, etc.
    UNIQUE (workflow_id, name, from_state_id)
);
```

---

## 2  Runtime table for active agents

```sql
-- One row per active LLM agent instance
CREATE TABLE agent_instance (
    agent_id      INTEGER PRIMARY KEY,
    workflow_id   INTEGER NOT NULL REFERENCES workflow(workflow_id)
                  ON DELETE CASCADE,
    agent_name    TEXT,                     -- optional human­-friendly tag
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
```

---

## 3  Agent-centric helper view

```sql
/* 
   For each agent, list:
     • its current state,
     • the guideline / reflection for that state,
     • every legal next transition leaving that state,
       with destination state and transition description.
*/
CREATE VIEW agent_next_actions AS
SELECT
    a.agent_id,
    a.agent_name,
    w.name                 AS workflow,
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
```

### How the view behaves

| agent\_id | current\_state  | state\_guideline         | next\_transition | next\_state | transition\_description |
| --------: | --------------- | ------------------------ | ---------------- | ----------- | ----------------------- |
|        17 | `AwaitResponse` | “Monitor inbox stream …” | `PositiveReply`  | `END`       | “Hand off to AE”        |
|        17 | `AwaitResponse` | “Monitor inbox stream …” | *(NULL)*         | *(NULL)*    | *(NULL)*                |

*If a state is terminal (no outgoing transitions), `next_transition` columns are `NULL`, signalling the agent to halt or escalate.*

---

## 4  Typical insertion snippet (Python-sqlite pseudo-code)

```python
cur.execute("INSERT INTO workflow (name) VALUES (?)", ("LeadQualAgent",))
wf_id = cur.lastrowid

cur.executemany(
    "INSERT INTO workflow_state (workflow_id, name, guideline) VALUES (?,?,?)",
    [
        (wf_id, "IngestLead",      "Summarize raw lead payload; highlight ICP features."),
        (wf_id, "ScoreAndReflect", "Calculate qualification score. Think step-by-step …"),
        ...
    ],
)

# Map name → id for convenience, then insert transitions the same way
```

---

### Why this schema works well with LLM agents

* **Declarativity:** You can swap workflows or tweak guidelines by `UPDATE` alone—no code changes.
* **Traceability:** Every transition is explicit and queryable for audits.
* **Agent UX:** One simple `SELECT * FROM agent_next_actions WHERE agent_id = ?` gives an agent all it needs to think and act.


