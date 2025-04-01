DROP TRIGGER IF EXISTS set_timestamp_columns ON columns;
DROP TRIGGER IF EXISTS set_timestamp_diary_entries ON diary_entries;
DROP TRIGGER IF EXISTS set_timestamp_exercise_records ON exercise_records;
DROP TRIGGER IF EXISTS set_timestamp_body_records ON body_records;
DROP TRIGGER IF EXISTS set_timestamp_users ON users;

DROP FUNCTION IF EXISTS trigger_set_timestamp();

DROP TABLE IF EXISTS columns;
DROP TABLE IF EXISTS diary_entries;
DROP TABLE IF EXISTS exercise_records;
DROP TABLE IF EXISTS body_records;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "pgcrypto";
