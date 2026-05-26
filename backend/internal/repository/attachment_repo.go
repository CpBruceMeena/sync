package repository

import (
	"context"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type attachmentRepository struct {
	db *gorm.DB
}

func NewAttachmentRepository(db *gorm.DB) AttachmentRepository {
	return &attachmentRepository{db: db}
}

func (r *attachmentRepository) Create(ctx context.Context, attachment *models.Attachment) error {
	return r.db.WithContext(ctx).Create(attachment).Error
}

func (r *attachmentRepository) GetByMessageID(ctx context.Context, messageID uuid.UUID) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.WithContext(ctx).Where("message_id = ?", messageID).Find(&attachments).Error
	return attachments, err
}
