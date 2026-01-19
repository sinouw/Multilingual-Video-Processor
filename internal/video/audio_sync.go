package video

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

// SyncAudioWithVideo replaces audio track in video with new TTS audio
func SyncAudioWithVideo(ctx context.Context, videoPath string, audioPath string, outputPath string) error {
	slog.Info("Synchronizing audio with video",
		"videoPath", videoPath,
		"audioPath", audioPath,
		"outputPath", outputPath)

	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return fmt.Errorf("audio sync cancelled: %w", ctx.Err())
	default:
	}

	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get video duration for logging
	videoDuration, err := GetVideoDuration(ctx, videoPath)
	if err != nil {
		slog.Warn("Failed to get video duration", "error", err)
	}

	// Get audio duration for logging
	audioDuration, err := GetAudioDuration(ctx, audioPath)
	if err != nil {
		slog.Warn("Failed to get audio duration", "error", err)
	}

	slog.Debug("Durations",
		"videoDuration", videoDuration,
		"audioDuration", audioDuration)

	// Check context again before FFmpeg operation
	select {
	case <-ctx.Done():
		return fmt.Errorf("audio sync cancelled: %w", ctx.Err())
	default:
	}

	// Use FFmpeg to replace audio track
	// ffmpeg -i video.mp4 -i audio.wav -c:v copy -c:a aac -map 0:v:0 -map 1:a:0 -shortest output.mp4
	// -shortest will trim to shortest stream (video or audio)
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy", // Copy video codec (no re-encoding)
		"-c:a", "aac", // Audio codec
		"-map", "0:v:0", // Map video from first input
		"-map", "1:a:0", // Map audio from second input
		"-shortest", // Finish encoding when the shortest input stream ends
		"-y",        // Overwrite output file
		outputPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() != nil {
			return fmt.Errorf("audio sync cancelled: %w", ctx.Err())
		}
		return fmt.Errorf("failed to sync audio with video: %w, stderr: %s", err, stderr.String())
	}

	slog.Info("Audio-video synchronization completed", "outputPath", outputPath)
	return nil
}
