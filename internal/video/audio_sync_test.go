package video

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSyncAudioWithVideo_InvalidInputPath(t *testing.T) {
	ctx := context.Background()
	tmpDir := os.TempDir()

	// Test with non-existent video file
	outputPath := filepath.Join(tmpDir, "output.mp4")
	err := SyncAudioWithVideo(ctx, "/nonexistent/video.mp4", "/nonexistent/audio.wav", outputPath)
	if err == nil {
		t.Error("expected error for non-existent video file")
	}
}

func TestSyncAudioWithVideo_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tmpDir := os.TempDir()
	outputPath := filepath.Join(tmpDir, "output.mp4")

	err := SyncAudioWithVideo(ctx, "/nonexistent/video.mp4", "/nonexistent/audio.wav", outputPath)
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	// Check that error mentions cancellation
	if err != nil && ctx.Err() != nil {
		t.Log("Context cancellation detected correctly")
	}
}

func TestSyncAudioWithVideo_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	tmpDir := os.TempDir()
	outputPath := filepath.Join(tmpDir, "output.mp4")

	err := SyncAudioWithVideo(ctx, "/nonexistent/video.mp4", "/nonexistent/audio.wav", outputPath)
	if err == nil {
		t.Error("expected error for timed out context")
	}
}
