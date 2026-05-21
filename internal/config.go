package internal

import (
	"bufio"
	"os"
	"strings"

	"log/slog"
)

// Config holds the AI service configuration.
type Config struct {
	BaseURL string
	APIKey  string
	Model   string
	LogLevel slog.Level
}

// LoadConfig reads configuration from environment variables or .env file.
func LoadConfig() *Config {
	cfg := &Config{
		BaseURL:  getEnv("AI_BASE_URL", "https://api.openai.com/v1"),
		APIKey:   getEnv("AI_API_KEY", ""),
		Model:    getEnv("AI_MODEL", "gpt-4o"),
		LogLevel: slog.LevelInfo,
	}

	if lvl := getEnv("AI_LOG_LEVEL", ""); lvl != "" {
		var level slog.Level
		if err := level.UnmarshalText([]byte(lvl)); err == nil {
			cfg.LogLevel = level
		}
	}

	// Try loading .env file if godotenv is available
	loadEnvFile()

	// Re-read after .env load (env file overrides defaults but not explicit env vars)
	if cfg.BaseURL == "https://api.openai.com/v1" {
		cfg.BaseURL = getEnv("AI_BASE_URL", cfg.BaseURL)
	}
	if cfg.APIKey == "" {
		cfg.APIKey = getEnv("AI_API_KEY", cfg.APIKey)
	}
	if cfg.Model == "gpt-4o" {
		cfg.Model = getEnv("AI_MODEL", cfg.Model)
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func loadEnvFile() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			k = strings.TrimSpace(k)
			v = strings.TrimSpace(v)
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
	}
}
