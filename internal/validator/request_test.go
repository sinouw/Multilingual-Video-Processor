package validator

import (
	"testing"

	"github.com/sinouw/multilingual-video-processor/internal/config"
	"github.com/sinouw/multilingual-video-processor/pkg/models"
)

func TestValidateTranslateRequest(t *testing.T) {
	cfg := &config.Config{
		SupportedLanguages: []string{"en", "ar", "de"},
	}

	tests := []struct {
		name    string
		req     *models.TranslateRequest
		wantErr bool
	}{
		{
			"valid request",
			&models.TranslateRequest{
				VideoURL:        "gs://bucket/path/to/video.mp4",
				TargetLanguages: []string{"en", "ar"},
				SourceLanguage:  "fr",
			},
			false,
		},
		{
			"missing video URL",
			&models.TranslateRequest{
				TargetLanguages: []string{"en"},
			},
			true,
		},
		{
			"invalid target language",
			&models.TranslateRequest{
				VideoURL:        "gs://bucket/video.mp4",
				TargetLanguages: []string{"fr"},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTranslateRequest(tt.req, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTranslateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateVideoURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid GCS URL", "gs://bucket/path/to/video.mp4", false},
		{"valid HTTPS URL", "https://example.com/video.mp4", false},
		{"empty URL", "", true},
		{"invalid format", "invalid-url", true},
		{"invalid GCS format", "gs://bucket", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVideoURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVideoURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
