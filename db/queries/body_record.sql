-- name: CreateBodyRecord :one
INSERT INTO body_records (user_id, date, weight_kg, body_fat_percentage, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, date) DO UPDATE SET
    weight_kg = EXCLUDED.weight_kg,
    body_fat_percentage = EXCLUDED.body_fat_percentage,
    updated_at = $6
RETURNING *;

-- name: ListBodyRecordsByUser :many
SELECT * FROM body_records
WHERE user_id = $1
ORDER BY date DESC
LIMIT $2 OFFSET $3; -- For pagination

-- name: ListBodyRecordsByUserDateRange :many
SELECT * FROM body_records
WHERE user_id = $1 AND date >= $2 AND date <= $3
ORDER BY date ASC;

-- name: CountBodyRecordsByUser :one
SELECT COUNT(*) FROM body_records
WHERE user_id = $1;
