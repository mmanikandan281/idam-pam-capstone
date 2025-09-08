package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	AWSRegion   string
	KMSKeyID    string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "5000"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/idam_pam?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		AWSRegion:   getEnv("AWS_REGION", "us-west-2"),
		KMSKeyID:    getEnv("KMS_KEY_ID", "alias/idam-pam-key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
