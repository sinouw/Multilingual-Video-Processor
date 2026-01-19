package stt

import (
	"context"
)

// SpeechToTextService defines the interface for Speech-to-Text operations
// This interface enables mocking for testing and allows alternative implementations
type SpeechToTextService interface {
	// SpeechToText converts audio to text
	SpeechToText(ctx context.Context, audioPath string, languageHint string) (*SpeechToTextResponse, error)

	// ExtractAudioFromVideo extracts audio from video file
	ExtractAudioFromVideo(ctx context.Context, videoPath string) (string, error)
}

// DefaultSpeechToTextService is the default implementation using Google Cloud Speech-to-Text API
type DefaultSpeechToTextService struct{}

// SpeechToText implements SpeechToTextService interface
func (s *DefaultSpeechToTextService) SpeechToText(ctx context.Context, audioPath string, languageHint string) (*SpeechToTextResponse, error) {
	return SpeechToText(ctx, audioPath, languageHint)
}

// ExtractAudioFromVideo implements SpeechToTextService interface
func (s *DefaultSpeechToTextService) ExtractAudioFromVideo(ctx context.Context, videoPath string) (string, error) {
	return ExtractAudioFromVideo(ctx, videoPath)
}
