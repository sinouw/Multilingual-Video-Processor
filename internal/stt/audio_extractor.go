package stt

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

// ExtractAudioFromVideo extracts audio from video file using FFmpeg
func ExtractAudioFromVideo(ctx context.Context, videoPath string) (string, error) {
	slog.Info("Extracting audio from video", "videoPath", videoPath)

	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("audio extraction cancelled: %w", ctx.Err())
	default:
	}

	// Create temporary audio file
	tmpDir := os.TempDir()
	audioPath := filepath.Join(tmpDir, fmt.Sprintf("audio_%d.wav", os.Getpid()))

	// Use FFmpeg command to extract audio
	// ffmpeg -i input.mp4 -vn -acodec pcm_s16le -ar 16000 -ac 1 output.wav
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", videoPath,
		"-vn",                  // No video
		"-acodec", "pcm_s16le", // Audio codec
		"-ar", "16000", // Sample rate
		"-ac", "1", // Mono
		"-y", // Overwrite output file
		audioPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() != nil {
			return "", fmt.Errorf("audio extraction cancelled: %w", ctx.Err())
		}
		return "", fmt.Errorf("failed to extract audio: %w, stderr: %s", err, stderr.String())
	}

	slog.Info("Audio extracted successfully", "audioPath", audioPath)
	return audioPath, nil
}
