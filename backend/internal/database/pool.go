package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool    *pgxpool.Pool
	Queries *Queries
}

func NewDB(ctx context.Context, dsn string) (*DB, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	config.MaxConns = 25

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Println("Connected to PostgreSQL database")
	return &DB{
		Pool:    pool,
		Queries: New(pool),
	}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

func (db *DB) RunMigrations(ctx context.Context, schemaDir string) error {
	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		return fmt.Errorf("read schema directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		path := filepath.Join(schemaDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read schema file %s: %w", file, err)
		}

		if _, err := db.Pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("execute schema %s: %w", file, err)
		}
		log.Printf("Applied migration: %s", file)
	}

	return nil
}
