package translation

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestTranslateText_MissingAPIKey(t *testing.T) {
	// Save original API key
	originalKey := os.Getenv("GOOGLE_TRANSLATE_API_KEY")
	defer os.Setenv("GOOGLE_TRANSLATE_API_KEY", originalKey)

	// Remove API key
	os.Unsetenv("GOOGLE_TRANSLATE_API_KEY")

	ctx := context.Background()
	_, err := TranslateText(ctx, "Hello", "en", "fr")
	if err == nil {
		t.Error("expected error for missing API key")
	}

	// Check error message
	if err != nil && err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestTranslateText_EmptyText(t *testing.T) {
	// Set a dummy API key to avoid missing key error
	os.Setenv("GOOGLE_TRANSLATE_API_KEY", "test-key")
	defer os.Unsetenv("GOOGLE_TRANSLATE_API_KEY")

	ctx := context.Background()
	result, err := TranslateText(ctx, "", "en", "fr")

	// Empty text might be valid (translate to empty string) or invalid depending on API
	// We just check it doesn't crash
	if err != nil {
		t.Logf("API rejected empty text (expected): %v", err)
	} else if result == "" {
		t.Log("API returned empty string for empty input")
	}
}

func TestTranslateText_ContextCancellation(t *testing.T) {
	os.Setenv("GOOGLE_TRANSLATE_API_KEY", "test-key")
	defer os.Unsetenv("GOOGLE_TRANSLATE_API_KEY")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := TranslateText(ctx, "Hello", "en", "fr")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestTranslateText_ContextTimeout(t *testing.T) {
	os.Setenv("GOOGLE_TRANSLATE_API_KEY", "test-key")
	defer os.Unsetenv("GOOGLE_TRANSLATE_API_KEY")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	_, err := TranslateText(ctx, "Hello", "en", "fr")
	if err == nil {
		t.Error("expected error for timed out context")
	}
}
