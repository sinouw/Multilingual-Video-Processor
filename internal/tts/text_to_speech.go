package tts

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"
)

// GenerateTTS generates text-to-speech audio using Google Cloud TTS
func GenerateTTS(ctx context.Context, text string, language string, originalDuration float64, outputPath string) error {
	slog.Info("Generating TTS",
		"language", language,
		"textLength", len(text),
		"originalDuration", originalDuration)

	// Initialize TTS client
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	var client *texttospeech.Client
	var err error

	if credentialsPath != "" {
		client, err = texttospeech.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
		if err != nil {
			slog.Warn("Failed to create client with credentials file, trying default", "error", err)
			client, err = texttospeech.NewClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to create TTS client: %w", err)
			}
		}
	} else {
		client, err = texttospeech.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create TTS client: %w", err)
		}
	}
	defer client.Close()

	// Get voice configuration for language
	voiceConfig := GetVoiceConfig(language)
	if voiceConfig == nil {
		return fmt.Errorf("unsupported language for TTS: %s", language)
	}

	// Calculate speed adjustment to match original duration
	speedRatio := calculateSpeedRatio(text, originalDuration, language)
	ssmlText := buildSSML(text, speedRatio)

	// Check context cancellation before making API call
	select {
	case <-ctx.Done():
		return fmt.Errorf("TTS generation cancelled: %w", ctx.Err())
	default:
	}

	// Build the request
	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Ssml{
				Ssml: ssmlText,
			},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: voiceConfig.LanguageCode,
			Name:         voiceConfig.VoiceName,
			SsmlGender:   voiceConfig.Gender,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding:   texttospeechpb.AudioEncoding_MP3,
			SpeakingRate:    1.0, // Speed controlled via SSML
			SampleRateHertz: 24000,
		},
	}

	// Perform the text-to-speech request with context
	resp, err := client.SynthesizeSpeech(ctx, req)
	if err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() != nil {
			return fmt.Errorf("TTS generation cancelled: %w", ctx.Err())
		}
		return fmt.Errorf("failed to synthesize speech: %w", err)
	}

	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the audio content to file
	err = os.WriteFile(outputPath, resp.AudioContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write audio file: %w", err)
	}

	slog.Info("TTS audio generated successfully", "outputPath", outputPath)
	return nil
}

// calculateSpeedRatio calculates the speed ratio to match original audio duration
// This is an approximation - actual TTS duration may vary
func calculateSpeedRatio(text string, originalDuration float64, language string) float64 {
	avgRate := GetSpeakingRate(language)

	// Estimate words in text
	words := len(strings.Fields(text))

	// Calculate expected duration at normal speed
	expectedDuration := float64(words) / avgRate * 60.0 // Convert to seconds

	if expectedDuration == 0 || originalDuration == 0 {
		return 1.0
	}

	// Calculate speed ratio
	speedRatio := expectedDuration / originalDuration

	// Clamp speed ratio to reasonable range (0.5x to 2.0x)
	if speedRatio < 0.5 {
		speedRatio = 0.5
	} else if speedRatio > 2.0 {
		speedRatio = 2.0
	}

	slog.Debug("Speed ratio calculated",
		"speedRatio", speedRatio,
		"originalDuration", originalDuration,
		"expectedDuration", expectedDuration)

	return speedRatio
}

// buildSSML builds SSML text with speed control
func buildSSML(text string, speedRatio float64) string {
	// Escape XML special characters
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")

	// Build SSML with prosody for speed control
	speedPercent := int(speedRatio * 100)
	if speedPercent < 50 {
		speedPercent = 50
	} else if speedPercent > 200 {
		speedPercent = 200
	}

	ssml := fmt.Sprintf(`<speak><prosody rate="%d%%">%s</prosody></speak>`, speedPercent, text)
	return ssml
}
