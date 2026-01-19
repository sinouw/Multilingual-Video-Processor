package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sinouw/multilingual-video-processor/internal/config"
	"github.com/sinouw/multilingual-video-processor/pkg/models"
)

// ValidateTranslateRequest validates a translation request
func ValidateTranslateRequest(req *models.TranslateRequest, cfg *config.Config) error {
	// Validate video URL
	if err := ValidateVideoURL(req.VideoURL); err != nil {
		return fmt.Errorf("invalid video URL: %w", err)
	}

	// Validate target languages
	if err := ValidateLanguageCodes(req.TargetLanguages, cfg.SupportedLanguages); err != nil {
		return fmt.Errorf("invalid target languages: %w", err)
	}

	// Validate source language if provided
	if req.SourceLanguage != "" {
		// Source language can be auto-detect or any valid language code
		// We don't restrict it to supported languages as it's just a hint
		if !isValidLanguageCode(req.SourceLanguage) {
			return fmt.Errorf("invalid source language code: %s", req.SourceLanguage)
		}
	}

	return nil
}

// ValidateVideoURL validates a video URL format
func ValidateVideoURL(url string) error {
	if url == "" {
		return fmt.Errorf("video URL is required")
	}

	// Check for GCS URL format (gs://bucket/path)
	if strings.HasPrefix(url, "gs://") {
		// Basic validation for gs:// URL
		parts := strings.SplitN(url[5:], "/", 2)
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid GCS URL format: %s", url)
		}
		return nil
	}

	// Check for HTTPS URL
	if strings.HasPrefix(url, "https://") {
		// Basic URL validation
		matched, err := regexp.MatchString(`^https://[a-zA-Z0-9\-._]+(/.*)?$`, url)
		if err != nil || !matched {
			return fmt.Errorf("invalid HTTPS URL format: %s", url)
		}
		return nil
	}

	return fmt.Errorf("unsupported URL format: %s (must be gs:// or https://)", url)
}

// isValidLanguageCode performs basic language code validation (ISO 639-1 format)
func isValidLanguageCode(code string) bool {
	// Basic validation: 2-5 character language code (e.g., "en", "en-US")
	matched, _ := regexp.MatchString(`^[a-z]{2}(-[A-Z]{2,3})?$`, strings.ToLower(code))
	return matched
}
