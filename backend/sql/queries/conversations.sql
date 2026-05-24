-- name: CreateConversation :one
INSERT INTO conversations (type, name, admin_id)
VALUES ($1, $2, $3)
RETURNING id, type, name, admin_id, created_at, updated_at;

-- name: GetConversationByID :one
SELECT c.id, c.type, c.name, c.admin_id, c.created_at, c.updated_at,
       COALESCE(
         (SELECT json_agg(json_build_object(
           'user_id', cm.user_id,
           'username', u.username,
           'role', cm.role,
           'joined_at', cm.joined_at
         )) FROM conversation_members cm JOIN users u ON cm.user_id = u.id WHERE cm.conversation_id = c.id),
         '[]'::json
       ) AS members
FROM conversations c
WHERE c.id = $1;

-- name: ListUserConversations :many
SELECT c.id, c.type, c.name, c.admin_id, c.created_at, c.updated_at,
       COALESCE(
         (SELECT json_agg(json_build_object(
           'user_id', cm.user_id,
           'username', u.username,
           'role', cm.role,
           'joined_at', cm.joined_at
         )) FROM conversation_members cm JOIN users u ON cm.user_id = u.id WHERE cm.conversation_id = c.id),
         '[]'::json
       ) AS members,
       (SELECT m.content FROM messages m WHERE m.conversation_id = c.id ORDER BY m.created_at DESC LIMIT 1) AS last_message_content,
       (SELECT m.created_at FROM messages m WHERE m.conversation_id = c.id ORDER BY m.created_at DESC LIMIT 1) AS last_message_at
FROM conversations c
JOIN conversation_members cm ON cm.conversation_id = c.id
WHERE cm.user_id = $1
ORDER BY COALESCE(last_message_at, c.created_at) DESC;

-- name: AddConversationMember :one
INSERT INTO conversation_members (conversation_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (conversation_id, user_id) DO NOTHING
RETURNING id;

-- name: RemoveConversationMember :exec
DELETE FROM conversation_members
WHERE conversation_id = $1 AND user_id = $2;

-- name: GetConversationMembers :many
SELECT cm.user_id, u.username, cm.role, cm.joined_at
FROM conversation_members cm
JOIN users u ON cm.user_id = u.id
WHERE cm.conversation_id = $1;

-- name: IsConversationMember :one
SELECT EXISTS(
  SELECT 1 FROM conversation_members
  WHERE conversation_id = $1 AND user_id = $2
) AS is_member;

-- name: FindPrivateConversation :one
SELECT c.id, c.type, c.name, c.admin_id, c.created_at, c.updated_at
FROM conversations c
WHERE c.type = 'private'
  AND EXISTS (
    SELECT 1 FROM conversation_members cm1
    WHERE cm1.conversation_id = c.id AND cm1.user_id = $1
  )
  AND EXISTS (
    SELECT 1 FROM conversation_members cm2
    WHERE cm2.conversation_id = c.id AND cm2.user_id = $2
  )
LIMIT 1;

-- name: DeleteConversation :exec
DELETE FROM conversations WHERE id = $1;
