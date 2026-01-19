package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/sinouw/multilingual-video-processor/pkg/models"
)

// WebhookPayload represents the payload sent to webhook URL
type WebhookPayload struct {
	Event     string                            `json:"event"`
	JobID     string                            `json:"jobId"`
	Status    models.TranslationStatus          `json:"status"`
	Results   map[string]*models.LanguageResult `json:"results,omitempty"`
	Timestamp string                            `json:"timestamp"`
	Error     string                            `json:"error,omitempty"`
}

// NotifyWebhook sends a webhook notification with job status
// This function is non-blocking and handles errors gracefully
func NotifyWebhook(ctx context.Context, webhookURL string, jobStatus *models.StatusResponse) error {
	if webhookURL == "" {
		return nil // No webhook configured, skip
	}

	// Determine event type based on status
	event := "job.completed"
	if jobStatus.Status == models.StatusFailed {
		event = "job.failed"
	} else if jobStatus.Status == models.StatusProcessing {
		event = "job.processing"
	}

	payload := WebhookPayload{
		Event:     event,
		JobID:     jobStatus.JobID,
		Status:    jobStatus.Status,
		Results:   jobStatus.Results,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Add error message if failed
	if jobStatus.Status == models.StatusFailed {
		// Try to extract error from results
		for _, result := range jobStatus.Results {
			if result.Error != "" {
				payload.Error = result.Error
				break
			}
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal webhook payload", "error", err, "jobID", jobStatus.JobID)
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Create HTTP request with timeout
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("Failed to create webhook request", "error", err, "jobID", jobStatus.JobID)
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "multilingual-video-processor/1.0")

	// Send webhook with retry logic
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	maxRetries := 2
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
				continue
			}
			slog.Warn("Failed to send webhook after retries", "error", err, "jobID", jobStatus.JobID, "attempt", attempt+1)
			return fmt.Errorf("failed to send webhook after %d attempts: %w", maxRetries, err)
		}

		resp.Body.Close()

		// Check status code
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			slog.Info("Webhook notification sent successfully", "jobID", jobStatus.JobID, "statusCode", resp.StatusCode)
			return nil
		}

		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		slog.Warn("Webhook returned non-2xx status after retries", "statusCode", resp.StatusCode, "jobID", jobStatus.JobID, "attempt", attempt+1)
	}

	// Log but don't fail the job if webhook fails
	slog.Error("Webhook notification failed", "error", lastErr, "jobID", jobStatus.JobID)
	return lastErr
}
