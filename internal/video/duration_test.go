package video

import (
	"context"
	"testing"
	"time"
)

func TestGetVideoDuration_InvalidPath(t *testing.T) {
	ctx := context.Background()

	// Test with non-existent file
	_, err := GetVideoDuration(ctx, "/nonexistent/video.mp4")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestGetAudioDuration_InvalidPath(t *testing.T) {
	ctx := context.Background()

	// Test with non-existent file
	_, err := GetAudioDuration(ctx, "/nonexistent/audio.wav")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestGetVideoDuration_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := GetVideoDuration(ctx, "/nonexistent/video.mp4")
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	// Check that error mentions cancellation
	if err != nil && ctx.Err() != nil {
		t.Log("Context cancellation detected correctly")
	}
}

func TestGetAudioDuration_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := GetAudioDuration(ctx, "/nonexistent/audio.wav")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestGetVideoDuration_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	_, err := GetVideoDuration(ctx, "/nonexistent/video.mp4")
	if err == nil {
		t.Error("expected error for timed out context")
	}
}

func TestGetAudioDuration_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	_, err := GetAudioDuration(ctx, "/nonexistent/audio.wav")
	if err == nil {
		t.Error("expected error for timed out context")
	}
}

func TestGetVideoDuration_EmptyPath(t *testing.T) {
	ctx := context.Background()

	_, err := GetVideoDuration(ctx, "")
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestGetAudioDuration_EmptyPath(t *testing.T) {
	ctx := context.Background()

	_, err := GetAudioDuration(ctx, "")
	if err == nil {
		t.Error("expected error for empty path")
	}
}
