package tts

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateTTS_UnsupportedLanguage(t *testing.T) {
	ctx := context.Background()
	tmpDir := os.TempDir()
	outputPath := filepath.Join(tmpDir, "test_output.mp3")

	err := GenerateTTS(ctx, "Hello", "xx", 10.0, outputPath)
	if err == nil {
		t.Error("expected error for unsupported language")
	}

	// Check error message mentions unsupported language
	if err != nil && err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestGenerateTTS_InvalidOutputPath(t *testing.T) {
	ctx := context.Background()

	// Test with invalid output path (read-only directory or invalid characters)
	err := GenerateTTS(ctx, "Hello", "en", 10.0, "")
	if err == nil {
		t.Error("expected error for empty output path")
	}
}

func TestGenerateTTS_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tmpDir := os.TempDir()
	outputPath := filepath.Join(tmpDir, "test_output.mp3")

	err := GenerateTTS(ctx, "Hello", "en", 10.0, outputPath)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestGenerateTTS_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	tmpDir := os.TempDir()
	outputPath := filepath.Join(tmpDir, "test_output.mp3")

	err := GenerateTTS(ctx, "Hello", "en", 10.0, outputPath)
	if err == nil {
		t.Error("expected error for timed out context")
	}
}

func TestGetVoiceConfig(t *testing.T) {
	tests := []struct {
		language string
		wantNil  bool
	}{
		{"en", false},
		{"ar", false},
		{"de", false},
		{"ru", false},
		{"xx", true},  // Unsupported
		{"", true},    // Empty
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			config := GetVoiceConfig(tt.language)
			if tt.wantNil && config != nil {
				t.Errorf("expected nil config for language %s", tt.language)
			}
			if !tt.wantNil && config == nil {
				t.Errorf("expected non-nil config for language %s", tt.language)
			}
		})
	}
}
