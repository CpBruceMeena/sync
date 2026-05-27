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
