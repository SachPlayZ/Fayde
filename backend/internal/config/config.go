// Package config loads application configuration from environment variables.
package config

import (
	"errors"
	"os"
)

// Config holds all application configuration values.
type Config struct {
	DatabaseURL        string
	JWTSecret          string
	Port               string
	CORSOrigin         string
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	S3Bucket           string
	SMTPHost           string
	SMTPPort           string
	SMTPUser           string
	SMTPPass           string
	FromEmail          string
}

// Load reads configuration from environment variables.
// DATABASE_URL and JWT_SECRET are required; PORT defaults to "8080"
// and CORS_ORIGIN defaults to "http://localhost:3000".
// AWS_* and S3_BUCKET are optional; if S3_BUCKET is empty, attachment
// endpoints will return 501 Not Implemented.
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3000"
	}

	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "587"
	}
	fromEmail := os.Getenv("FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = os.Getenv("SMTP_USER")
	}

	return &Config{
		DatabaseURL:        dbURL,
		JWTSecret:          jwtSecret,
		Port:               port,
		CORSOrigin:         corsOrigin,
		AWSRegion:          os.Getenv("AWS_REGION"),
		AWSAccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		S3Bucket:           os.Getenv("S3_BUCKET"),
		SMTPHost:           os.Getenv("SMTP_HOST"),
		SMTPPort:           smtpPort,
		SMTPUser:           os.Getenv("SMTP_USER"),
		SMTPPass:           os.Getenv("SMTP_PASS"),
		FromEmail:          fromEmail,
	}, nil
}
