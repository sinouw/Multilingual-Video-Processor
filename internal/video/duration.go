package video

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
)

// GetVideoDuration gets the duration of a video file using ffprobe
func GetVideoDuration(ctx context.Context, videoPath string) (float64, error) {
	slog.Debug("Getting video duration", "videoPath", videoPath)

	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("video duration check cancelled: %w", ctx.Err())
	default:
	}

	// Use ffprobe to get video duration
	// ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 video.mp4
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() != nil {
			return 0, fmt.Errorf("video duration check cancelled: %w", ctx.Err())
		}
		return 0, fmt.Errorf("failed to get video duration: %w, stderr: %s", err, stderr.String())
	}

	durationStr := strings.TrimSpace(stdout.String())
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse video duration: %w", err)
	}

	slog.Debug("Video duration retrieved", "duration", duration)
	return duration, nil
}

// GetAudioDuration gets the duration of an audio file using ffprobe
func GetAudioDuration(ctx context.Context, audioPath string) (float64, error) {
	slog.Debug("Getting audio duration", "audioPath", audioPath)

	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("audio duration check cancelled: %w", ctx.Err())
	default:
	}

	// Use ffprobe to get audio duration
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioPath,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() != nil {
			return 0, fmt.Errorf("audio duration check cancelled: %w", ctx.Err())
		}
		return 0, fmt.Errorf("failed to get audio duration: %w, stderr: %s", err, stderr.String())
	}

	durationStr := strings.TrimSpace(stdout.String())
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse audio duration: %w", err)
	}

	slog.Debug("Audio duration retrieved", "duration", duration)
	return duration, nil
}
