-- name: CreateExerciseRecord :one
INSERT INTO exercise_records (user_id, exercise_name, duration_minutes, calories_burned, recorded_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListExerciseRecordsByUser :many
SELECT * FROM exercise_records
WHERE user_id = $1
ORDER BY recorded_at DESC
LIMIT $2 OFFSET $3; -- For pagination

-- name: DeleteExerciseRecord :exec
DELETE FROM exercise_records
WHERE id = $1 AND user_id = $2;

-- name: CountExerciseRecordsByUser :one
SELECT COUNT(*) FROM exercise_records
WHERE user_id = $1;
