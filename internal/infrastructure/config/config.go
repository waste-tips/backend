package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	ProjectID        string
	ApplicationName  string
	RecaptchaSiteKey string
	GCPEnabled       bool
	LogLevel         int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		ProjectID:        getEnv("PROJECT_ID", "waste-tips"),
		ApplicationName:  getEnv("APPLICATION_NAME", "Waste Tips"),
		RecaptchaSiteKey: getEnv("RECAPTCHA_SITE_KEY", ""),
		GCPEnabled:       getEnv("GCP_ENABLED", "true") == "true",
		LogLevel:         100, // Default log level
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
