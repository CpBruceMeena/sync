package repository

import (
	"context"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type presenceRepository struct {
	db *gorm.DB
}

func NewPresenceRepository(db *gorm.DB) PresenceRepository {
	return &presenceRepository{db: db}
}

func (r *presenceRepository) Upsert(ctx context.Context, presence *models.Presence) error {
	return r.db.WithContext(ctx).Save(presence).Error
}

func (r *presenceRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Presence, error) {
	var presence models.Presence
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&presence).Error
	if err != nil {
		return nil, err
	}
	return &presence, nil
}

func (r *presenceRepository) GetOnline(ctx context.Context) ([]models.Presence, error) {
	var presences []models.Presence
	err := r.db.WithContext(ctx).Where("status != ?", "offline").Find(&presences).Error
	return presences, err
}
