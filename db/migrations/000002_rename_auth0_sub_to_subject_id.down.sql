-- Revert the column name change
ALTER TABLE users RENAME COLUMN subject_id TO auth0_sub;

-- Revert the index name change
DROP INDEX idx_users_subject_id;
CREATE INDEX idx_users_auth0_sub ON users (auth0_sub);
