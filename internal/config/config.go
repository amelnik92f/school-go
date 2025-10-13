package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	Env                    string
	DBPath                 string
	FetchSchedule          string
	APITimeout             time.Duration
	APIKey                 string
	GeminiAPIKey           string
	OpenRouteServiceAPIKey string
}

func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if it doesn't)
	_ = godotenv.Load()

	cfg := &Config{
		Port:                   getEnv("PORT", "8080"),
		Env:                    getEnv("ENV", "development"),
		DBPath:                 getEnv("DB_PATH", "./data/schools.db"),
		FetchSchedule:          getEnv("FETCH_SCHEDULE", "0 2 * * *"), // 2 AM daily
		APITimeout:             parseDuration(getEnv("API_TIMEOUT", "30s"), 30*time.Second),
		APIKey:                 getEnv("API_KEY", ""),
		GeminiAPIKey:           getEnv("GEMINI_API_KEY", ""),
		OpenRouteServiceAPIKey: getEnv("OPENROUTESERVICE_API_KEY", ""),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string, defaultValue time.Duration) time.Duration {
	duration, err := time.ParseDuration(s)
	if err != nil {
		return defaultValue
	}
	return duration
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func (c *Config) GetServerAddr() string {
	return fmt.Sprintf(":%s", c.Port)
}
