-- Complete database initialization with schema and Research Paper Scout workflow

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

-- Insert the Research Paper Scout workflow
INSERT INTO workflow (name, description) VALUES (
    'ResearchPaperScoutAgent',
    'Automated scientific literature discovery agent that performs daily searches across academic databases, extracts paper metadata, deduplicates against existing archives, and presents scored summaries for researcher review and curation.'
);

-- Insert workflow states
INSERT INTO workflow_state (workflow_id, name, guideline) VALUES
((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'Start', 'Initialize daily research scanning session. Verify system connectivity, load configuration parameters, and prepare for automated literature discovery. This is the entry point where you should establish logging and ensure all required resources are available for the research pipeline.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'LoadQueries', 'Load and validate saved research queries from configuration. Check for query syntax correctness, verify search terms are current and relevant to ongoing research interests. Evaluate if the query set covers all necessary research domains and sub-fields. Branch execution based on query availability - proceed if queries exist, otherwise guide user to configure search terms.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'SearchWeb', 'Execute systematic literature searches across multiple academic databases including arXiv, Semantic Scholar, and other relevant repositories. For each query term, perform site-restricted searches with appropriate date filters, pagination handling, and rate limiting. Collect comprehensive result sets while respecting API limits and terms of service.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'ExtractMetadata', 'Parse and extract structured metadata from search results including paper identifiers (DOI, arXiv ID), titles, author lists, publication venues, abstracts, publication dates, and direct PDF/HTML links. Validate extracted data for completeness and accuracy, handle various citation formats, and normalize metadata across different source formats.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'Deduplicate', 'Compare newly discovered papers against existing research archive using multiple matching criteria: DOI exact match, arXiv ID comparison, title similarity analysis, and author-date combinations. Implement fuzzy matching for titles to catch minor variations. Filter out previously catalogued papers to focus only on genuinely new discoveries.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'SummariseAndScore', 'Generate concise technical summaries (TL;DR) and calculate relevance scores (0-100 scale) for each new paper. Analyze abstracts for methodology, key findings, and potential impact. Score based on alignment with research interests, novelty of approach, citation potential, and methodological rigor. Consider interdisciplinary connections and practical applications.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'AskUser', 'Present curated list of new papers with summaries and relevance scores to researcher for review. Display papers in ranked order with key metadata, scores, and generated summaries. Provide options to Accept (archive), Skip (ignore), or Quit (terminate session). Support batch operations and allow detailed paper inspection before decision-making.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'SaveAccepted', 'Archive user-approved papers into the research database with comprehensive metadata, timestamps, and categorization tags. Ensure data integrity, create backup records, and update search indices. Log acceptance decisions and maintain audit trail for research workflow tracking and future query refinement.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'Report', 'Generate comprehensive session summary including total papers discovered, new additions to archive, top-scored papers by relevance, search query performance statistics, and recommendations for query optimization. Format report for easy review and integration into research documentation workflows.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'End', 'Terminate research session gracefully. Clean up temporary resources, close database connections, save session logs, and prepare system for next scheduled run. Ensure all data is properly persisted and system is ready for subsequent automated or manual research discovery sessions.');

-- Insert workflow transitions
INSERT INTO workflow_transition (workflow_id, name, from_state_id, to_state_id, description) VALUES
((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'DailyCron', 
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'Start'),
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'LoadQueries'),
 'Trigger automated daily research discovery session. Initialize logging, verify system health, check network connectivity to academic databases, and prepare environment for literature search pipeline execution.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'QueriesLoaded', 
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'LoadQueries'),
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'SearchWeb'),
 'Successfully loaded and validated research queries from configuration. Verified query syntax, confirmed search terms are current and comprehensive. Proceed to execute systematic searches across academic databases with loaded query set.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'NoQueries', 
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'LoadQueries'),
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'End'),
 'No saved research queries found in configuration. Unable to proceed with automated literature discovery. Alert user to configure search terms covering relevant research domains and sub-fields before next scheduled run.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'SearchesDone', 
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'SearchWeb'),
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'ExtractMetadata'),
 'Completed systematic literature searches across all configured academic databases. Successfully executed site-restricted searches for each query term with proper pagination, rate limiting, and result collection. Gathered comprehensive result sets ready for metadata extraction.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'MetadataReady', 
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'ExtractMetadata'),
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'Deduplicate'),
 'Successfully extracted and normalized metadata from all search results. Parsed paper identifiers, titles, author information, abstracts, publication details, and access links. Validated data completeness and converted various citation formats into standardized structure.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'AllKnown', 
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'Deduplicate'),
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'Report'),
 'All discovered papers already exist in research archive. No new literature found in current search session. Skip scoring and user review phases, proceed directly to generate session summary with statistics on search coverage and archive completeness.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'NewCandidates', 
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'Deduplicate'),
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'SummariseAndScore'),
 'Identified new papers not present in existing research archive through multi-criteria matching analysis. Successfully filtered out duplicates using DOI comparison, title similarity, and author-date combinations. Proceed to analyze and score newly discovered literature.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'ScoresReady', 
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'SummariseAndScore'),
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'AskUser'),
 'Generated technical summaries and calculated relevance scores for all new papers. Analyzed abstracts for methodology, findings, and potential impact. Computed 0-100 relevance scores based on research alignment, novelty, and methodological rigor. Present ranked results for researcher review.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'UserAccept', 
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'AskUser'),
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'SaveAccepted'),
 'Researcher approved specific papers for archival after reviewing summaries and scores. User selected subset of discovered literature for permanent inclusion in research database. Proceed to archive approved papers with full metadata and categorization.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'UserSkipAll', 
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'AskUser'),
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'Report'),
 'Researcher chose to skip entire batch of discovered papers. No new literature selected for archival. User determined current results do not meet research criteria or relevance threshold. Proceed to session summary with rejection statistics.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'UserQuit', 
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'AskUser'),
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'Report'),
 'Researcher terminated review session early before completing evaluation of all discovered papers. Partial user decisions recorded. Proceed to generate session summary with status of reviewed versus pending papers for next session.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'Saved', 
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'SaveAccepted'),
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'Report'),
 'Successfully archived all user-approved papers into research database with comprehensive metadata, timestamps, and categorization tags. Updated search indices, maintained data integrity, and created audit trail of acceptance decisions for future query optimization.'),

((SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'), 'Reported', 
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'Report'),
 (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = 'End'),
 'Generated and delivered comprehensive session summary including discovery statistics, archive additions, top-scored papers, and search performance metrics. Provided recommendations for query refinement and research workflow optimization. Session documentation complete.');
