package tts

import (
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

// VoiceConfig holds voice configuration for a language
type VoiceConfig struct {
	LanguageCode string
	VoiceName    string
	Gender       texttospeechpb.SsmlVoiceGender
}

// GetVoiceConfig returns voice configuration for a language
// Returns nil if language is not supported
func GetVoiceConfig(language string) *VoiceConfig {
	configs := map[string]*VoiceConfig{
		"en": {
			LanguageCode: "en-US",
			VoiceName:    "en-US-Neural2-F", // Natural female voice
			Gender:       texttospeechpb.SsmlVoiceGender_FEMALE,
		},
		"ar": {
			LanguageCode: "ar-XA",
			VoiceName:    "ar-XA-Wavenet-A", // Arabic voice
			Gender:       texttospeechpb.SsmlVoiceGender_FEMALE,
		},
		"de": {
			LanguageCode: "de-DE",
			VoiceName:    "de-DE-Neural2-F",
			Gender:       texttospeechpb.SsmlVoiceGender_FEMALE,
		},
		"ru": {
			LanguageCode: "ru-RU",
			VoiceName:    "ru-RU-Wavenet-E", // Russian voice
			Gender:       texttospeechpb.SsmlVoiceGender_FEMALE,
		},
		"fr": {
			LanguageCode: "fr-FR",
			VoiceName:    "fr-FR-Neural2-C",
			Gender:       texttospeechpb.SsmlVoiceGender_FEMALE,
		},
	}

	return configs[language]
}

// GetSpeakingRate returns the average speaking rate (words per minute) for a language
func GetSpeakingRate(language string) float64 {
	rates := map[string]float64{
		"en": 150.0,
		"ar": 140.0,
		"de": 145.0,
		"ru": 140.0,
		"fr": 145.0,
	}

	rate, exists := rates[language]
	if !exists {
		return 150.0 // Default
	}
	return rate
}
