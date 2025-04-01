-- name: CreateDiaryEntry :one
INSERT INTO diary_entries (user_id, title, content, entry_date)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateDiaryEntry :one
UPDATE diary_entries
SET title = $2, content = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $4
RETURNING *;

-- name: ListDiaryEntriesByUser :many
SELECT * FROM diary_entries
WHERE user_id = $1
ORDER BY entry_date DESC
LIMIT $2 OFFSET $3; -- For pagination

-- name: GetDiaryEntryByID :one
SELECT * FROM diary_entries
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: DeleteDiaryEntry :exec
DELETE FROM diary_entries
WHERE id = $1 AND user_id = $2;

-- name: CountDiaryEntriesByUser :one
SELECT COUNT(*) FROM diary_entries
WHERE user_id = $1;
