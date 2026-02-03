package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	HTTPAddress string
	DatabaseURL string
	CredentialsKey string
	CredentialsKeyID string
	AdminAPIKey string
	ViewerAPIKey string
	RetentionDays int
}

func NewConfig() (Config, error) {
	httpAddress := readEnv("APP_HTTP_ADDRESS", ":8080")
	databaseURL := readEnv("DATABASE_URL", "")
	credentialsKey := readEnv("CREDENTIALS_KEY", "")
	credentialsKeyID := readEnv("CREDENTIALS_KEY_ID", "default")
	adminAPIKey := readEnv("ADMIN_API_KEY", "")
	viewerAPIKey := readEnv("VIEWER_API_KEY", "")
	retentionDays := readEnvInt("RETENTION_DAYS", 90)

	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if credentialsKey == "" {
		return Config{}, errors.New("CREDENTIALS_KEY is required")
	}
	if adminAPIKey == "" {
		return Config{}, errors.New("ADMIN_API_KEY is required")
	}

	return Config{
		HTTPAddress: httpAddress,
		DatabaseURL: databaseURL,
		CredentialsKey: credentialsKey,
		CredentialsKeyID: credentialsKeyID,
		AdminAPIKey: adminAPIKey,
		ViewerAPIKey: viewerAPIKey,
		RetentionDays: retentionDays,
	}, nil
}

func readEnv(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if ok {
		return value
	}

	return fallback
}

func readEnvInt(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}
