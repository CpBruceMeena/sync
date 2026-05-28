package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Username     string    `gorm:"unique;not null;size:50" json:"username"`
	Email        string    `gorm:"unique;not null;size:255" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	DisplayName  string    `gorm:"column:display_name;size:100" json:"display_name"`
	AvatarUrl    string    `gorm:"column:avatar_url;size:500" json:"avatar_url"`
	Status       string    `gorm:"default:'offline';size:20" json:"status"`
	Bio          string    `gorm:"size:500" json:"bio"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
