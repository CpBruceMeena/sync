-- name: CreateSession :one
INSERT INTO sessions (user_id, refresh_token, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, refresh_token, expires_at, created_at;

-- name: GetSessionByToken :one
SELECT s.id, s.user_id, u.username, u.email, s.refresh_token, s.expires_at, s.created_at
FROM sessions s
JOIN users u ON s.user_id = u.id
WHERE s.refresh_token = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteUserSessions :exec
DELETE FROM sessions WHERE user_id = $1;

-- name: CleanExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < NOW();
