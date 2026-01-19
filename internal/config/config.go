package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	GoogleCredentialsPath     string
	TranslateAPIKey           string
	GCSInputBucket            string
	GCSOutputBucket           string
	SupportedLanguages        []string
	DefaultSourceLanguage     string
	MaxVideoDuration          time.Duration
	MaxVideoSizeMB            int
	MaxConcurrentJobs         int
	MaxConcurrentTranslations int
	RequestTimeout            time.Duration
	LogLevel                  string
	APIVersion                string
	EnableHealthCheck         bool
	RateLimitRPM              int
	WebhookURL                string
	CORSOrigins               []string
	JobTTL                    time.Duration
	MaxRequestBodySize        int64
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() (*Config, error) {
	cfg := &Config{
		GoogleCredentialsPath:     getEnv("GOOGLE_APPLICATION_CREDENTIALS", ""),
		TranslateAPIKey:           getEnv("GOOGLE_TRANSLATE_API_KEY", ""),
		GCSInputBucket:            getEnv("GCS_BUCKET_INPUT", ""),
		GCSOutputBucket:           getEnv("GCS_BUCKET_OUTPUT", ""),
		SupportedLanguages:        parseStringSlice(getEnv("SUPPORTED_LANGUAGES", "en,ar,de,ru")),
		DefaultSourceLanguage:     getEnv("SOURCE_LANGUAGE", ""),
		MaxVideoDuration:          parseDuration(getEnv("MAX_VIDEO_DURATION", "600")),
		MaxVideoSizeMB:            parseInt(getEnv("MAX_VIDEO_SIZE_MB", "500")),
		MaxConcurrentJobs:         parseInt(getEnv("MAX_CONCURRENT_JOBS", "10")),
		MaxConcurrentTranslations: parseInt(getEnv("MAX_CONCURRENT_TRANSLATIONS", "3")),
		RequestTimeout:            parseDuration(getEnv("REQUEST_TIMEOUT", "540")),
		LogLevel:                  getEnv("LOG_LEVEL", "info"),
		APIVersion:                getEnv("API_VERSION", "v1"),
		EnableHealthCheck:         parseBool(getEnv("ENABLE_HEALTH_CHECK", "true")),
		RateLimitRPM:              parseInt(getEnv("RATE_LIMIT_RPM", "60")),
		WebhookURL:                getEnv("WEBHOOK_URL", ""),
		CORSOrigins:               parseStringSlice(getEnv("CORS_ORIGINS", "*")),
		JobTTL:                    parseDurationString(getEnv("JOB_TTL", "24h")),
		MaxRequestBodySize:        parseInt64(getEnv("MAX_REQUEST_BODY_SIZE_BYTES", "1048576")),
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.GCSOutputBucket == "" {
		return fmt.Errorf("GCS_BUCKET_OUTPUT is required")
	}

	if len(c.SupportedLanguages) == 0 {
		return fmt.Errorf("at least one supported language must be specified")
	}

	if c.MaxVideoDuration <= 0 {
		return fmt.Errorf("MAX_VIDEO_DURATION must be greater than 0")
	}

	if c.MaxVideoSizeMB <= 0 {
		return fmt.Errorf("MAX_VIDEO_SIZE_MB must be greater than 0")
	}

	if c.MaxConcurrentTranslations <= 0 {
		return fmt.Errorf("MAX_CONCURRENT_TRANSLATIONS must be greater than 0")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[strings.ToLower(c.LogLevel)] {
		return fmt.Errorf("invalid LOG_LEVEL: %s (must be one of: debug, info, warn, error)", c.LogLevel)
	}

	return nil
}

// GetLoggerLevel returns the slog.Level based on LogLevel string
func (c *Config) GetLoggerLevel() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// IsLanguageSupported checks if a language code is supported
func (c *Config) IsLanguageSupported(lang string) bool {
	lang = strings.ToLower(strings.TrimSpace(lang))
	for _, supported := range c.SupportedLanguages {
		if strings.ToLower(supported) == lang {
			return true
		}
	}
	return false
}

// Helper functions

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func parseStringSlice(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseInt(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func parseDuration(value string) time.Duration {
	seconds := parseInt(value)
	return time.Duration(seconds) * time.Second
}

func parseDurationString(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		// Default to 24 hours if parsing fails
		return 24 * time.Hour
	}
	return duration
}

func parseInt64(value string) int64 {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func parseBool(value string) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return parsed
}
