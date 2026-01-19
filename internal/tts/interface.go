package tts

import (
	"context"
)

// TTSService defines the interface for Text-to-Speech operations
// This interface enables mocking for testing and allows alternative implementations
type TTSService interface {
	// GenerateTTS generates text-to-speech audio
	GenerateTTS(ctx context.Context, text string, language string, originalDuration float64, outputPath string) error
}

// DefaultTTSService is the default implementation using Google Cloud TTS API
type DefaultTTSService struct{}

// GenerateTTS implements TTSService interface
func (s *DefaultTTSService) GenerateTTS(ctx context.Context, text string, language string, originalDuration float64, outputPath string) error {
	return GenerateTTS(ctx, text, language, originalDuration, outputPath)
}
