package repository

import (
	"context"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type messageReadRepository struct {
	db *gorm.DB
}

func NewMessageReadRepository(db *gorm.DB) MessageReadRepository {
	return &messageReadRepository{db: db}
}

func (r *messageReadRepository) Upsert(ctx context.Context, convID, userID uuid.UUID) error {
	// Upsert: insert or update last_read_at
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO message_reads (conversation_id, user_id, last_read_at)
		VALUES (?, ?, NOW())
		ON CONFLICT (conversation_id, user_id)
		DO UPDATE SET last_read_at = NOW()
	`, convID, userID).Error
}

func (r *messageReadRepository) GetByConversation(ctx context.Context, convID uuid.UUID) ([]models.MessageRead, error) {
	var reads []models.MessageRead
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", convID).
		Find(&reads).Error
	return reads, err
}

// GetUnreadCounts returns a map of conversation_id -> unread message count for a user.
// Unread count is the number of messages created after the user's last_read_at.
func (r *messageReadRepository) GetUnreadCounts(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int64, error) {
	type result struct {
		ConversationID uuid.UUID
		Count          int64
	}
	var rows []result
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			cm.conversation_id,
			COUNT(m.id) AS count
		FROM conversation_members cm
		LEFT JOIN messages m ON m.conversation_id = cm.conversation_id
		LEFT JOIN message_reads mr
			ON mr.conversation_id = cm.conversation_id
			AND mr.user_id = ?
		WHERE cm.user_id = ?
		AND (mr.last_read_at IS NULL OR m.created_at > mr.last_read_at)
		GROUP BY cm.conversation_id
	`, userID, userID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	counts := make(map[uuid.UUID]int64, len(rows))
	for _, r := range rows {
		counts[r.ConversationID] = r.Count
	}
	return counts, nil
}
