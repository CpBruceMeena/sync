package repository

import (
	"gorm.io/gorm"
)

// GormDB holds the GORM database connection
type GormDB struct {
	DB *gorm.DB
}

// NewGormDB creates a new GormDB wrapper
func NewGormDB(db *gorm.DB) *GormDB {
	return &GormDB{DB: db}
}
