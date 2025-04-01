-- name: CreateUser :one
INSERT INTO users (subject_id)
VALUES ($1)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserBySubjectID :one
SELECT * FROM users
WHERE subject_id = $1 LIMIT 1;
