-- Drop indices
DROP INDEX IF EXISTS idx_resources_talk_id;
DROP INDEX IF EXISTS idx_attendance_user_id;
DROP INDEX IF EXISTS idx_attendance_talk_id;
DROP INDEX IF EXISTS idx_votes_user_id;
DROP INDEX IF EXISTS idx_votes_talk_id;
DROP INDEX IF EXISTS idx_talks_speaker_id;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS resources;
DROP TABLE IF EXISTS attendance;
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS talks;
DROP TABLE IF EXISTS users; 