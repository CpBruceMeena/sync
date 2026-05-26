package repository

import (
	"context"
	"errors"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type conversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) Create(ctx context.Context, conv *models.Conversation) error {
	return r.db.WithContext(ctx).Create(conv).Error
}

func (r *conversationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *conversationRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Conversation, error) {
	var convs []models.Conversation
	err := r.db.WithContext(ctx).
		Joins("JOIN conversation_members ON conversation_members.conversation_id = conversations.id").
		Where("conversation_members.user_id = ?", userID).
		Order("conversations.updated_at DESC").
		Find(&convs).Error
	return convs, err
}

func (r *conversationRepository) ListPublic(ctx context.Context, limit, offset int) ([]models.Conversation, error) {
	var convs []models.Conversation
	err := r.db.WithContext(ctx).
		Where("type = 'group' AND is_public = ?", true).
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&convs).Error
	return convs, err
}

func (r *conversationRepository) SearchPublic(ctx context.Context, query string, limit int) ([]models.Conversation, error) {
	var convs []models.Conversation
	pattern := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Where("type = 'group' AND is_public = ? AND name ILIKE ?", true, pattern).
		Order("updated_at DESC").
		Limit(limit).
		Find(&convs).Error
	return convs, err
}

func (r *conversationRepository) FindPrivate(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.WithContext(ctx).
		Where("type = 'private'").
		Where("EXISTS (SELECT 1 FROM conversation_members cm1 WHERE cm1.conversation_id = conversations.id AND cm1.user_id = ?)", userID1).
		Where("EXISTS (SELECT 1 FROM conversation_members cm2 WHERE cm2.conversation_id = conversations.id AND cm2.user_id = ?)", userID2).
		First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *conversationRepository) AddMember(ctx context.Context, member *models.ConversationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *conversationRepository) RemoveMember(ctx context.Context, convID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("conversation_id = ? AND user_id = ?", convID, userID).
		Delete(&models.ConversationMember{}).Error
}

func (r *conversationRepository) GetMembers(ctx context.Context, convID uuid.UUID) ([]models.ConversationMember, error) {
	var members []models.ConversationMember
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", convID).
		Order("joined_at ASC").
		Find(&members).Error
	return members, err
}

func (r *conversationRepository) IsMember(ctx context.Context, convID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", convID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *conversationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Conversation{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("conversation not found")
	}
	return nil
}
