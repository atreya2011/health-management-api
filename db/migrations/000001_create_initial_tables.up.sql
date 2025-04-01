-- Enable UUID generation if not already enabled
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table (linked via Auth0 subject claim)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auth0_sub TEXT UNIQUE NOT NULL, -- Auth0 subject claim (e.g., "auth0|...")
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_users_auth0_sub ON users (auth0_sub);

-- Body composition records
CREATE TABLE body_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    date DATE NOT NULL, -- Date of the record
    weight_kg NUMERIC(5, 2), -- Weight in kilograms, e.g., 75.50
    body_fat_percentage NUMERIC(4, 2), -- Body fat percentage, e.g., 15.25
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (user_id, date) -- Allow only one record per user per day
);
CREATE INDEX idx_body_records_user_date ON body_records (user_id, date DESC);

-- Exercise records
CREATE TABLE exercise_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    exercise_name TEXT NOT NULL, -- e.g., "Running", "Weight Lifting"
    duration_minutes INTEGER, -- Duration in minutes
    calories_burned INTEGER, -- Estimated calories burned
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, -- When the exercise was performed/logged
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_exercise_records_user_recorded_at ON exercise_records (user_id, recorded_at DESC);

-- Diary entries
CREATE TABLE diary_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    title TEXT, -- Optional title
    content TEXT NOT NULL, -- The main diary text
    entry_date DATE NOT NULL, -- Date the diary entry pertains to
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_diary_entries_user_entry_date ON diary_entries (user_id, entry_date DESC);

-- Columns/Articles
CREATE TABLE columns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    category TEXT, -- e.g., "Diet", "Exercise", "Mental Health"
    tags TEXT[], -- Array of tags
    published_at TIMESTAMPTZ, -- Nullable, only show if not null and in the past
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_columns_published_at ON columns (published_at DESC NULLS LAST);
CREATE INDEX idx_columns_category ON columns (category);
CREATE INDEX idx_columns_tags ON columns USING GIN (tags); -- GIN index for array searching

-- Function to automatically update 'updated_at' timestamps
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply the trigger to relevant tables
CREATE TRIGGER set_timestamp_users BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_timestamp_body_records BEFORE UPDATE ON body_records FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_timestamp_exercise_records BEFORE UPDATE ON exercise_records FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_timestamp_diary_entries BEFORE UPDATE ON diary_entries FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_timestamp_columns BEFORE UPDATE ON columns FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
