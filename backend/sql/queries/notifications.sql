-- name: CreateNotification :one
INSERT INTO notifications (user_id, type, reference_id, content, is_read)
VALUES ($1, $2, $3, $4, FALSE)
RETURNING id, user_id, type, reference_id, content, is_read, created_at;

-- name: ListNotifications :many
SELECT id, user_id, type, reference_id, content, is_read, created_at
FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: GetUnreadNotificationCount :one
SELECT COUNT(*) AS count
FROM notifications
WHERE user_id = $1 AND is_read = FALSE;

-- name: MarkNotificationRead :exec
UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications SET is_read = TRUE WHERE user_id = $1;

-- name: DeleteNotification :exec
DELETE FROM notifications WHERE id = $1 AND user_id = $2;
