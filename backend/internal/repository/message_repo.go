package repository

import (
	"context"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, msg *models.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *messageRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	var msg models.Message
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *messageRepository) ListByConversation(ctx context.Context, convID uuid.UUID, cursor uuid.UUID, limit int) ([]models.Message, error) {
	var msgs []models.Message
	query := r.db.WithContext(ctx).Where("conversation_id = ?", convID)
	if cursor != uuid.Nil {
		query = query.Where("id < ?", cursor)
	}
	err := query.Order("created_at DESC").Limit(limit).Find(&msgs).Error
	return msgs, err
}

func (r *messageRepository) Delete(ctx context.Context, id, senderID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND sender_id = ?", id, senderID).
		Delete(&models.Message{}).Error
}

func (r *messageRepository) AddReaction(ctx context.Context, reaction *models.Reaction) error {
	return r.db.WithContext(ctx).Create(reaction).Error
}

func (r *messageRepository) RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	return r.db.WithContext(ctx).
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		Delete(&models.Reaction{}).Error
}

func (r *messageRepository) GetReactionsByMessage(ctx context.Context, messageID uuid.UUID) ([]models.Reaction, error) {
	var reactions []models.Reaction
	err := r.db.WithContext(ctx).Where("message_id = ?", messageID).Find(&reactions).Error
	return reactions, err
}

func (r *messageRepository) SearchByConversation(ctx context.Context, convID uuid.UUID, query string, limit, offset int) ([]models.Message, error) {
	var msgs []models.Message
	err := r.db.WithContext(ctx).
		Where("conversation_id = ? AND content ILIKE ?", convID, "%"+query+"%").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&msgs).Error
	return msgs, err
}
