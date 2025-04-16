-- Drop the automatic updated_at triggers to allow application control

DROP TRIGGER IF EXISTS set_timestamp_users ON users;
DROP TRIGGER IF EXISTS set_timestamp_body_records ON body_records;
DROP TRIGGER IF EXISTS set_timestamp_exercise_records ON exercise_records;
DROP TRIGGER IF EXISTS set_timestamp_diary_entries ON diary_entries;
DROP TRIGGER IF EXISTS set_timestamp_columns ON columns;

-- Optionally, drop the function if it's no longer needed by any other triggers
-- DROP FUNCTION IF EXISTS trigger_set_timestamp();
-- Decided against dropping the function for now, in case it's used elsewhere or needed later.
