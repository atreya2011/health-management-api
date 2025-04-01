-- Rename auth0_sub column to subject_id in users table
ALTER TABLE users RENAME COLUMN auth0_sub TO subject_id;

-- Rename the index
DROP INDEX idx_users_auth0_sub;
CREATE INDEX idx_users_subject_id ON users (subject_id);
