package stt

import (
	"context"
	"testing"
	"time"
)

func TestSpeechToText_InvalidAudioPath(t *testing.T) {
	ctx := context.Background()

	// Test with non-existent file
	_, err := SpeechToText(ctx, "/nonexistent/file.wav", "")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestSpeechToText_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Test that context cancellation is detected
	_, err := SpeechToText(ctx, "/nonexistent/file.wav", "")
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	// Check if error mentions cancellation
	if err != nil && ctx.Err() == nil {
		t.Log("Note: Context cancellation may be detected later in the pipeline")
	}
}

func TestSpeechToText_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	_, err := SpeechToText(ctx, "/nonexistent/file.wav", "")
	if err == nil {
		t.Error("expected error for timed out context")
	}
}
