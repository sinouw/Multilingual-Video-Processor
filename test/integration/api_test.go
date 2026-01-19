// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/sinouw/multilingual-video-processor/internal/api"
	"github.com/sinouw/multilingual-video-processor/internal/config"
	"github.com/sinouw/multilingual-video-processor/pkg/models"
)

// TestMain sets up test environment for integration tests
func TestMain(m *testing.M) {
	// Set minimal required environment variables for testing
	os.Setenv("GCS_BUCKET_OUTPUT", "test-bucket")
	os.Setenv("GOOGLE_TRANSLATE_API_KEY", "test-key")
	os.Setenv("RATE_LIMIT_RPM", "100")
	os.Setenv("JOB_TTL", "1h")
	os.Setenv("LOG_LEVEL", "error")

	code := m.Run()
	os.Exit(code)
}

func TestAPI_SubmitTranslationJob(t *testing.T) {
	// Skip if integration tests are not enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=1 to run.")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	jobStore := api.NewInMemoryJobStore(cfg.JobTTL)
	rateLimiter := api.NewRateLimiter(cfg.RateLimitRPM)

	// Create a valid translation request
	request := models.TranslateRequest{
		VideoURL:        "gs://test-bucket/test-video.mp4",
		TargetLanguages: []string{"en"},
	}
	body, _ := json.Marshal(request)

	// Note: This test requires mocking external services or using test credentials
	// For now, we test the request validation and job creation flow
	req := httptest.NewRequest(http.MethodPost, "/v1/translate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Test would require full handler setup with mocked dependencies
	// This is a placeholder for actual integration testing
	t.Log("Integration test placeholder - requires mocked external services")
}

func TestAPI_JobStatusFlow(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=1 to run.")
	}

	cfg, _ := config.LoadConfig()
	jobStore := api.NewInMemoryJobStore(cfg.JobTTL)

	// Create a test job
	jobID := "integration-test-job"
	now := time.Now()
	testStatus := &models.StatusResponse{
		JobID:     jobID,
		Status:    models.StatusProcessing,
		Results:   make(map[string]*models.LanguageResult),
		CreatedAt: &now,
		UpdatedAt: now,
	}
	jobStore.SetStatus(jobID, testStatus)

	// Test status retrieval
	status, err := jobStore.GetStatus(jobID)
	if err != nil {
		t.Fatalf("failed to get job status: %v", err)
	}

	if status.JobID != jobID {
		t.Errorf("expected jobID %s, got %s", jobID, status.JobID)
	}

	if status.Status != models.StatusProcessing {
		t.Errorf("expected status %s, got %s", models.StatusProcessing, status.Status)
	}

	// Test status update
	err = jobStore.UpdateStatusSafely(jobID, func(s *models.StatusResponse) {
		s.Status = models.StatusCompleted
	})
	if err != nil {
		t.Fatalf("failed to update job status: %v", err)
	}

	// Verify update
	updatedStatus, err := jobStore.GetStatus(jobID)
	if err != nil {
		t.Fatalf("failed to get updated job status: %v", err)
	}

	if updatedStatus.Status != models.StatusCompleted {
		t.Errorf("expected status %s, got %s", models.StatusCompleted, updatedStatus.Status)
	}
}

func TestAPI_JobNotFound(t *testing.T) {
	cfg, _ := config.LoadConfig()
	jobStore := api.NewInMemoryJobStore(cfg.JobTTL)

	_, err := jobStore.GetStatus("nonexistent-job")
	if err == nil {
		t.Error("expected error for nonexistent job")
	}

	if _, ok := err.(*api.StatusNotFoundError); !ok {
		t.Errorf("expected StatusNotFoundError, got %T", err)
	}
}
