package database

import (
	"log"

	"gorm.io/gorm"
)

// RunMigrations applies schema changes that cannot be handled by GORM AutoMigrate.
// We use manual SQL migrations for schema changes to maintain full control.
func RunMigrations(db *gorm.DB) error {
	migrations := []string{
		`ALTER TABLE conversations ADD COLUMN IF NOT EXISTS is_public BOOLEAN NOT NULL DEFAULT false;`,
		`CREATE TABLE IF NOT EXISTS presence (
			user_id UUID PRIMARY KEY,
			status VARCHAR(20) NOT NULL DEFAULT 'offline',
			last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
	}

	for _, m := range migrations {
		if err := db.Exec(m).Error; err != nil {
			log.Printf("Migration failed: %v", err)
			return err
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}
