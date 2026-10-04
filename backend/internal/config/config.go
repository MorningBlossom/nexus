package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	GithubWebhookSecret     string
	GithubAppId             int64
	GithubAppPrivateKeyPath string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		GithubWebhookSecret:     os.Getenv("GITHUB_WEBHOOK_SECRET"),
		GithubAppId:             getInt64Env("GITHUB_APP_ID", 0),
		GithubAppPrivateKeyPath: os.Getenv("GITHUB_PRIVATE_KEY_PATH"),
	}
}

func getInt64Env(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}

	return parsed
}
