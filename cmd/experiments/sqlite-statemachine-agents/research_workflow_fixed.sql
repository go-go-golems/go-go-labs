-- Enhanced Research Paper Scout Workflow
-- More detailed descriptions and guidelines for scientific research agents

BEGIN TRANSACTION;

-- Insert the Research Paper Scout workflow with description
INSERT INTO workflow (name, description) VALUES (
    'ResearchPaperScoutAgent',
    'Automated scientific literature discovery agent that performs daily searches across academic databases, extracts paper metadata, deduplicates against existing archives, and presents scored summaries for researcher review and curation.'
);

-- Insert states with enhanced guidelines for scientific research
INSERT INTO workflow_state (workflow_id, name, guideline)
SELECT 
    (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'),
    s.name,
    s.guideline
FROM (VALUES
    ('Start',              
     'Initialize daily research scanning session. Verify system connectivity, load configuration parameters, and prepare for automated literature discovery. This is the entry point where you should establish logging and ensure all required resources are available for the research pipeline.'),
    
    ('LoadQueries',        
     'Load and validate saved research queries from configuration. Check for query syntax correctness, verify search terms are current and relevant to ongoing research interests. Evaluate if the query set covers all necessary research domains and sub-fields. Branch execution based on query availability - proceed if queries exist, otherwise guide user to configure search terms.'),
    
    ('SearchWeb',          
     'Execute systematic literature searches across multiple academic databases including arXiv, Semantic Scholar, and other relevant repositories. For each query term, perform site-restricted searches with appropriate date filters, pagination handling, and rate limiting. Collect comprehensive result sets while respecting API limits and terms of service.'),
    
    ('ExtractMetadata',    
     'Parse and extract structured metadata from search results including paper identifiers (DOI, arXiv ID), titles, author lists, publication venues, abstracts, publication dates, and direct PDF/HTML links. Validate extracted data for completeness and accuracy, handle various citation formats, and normalize metadata across different source formats.'),
    
    ('Deduplicate',        
     'Compare newly discovered papers against existing research archive using multiple matching criteria: DOI exact match, arXiv ID comparison, title similarity analysis, and author-date combinations. Implement fuzzy matching for titles to catch minor variations. Filter out previously catalogued papers to focus only on genuinely new discoveries.'),
    
    ('SummariseAndScore',  
     'Generate concise technical summaries (TL;DR) and calculate relevance scores (0-100 scale) for each new paper. Analyze abstracts for methodology, key findings, and potential impact. Score based on alignment with research interests, novelty of approach, citation potential, and methodological rigor. Consider interdisciplinary connections and practical applications.'),
    
    ('AskUser',            
     'Present curated list of new papers with summaries and relevance scores to researcher for review. Display papers in ranked order with key metadata, scores, and generated summaries. Provide options to Accept (archive), Skip (ignore), or Quit (terminate session). Support batch operations and allow detailed paper inspection before decision-making.'),
    
    ('SaveAccepted',       
     'Archive user-approved papers into the research database with comprehensive metadata, timestamps, and categorization tags. Ensure data integrity, create backup records, and update search indices. Log acceptance decisions and maintain audit trail for research workflow tracking and future query refinement.'),
    
    ('Report',             
     'Generate comprehensive session summary including total papers discovered, new additions to archive, top-scored papers by relevance, search query performance statistics, and recommendations for query optimization. Format report for easy review and integration into research documentation workflows.'),
    
    ('End',                
     'Terminate research session gracefully. Clean up temporary resources, close database connections, save session logs, and prepare system for next scheduled run. Ensure all data is properly persisted and system is ready for subsequent automated or manual research discovery sessions.')
) AS s(name, guideline);

-- Insert transitions with enhanced action descriptions for scientific research
INSERT INTO workflow_transition (workflow_id, name, from_state_id, to_state_id, description)
SELECT 
    (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent'),
    t.event,
    (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = t.from_state),
    (SELECT state_id FROM workflow_state WHERE workflow_id = (SELECT workflow_id FROM workflow WHERE name = 'ResearchPaperScoutAgent') AND name = t.to_state),
    t.action
FROM (VALUES
   ('DailyCron',      'Start',            'LoadQueries',
    'Trigger automated daily research discovery session. Initialize logging, verify system health, check network connectivity to academic databases, and prepare environment for literature search pipeline execution.'),

   ('QueriesLoaded',  'LoadQueries',      'SearchWeb',
    'Successfully loaded and validated research queries from configuration. Verified query syntax, confirmed search terms are current and comprehensive. Proceed to execute systematic searches across academic databases with loaded query set.'),

   ('NoQueries',      'LoadQueries',      'End',
    'No saved research queries found in configuration. Unable to proceed with automated literature discovery. Alert user to configure search terms covering relevant research domains and sub-fields before next scheduled run.'),

   ('SearchesDone',   'SearchWeb',        'ExtractMetadata',
    'Completed systematic literature searches across all configured academic databases. Successfully executed site-restricted searches for each query term with proper pagination, rate limiting, and result collection. Gathered comprehensive result sets ready for metadata extraction.'),

   ('MetadataReady',  'ExtractMetadata',  'Deduplicate',
    'Successfully extracted and normalized metadata from all search results. Parsed paper identifiers, titles, author information, abstracts, publication details, and access links. Validated data completeness and converted various citation formats into standardized structure.'),

   ('AllKnown',       'Deduplicate',      'Report',
    'All discovered papers already exist in research archive. No new literature found in current search session. Skip scoring and user review phases, proceed directly to generate session summary with statistics on search coverage and archive completeness.'),

   ('NewCandidates',  'Deduplicate',      'SummariseAndScore',
    'Identified new papers not present in existing research archive through multi-criteria matching analysis. Successfully filtered out duplicates using DOI comparison, title similarity, and author-date combinations. Proceed to analyze and score newly discovered literature.'),

   ('ScoresReady',    'SummariseAndScore','AskUser',
    'Generated technical summaries and calculated relevance scores for all new papers. Analyzed abstracts for methodology, findings, and potential impact. Computed 0-100 relevance scores based on research alignment, novelty, and methodological rigor. Present ranked results for researcher review.'),

   ('UserAccept',     'AskUser',          'SaveAccepted',
    'Researcher approved specific papers for archival after reviewing summaries and scores. User selected subset of discovered literature for permanent inclusion in research database. Proceed to archive approved papers with full metadata and categorization.'),

   ('UserSkipAll',    'AskUser',          'Report',
    'Researcher chose to skip entire batch of discovered papers. No new literature selected for archival. User determined current results do not meet research criteria or relevance threshold. Proceed to session summary with rejection statistics.'),

   ('UserQuit',       'AskUser',          'Report',
    'Researcher terminated review session early before completing evaluation of all discovered papers. Partial user decisions recorded. Proceed to generate session summary with status of reviewed versus pending papers for next session.'),

   ('Saved',          'SaveAccepted',     'Report',
    'Successfully archived all user-approved papers into research database with comprehensive metadata, timestamps, and categorization tags. Updated search indices, maintained data integrity, and created audit trail of acceptance decisions for future query optimization.'),

   ('Reported',       'Report',           'End',
    'Generated and delivered comprehensive session summary including discovery statistics, archive additions, top-scored papers, and search performance metrics. Provided recommendations for query refinement and research workflow optimization. Session documentation complete.')
) AS t(event, from_state, to_state, action);

COMMIT;
