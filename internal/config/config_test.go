package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Set test environment variables
	os.Setenv("GCS_BUCKET_OUTPUT", "test-bucket")
	os.Setenv("GOOGLE_TRANSLATE_API_KEY", "test-key")
	defer func() {
		os.Unsetenv("GCS_BUCKET_OUTPUT")
		os.Unsetenv("GOOGLE_TRANSLATE_API_KEY")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.GCSOutputBucket != "test-bucket" {
		t.Errorf("Expected GCSOutputBucket to be 'test-bucket', got '%s'", cfg.GCSOutputBucket)
	}

	if cfg.TranslateAPIKey != "test-key" {
		t.Errorf("Expected TranslateAPIKey to be 'test-key', got '%s'", cfg.TranslateAPIKey)
	}
}

func TestConfigValidation(t *testing.T) {
	cfg := &Config{
		GCSOutputBucket:           "",
		SupportedLanguages:        []string{},
		MaxVideoDuration:          0,
		MaxVideoSizeMB:            0,
		MaxConcurrentTranslations: 0,
		LogLevel:                  "invalid",
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation to fail")
	}
}

func TestIsLanguageSupported(t *testing.T) {
	cfg := &Config{
		SupportedLanguages: []string{"en", "ar", "de"},
	}

	if !cfg.IsLanguageSupported("en") {
		t.Error("Expected 'en' to be supported")
	}

	if cfg.IsLanguageSupported("fr") {
		t.Error("Expected 'fr' not to be supported")
	}
}
