Below is a **copy-paste SQL script** that loads the *Research Paper Scout* workflow into the schema we defined earlier.
All actions that had been in `on_enter` blocks are now attached to the *transition* that actually performs them.

> *Tip – if you already have a `workflow` table, keep the `name` unique.
> Every `SELECT … WHERE name = 'ResearchPaperScoutAgent'` picks up the correct ID, so you can run the whole script in one go.*

```sql
BEGIN TRANSACTION;

/* ──────────────────────────────────────────────────────────────
   1.  WORKFLOW HEADER
   ────────────────────────────────────────────────────────────── */
INSERT INTO workflow (name) VALUES ('ResearchPaperScoutAgent');

/* Convenience alias */
WITH wf AS (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent')

/* ──────────────────────────────────────────────────────────────
   2.  STATES
   ────────────────────────────────────────────────────────────── */
INSERT INTO workflow_state (workflow_id, name, guideline)
SELECT wf.workflow_id,
       s.name,
       s.guideline
FROM wf,
     (VALUES
        ('Start',              'Entry point for the daily run.'),
        ('LoadQueries',        'Load saved search terms and branch on emptiness.'),
        ('SearchWeb',          'Web-search for each term on arXiv & Semantic Scholar.'),
        ('ExtractMetadata',    'Pull identifiers, titles, URLs, abstracts from hits.'),
        ('Deduplicate',        'Discard candidates already stored in the archive.'),
        ('SummariseAndScore',  'Generate TL;DR and relevance score for each new paper.'),
        ('AskUser',            'Show list; user chooses Accept / Skip / Quit.'),
        ('SaveAccepted',       'Archive user-approved papers with timestamp.'),
        ('Report',             'Brief session summary (count, top 5).'),
        ('End',                'Terminal state.')
     ) AS s(name, guideline);


/* ──────────────────────────────────────────────────────────────
   3.  TRANSITIONS  (actions now live here)
   ────────────────────────────────────────────────────────────── */
INSERT INTO workflow_transition (workflow_id, name, from_state_id, to_state_id, description)
SELECT wf.workflow_id,
       t.event,
       (SELECT state_id FROM workflow_state WHERE workflow_id = wf.workflow_id AND name = t.from_state),
       (SELECT state_id FROM workflow_state WHERE workflow_id = wf.workflow_id AND name = t.to_state),
       t.action
FROM wf,
     (VALUES
       /*  ⬇  DAILY START  */
       ('DailyCron',      'Start',            'LoadQueries',
        'Trigger on the daily schedule.'),

       /*  ⬇  LOAD QUERIES BRANCH  */
       ('QueriesLoaded',  'LoadQueries',      'SearchWeb',
        'List of queries found – proceed to web search.'),

       ('NoQueries',      'LoadQueries',      'End',
        'No saved queries.  Tell the user to add some and stop.'),

       /*  ⬇  SEARCH & PARSE  */
       ('SearchesDone',   'SearchWeb',        'ExtractMetadata',
        'For each query, perform site-restricted searches and collect links.'),

       ('MetadataReady',  'ExtractMetadata',  'Deduplicate',
        'Parse identifier / title / URL / abstract from every hit.'),

       /*  ⬇  DEDUPE BRANCH  */
       ('AllKnown',       'Deduplicate',      'Report',
        'Every candidate already exists – skip straight to report.'),

       ('NewCandidates',  'Deduplicate',      'SummariseAndScore',
        'Keep only papers not present in the archive.'),

       /*  ⬇  SCORING TO USER DIALOG  */
       ('ScoresReady',    'SummariseAndScore','AskUser',
        'Generate TL;DR & 0-100 relevance score for each new paper.'),

       ('UserAccept',     'AskUser',          'SaveAccepted',
        'User selected specific items to archive.'),

       ('UserSkipAll',    'AskUser',          'Report',
        'User skipped the whole batch; nothing to save.'),

       ('UserQuit',       'AskUser',          'Report',
        'User quit early; finish the run.'),

       /*  ⬇  SAVE & WRAP-UP  */
       ('Saved',          'SaveAccepted',     'Report',
        'Store accepted papers in the archive with timestamp.'),

       ('Reported',       'Report',           'End',
        'Send / display session summary and terminate.')
     ) AS t(event, from_state, to_state, action);

COMMIT;
```

### What moved where?

| Transition                    | **Action now performed during the transition**                 |
| ----------------------------- | -------------------------------------------------------------- |
| `QueriesLoaded` / `NoQueries` | Reading the saved query list and deciding if it’s empty.       |
| `SearchesDone`                | Running the two web searches per query and collecting links.   |
| `MetadataReady`               | Extracting identifiers, titles, URLs, abstracts from each hit. |
| `NewCandidates` / `AllKnown`  | Comparing new IDs/URLs against the archive.                    |
| `ScoresReady`                 | Summarising each paper and assigning a relevance score.        |
| `UserAccept`                  | Writing the user-approved papers into the archive.             |
| `Reported`                    | Producing and delivering the run summary.                      |

Everything is now declarative inside the SQLite catalog: **states carry brief guidelines**, and **transitions carry the real work** to execute when an event fires.

