package internal

import (
	"os"

	"log/slog"

	"github.com/joho/godotenv"
)

// Config holds the AI service configuration.
type Config struct {
	BaseURL      string
	APIKey       string
	Model        string
	LogLevel     slog.Level
	SystemPrompt string
}

// LoadConfig reads configuration from environment variables or .env file.
func LoadConfig() *Config {
	// Load .env file if it exists (does not override existing env vars)
	godotenv.Load()

	cfg := &Config{
		BaseURL:      getEnv("AI_BASE_URL", "https://api.openai.com/v1"),
		APIKey:       getEnv("AI_API_KEY", ""),
		Model:        getEnv("AI_MODEL", "gpt-4o"),
		LogLevel:     slog.LevelInfo,
		SystemPrompt: getEnv("AI_SYSTEM_PROMPT", ""),
	}

	if lvl := getEnv("AI_LOG_LEVEL", ""); lvl != "" {
		var level slog.Level
		if err := level.UnmarshalText([]byte(lvl)); err == nil {
			cfg.LogLevel = level
		}
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
