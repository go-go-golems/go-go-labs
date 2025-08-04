-- Add audit logging to track all agent state transitions

-- Agent state transition log table
CREATE TABLE agent_state_log (
    log_id        INTEGER PRIMARY KEY,
    agent_id      INTEGER NOT NULL REFERENCES agent_instance(agent_id)
                  ON DELETE CASCADE,
    workflow_id   INTEGER NOT NULL REFERENCES workflow(workflow_id)
                  ON DELETE CASCADE,
    from_state_id INTEGER REFERENCES workflow_state(state_id),  -- NULL for initial state
    to_state_id   INTEGER NOT NULL REFERENCES workflow_state(state_id),
    transition_name TEXT,                                       -- NULL for initial state
    logged_at     TEXT DEFAULT CURRENT_TIMESTAMP,
    session_info  TEXT                                          -- Optional context/metadata
);

-- Index for efficient querying by agent and time
CREATE INDEX idx_agent_state_log_agent_time ON agent_state_log(agent_id, logged_at);
CREATE INDEX idx_agent_state_log_workflow ON agent_state_log(workflow_id, logged_at);

-- Enhanced trigger to log state changes with transition info
DROP TRIGGER IF EXISTS agent_state_touch;

CREATE TRIGGER agent_state_change_logger
AFTER UPDATE OF current_state_id ON agent_instance
FOR EACH ROW
WHEN OLD.current_state_id != NEW.current_state_id
BEGIN
    -- Update the timestamp
    UPDATE agent_instance
    SET updated_at = CURRENT_TIMESTAMP
    WHERE agent_id = NEW.agent_id;
    
    -- Log the state transition
    INSERT INTO agent_state_log (
        agent_id, 
        workflow_id, 
        from_state_id, 
        to_state_id, 
        transition_name,
        session_info
    ) VALUES (
        NEW.agent_id,
        NEW.workflow_id,
        OLD.current_state_id,
        NEW.current_state_id,
        -- Try to determine the transition name from the state change
        (SELECT name FROM workflow_transition 
         WHERE workflow_id = NEW.workflow_id 
           AND from_state_id = OLD.current_state_id 
           AND to_state_id = NEW.current_state_id
         LIMIT 1),
        'Auto-logged state transition'
    );
END;

-- Trigger for initial agent creation (first state entry)
CREATE TRIGGER agent_creation_logger
AFTER INSERT ON agent_instance
FOR EACH ROW
BEGIN
    INSERT INTO agent_state_log (
        agent_id, 
        workflow_id, 
        from_state_id, 
        to_state_id, 
        transition_name,
        session_info
    ) VALUES (
        NEW.agent_id,
        NEW.workflow_id,
        NULL,  -- No previous state
        NEW.current_state_id,
        'AGENT_CREATED',
        'Agent instance created: ' || COALESCE(NEW.agent_name, 'unnamed')
    );
END;

-- View for easy log analysis with human-readable state names
CREATE VIEW agent_transition_history AS
SELECT 
    l.log_id,
    l.agent_id,
    ai.agent_name,
    w.name as workflow_name,
    fs.name as from_state,
    ts.name as to_state,
    l.transition_name,
    l.logged_at,
    l.session_info
FROM agent_state_log l
JOIN agent_instance ai ON ai.agent_id = l.agent_id
JOIN workflow w ON w.workflow_id = l.workflow_id
LEFT JOIN workflow_state fs ON fs.state_id = l.from_state_id
JOIN workflow_state ts ON ts.state_id = l.to_state_id
ORDER BY l.agent_id, l.logged_at;
