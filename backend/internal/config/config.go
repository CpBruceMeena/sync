package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	ServerPort int
	JWTSecret  string
	AccessTTL  int // minutes
	RefreshTTL int // days
	UploadDir  string
}

func Load() *Config {
	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "sync"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		ServerPort: getEnvInt("SERVER_PORT", 8080),
		JWTSecret:  getEnv("JWT_SECRET", "super-secret-key-change-in-production"),
		AccessTTL:  getEnvInt("JWT_ACCESS_TTL", 15), // 15 minutes
		RefreshTTL: getEnvInt("JWT_REFRESH_TTL", 7), // 7 days
		UploadDir:  getEnv("UPLOAD_DIR", "./uploads"),
	}
	return cfg
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
