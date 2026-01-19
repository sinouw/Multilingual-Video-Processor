package stt

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
	"google.golang.org/api/option"
)

// SpeechToTextResponse represents the response from Google Cloud Speech-to-Text API
type SpeechToTextResponse struct {
	Text     string `json:"text"`
	Language string `json:"language,omitempty"` // Detected language code
}

// SpeechToText converts audio to text using Google Cloud Speech-to-Text API
// languageHint: Optional language code hint (e.g., "fr", "en"). If empty, Google Cloud Speech-to-Text will auto-detect.
func SpeechToText(ctx context.Context, audioPath string, languageHint string) (*SpeechToTextResponse, error) {
	slog.Info("Converting speech to text", "audioPath", audioPath, "languageHint", languageHint)

	// Initialize Speech-to-Text client
	// Use service account from environment or default credentials
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	var client *speech.Client
	var err error

	if credentialsPath != "" {
		client, err = speech.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
		if err != nil {
			slog.Warn("Failed to create client with credentials file, trying default", "error", err)
			client, err = speech.NewClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to create Speech-to-Text client: %w", err)
			}
		}
	} else {
		client, err = speech.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create Speech-to-Text client: %w", err)
		}
	}
	defer client.Close()

	// Read audio file
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	// Check context cancellation before making API call
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("speech-to-text cancelled: %w", ctx.Err())
	default:
	}

	// Build recognition config
	config := &speechpb.RecognitionConfig{
		Encoding:        speechpb.RecognitionConfig_LINEAR16,
		SampleRateHertz: 16000,
	}

	// Set language code if hint is provided, otherwise auto-detect
	if languageHint != "" {
		config.LanguageCode = languageHint
		slog.Info("Using language hint", "language", languageHint)
	} else {
		slog.Info("No language hint provided, Google Cloud Speech-to-Text will auto-detect")
	}

	// Build recognition audio
	audio := &speechpb.RecognitionAudio{
		AudioSource: &speechpb.RecognitionAudio_Content{
			Content: audioData,
		},
	}

	// Build the request
	req := &speechpb.RecognizeRequest{
		Config: config,
		Audio:  audio,
	}

	// Perform the recognition with context
	resp, err := client.Recognize(ctx, req)
	if err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() != nil {
			return nil, fmt.Errorf("speech-to-text cancelled: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to recognize speech: %w", err)
	}

	// Extract transcribed text and detected language
	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no speech recognition results returned")
	}

	// Concatenate all alternative transcripts
	var fullText strings.Builder
	for _, result := range resp.Results {
		if len(result.Alternatives) > 0 {
			if fullText.Len() > 0 {
				fullText.WriteString(" ")
			}
			fullText.WriteString(result.Alternatives[0].Transcript)
		}
	}

	transcribedText := fullText.String()
	if transcribedText == "" {
		return nil, fmt.Errorf("no transcribed text found in results")
	}

	// Use language hint if provided, otherwise try to detect from response
	detectedLanguage := languageHint
	if detectedLanguage == "" {
		// Try to get detected language from response
		if len(resp.Results) > 0 && len(resp.Results[0].Alternatives) > 0 {
			// Note: The API might not always return detected language in this format
			// In that case, we'll use the language from config
			if resp.Results[0].LanguageCode != "" {
				detectedLanguage = resp.Results[0].LanguageCode
			}
		}
		// If still empty, use a default or return empty
		if detectedLanguage == "" {
			detectedLanguage = "" // Return empty to indicate auto-detection was used
		}
	}

	slog.Info("Speech-to-text completed",
		"textLength", len(transcribedText),
		"detectedLanguage", detectedLanguage)

	return &SpeechToTextResponse{
		Text:     transcribedText,
		Language: detectedLanguage,
	}, nil
}
