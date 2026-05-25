package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	DB *gorm.DB
}

func NewDB(dsn string) (*DB, error) {
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Println("Connected to PostgreSQL database")
	return &DB{DB: db}, nil
}

func (db *DB) Close() {
	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Printf("Error getting underlying sql db: %v", err)
		return
	}
	sqlDB.Close()
}
