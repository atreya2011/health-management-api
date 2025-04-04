-- name: ListPublishedColumns :many
SELECT * FROM columns
WHERE published_at IS NOT NULL AND published_at <= CURRENT_TIMESTAMP
ORDER BY published_at DESC
LIMIT $1 OFFSET $2; -- For pagination

-- name: GetColumnByID :one
SELECT * FROM columns
WHERE id = $1 AND (published_at IS NOT NULL AND published_at <= CURRENT_TIMESTAMP)
LIMIT 1;

-- name: ListColumnsByCategory :many
SELECT * FROM columns
WHERE category = $1 AND published_at IS NOT NULL AND published_at <= CURRENT_TIMESTAMP
ORDER BY published_at DESC
LIMIT $2 OFFSET $3; -- For pagination

-- name: ListColumnsByTag :many
SELECT * FROM columns
WHERE $1::text = ANY(tags) AND published_at IS NOT NULL AND published_at <= CURRENT_TIMESTAMP
ORDER BY published_at DESC
LIMIT $2 OFFSET $3; -- For pagination

-- name: CountPublishedColumns :one
SELECT COUNT(*) FROM columns
WHERE published_at IS NOT NULL AND published_at <= CURRENT_TIMESTAMP;

-- name: CountColumnsByCategory :one
SELECT COUNT(*) FROM columns
WHERE category = $1 AND published_at IS NOT NULL AND published_at <= CURRENT_TIMESTAMP;

-- name: CountColumnsByTag :one
SELECT COUNT(*) FROM columns
WHERE $1::text = ANY(tags) AND published_at IS NOT NULL AND published_at <= CURRENT_TIMESTAMP;
