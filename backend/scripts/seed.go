package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type testUser struct {
	Username string
	Email    string
	Password string
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/sync?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Connected to PostgreSQL database")

	users := []testUser{
		{Username: "alice", Email: "alice@test.com", Password: "password123"},
		{Username: "bob", Email: "bob@test.com", Password: "password123"},
		{Username: "charlie", Email: "charlie@test.com", Password: "password123"},
		{Username: "diana", Email: "diana@test.com", Password: "password123"},
		{Username: "eve", Email: "eve@test.com", Password: "password123"},
	}

	fmt.Println("\nSeeding users...")
	for _, u := range users {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash password for %s: %v", u.Username, err)
			continue
		}

		var userID string
		err = pool.QueryRow(ctx,
			`INSERT INTO users (username, email, password_hash, display_name, status)
			 VALUES ($1, $2, $3, $4, 'offline')
			 ON CONFLICT (username) DO UPDATE SET username = EXCLUDED.username
			 RETURNING id`,
			u.Username, u.Email, string(hash), u.Username,
		).Scan(&userID)

		if err != nil {
			log.Printf("Failed to insert user %s: %v", u.Username, err)
			continue
		}

		fmt.Printf("  ✓ Created user: %-10s (ID: %s)\n", u.Username, userID)
	}

	fmt.Println("\n✓ All users seeded successfully!")
	fmt.Println("\n--- Test Credentials ---")
	for _, u := range users {
		fmt.Printf("  Username: %-10s Email: %-20s Password: %s\n", u.Username, u.Email, u.Password)
	}
	fmt.Println("------------------------")
}
