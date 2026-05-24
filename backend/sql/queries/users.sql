-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, display_name)
VALUES ($1, $2, $3, $4)
RETURNING id, username, email, display_name, avatar_url, status, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, username, email, display_name, avatar_url, status, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, email, display_name, avatar_url, status, created_at, updated_at
FROM users
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT id, username, email, display_name, avatar_url, status, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByEmailWithPassword :one
SELECT id, username, email, password_hash, display_name, avatar_url, status, created_at, updated_at
FROM users
WHERE email = $1;

-- name: ListUsers :many
SELECT id, username, email, display_name, avatar_url, status, created_at, updated_at
FROM users
ORDER BY username ASC;

-- name: UpdateUser :one
UPDATE users
SET display_name = COALESCE($2, display_name),
    avatar_url = COALESCE($3, avatar_url),
    status = COALESCE($4, status),
    updated_at = NOW()
WHERE id = $1
RETURNING id, username, email, display_name, avatar_url, status, created_at, updated_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserStatus :exec
UPDATE users
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
