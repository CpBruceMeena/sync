-- name: CreateMessage :one
INSERT INTO messages (conversation_id, sender_id, content, type)
VALUES ($1, $2, $3, $4)
RETURNING id, conversation_id, sender_id, content, type, created_at;

-- name: GetMessageByID :one
SELECT m.id, m.conversation_id, m.sender_id, u.username AS sender_username, m.content, m.type, m.created_at
FROM messages m
JOIN users u ON m.sender_id = u.id
WHERE m.id = $1;

-- name: ListMessagesByConversation :many
SELECT m.id, m.conversation_id, m.sender_id, u.username AS sender_username, m.content, m.type, m.created_at,
       COALESCE(
         (SELECT json_agg(json_build_object(
           'user_id', r.user_id,
           'username', ru.username,
           'emoji', r.emoji,
           'created_at', r.created_at
         )) FROM reactions r JOIN users ru ON r.user_id = ru.id WHERE r.message_id = m.id),
         '[]'::json
       ) AS reactions
FROM messages m
JOIN users u ON m.sender_id = u.id
WHERE m.conversation_id = $1
  AND ($2::uuid IS NULL OR m.id < $2::uuid)
ORDER BY m.created_at DESC
LIMIT $3;

-- name: DeleteMessage :exec
DELETE FROM messages WHERE id = $1 AND sender_id = $2;

-- name: AddReaction :one
INSERT INTO reactions (message_id, user_id, emoji)
VALUES ($1, $2, $3)
ON CONFLICT (message_id, user_id, emoji) DO NOTHING
RETURNING id;

-- name: RemoveReaction :exec
DELETE FROM reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3;
