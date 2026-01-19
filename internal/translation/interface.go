package translation

import (
	"context"
)

// TranslationService defines the interface for translation operations
// This interface enables mocking for testing and allows alternative implementations
type TranslationService interface {
	// TranslateText translates text from source language to target language
	TranslateText(ctx context.Context, text string, sourceLanguage string, targetLanguage string) (string, error)
}

// DefaultTranslationService is the default implementation using Google Cloud Translation API
type DefaultTranslationService struct{}

// TranslateText implements TranslationService interface
func (s *DefaultTranslationService) TranslateText(ctx context.Context, text string, sourceLanguage string, targetLanguage string) (string, error) {
	return TranslateText(ctx, text, sourceLanguage, targetLanguage)
}
