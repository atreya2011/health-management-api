-- Re-create the trigger function if it was dropped (it wasn't in the corresponding .up.sql, but included for completeness)
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Re-apply the triggers to relevant tables
-- Note: Using CREATE TRIGGER without IF NOT EXISTS will error if the trigger already exists,
-- which is the desired behavior when reverting the 'up' migration.
CREATE TRIGGER set_timestamp_users BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_timestamp_body_records BEFORE UPDATE ON body_records FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_timestamp_exercise_records BEFORE UPDATE ON exercise_records FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_timestamp_diary_entries BEFORE UPDATE ON diary_entries FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_timestamp_columns BEFORE UPDATE ON columns FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
