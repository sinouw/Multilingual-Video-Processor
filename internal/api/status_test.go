package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sinouw/multilingual-video-processor/pkg/models"
)

// mockJobStore is a mock implementation of JobStatusStore for testing
type mockJobStore struct {
	jobs map[string]*models.StatusResponse
}

func newMockJobStore() *mockJobStore {
	return &mockJobStore{
		jobs: make(map[string]*models.StatusResponse),
	}
}

func (m *mockJobStore) GetStatus(jobID string) (*models.StatusResponse, error) {
	status, exists := m.jobs[jobID]
	if !exists {
		return nil, &StatusNotFoundError{JobID: jobID}
	}
	return status, nil
}

func (m *mockJobStore) SetStatus(jobID string, status *models.StatusResponse) {
	m.jobs[jobID] = status
}

func (m *mockJobStore) UpdateStatusSafely(jobID string, updater func(*models.StatusResponse)) error {
	status, exists := m.jobs[jobID]
	if !exists {
		return &StatusNotFoundError{JobID: jobID}
	}
	updater(status)
	return nil
}

func TestStatusHandler_Get(t *testing.T) {
	store := newMockJobStore()
	handler := StatusHandler(store)

	// Create a test job
	jobID := "test-job-123"
	now := time.Now()
	testStatus := &models.StatusResponse{
		JobID:     jobID,
		Status:    models.StatusProcessing,
		Results:   make(map[string]*models.LanguageResult),
		CreatedAt: &now,
		UpdatedAt: now,
	}
	store.SetStatus(jobID, testStatus)

	req := httptest.NewRequest(http.MethodGet, "/v1/status/"+jobID, nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.StatusResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.JobID != jobID {
		t.Errorf("expected jobID '%s', got '%s'", jobID, response.JobID)
	}

	if response.Status != models.StatusProcessing {
		t.Errorf("expected status '%s', got '%s'", models.StatusProcessing, response.Status)
	}
}

func TestStatusHandler_NotFound(t *testing.T) {
	store := newMockJobStore()
	handler := StatusHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/v1/status/nonexistent-job", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var response models.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Message != "job not found" {
		t.Errorf("expected message 'job not found', got '%s'", response.Message)
	}
}

func TestStatusHandler_EmptyJobID(t *testing.T) {
	store := newMockJobStore()
	handler := StatusHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/v1/status/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestStatusHandler_MethodNotAllowed(t *testing.T) {
	store := newMockJobStore()
	handler := StatusHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/v1/status/test-job", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestStatusHandler_CompletedJob(t *testing.T) {
	store := newMockJobStore()
	handler := StatusHandler(store)

	jobID := "completed-job-123"
	now := time.Now()
	testStatus := &models.StatusResponse{
		JobID:  jobID,
		Status: models.StatusCompleted,
		Results: map[string]*models.LanguageResult{
			"en": {
				Status:        models.StatusCompleted,
				VideoURL:      "gs://bucket/translations/job-id/en.mp4",
				TranslatedText: "Hello, world!",
				Progress:      100,
				ProcessedAt:   &now,
			},
		},
		CreatedAt: &now,
		UpdatedAt: now,
	}
	store.SetStatus(jobID, testStatus)

	req := httptest.NewRequest(http.MethodGet, "/v1/status/"+jobID, nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.StatusResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != models.StatusCompleted {
		t.Errorf("expected status '%s', got '%s'", models.StatusCompleted, response.Status)
	}

	if len(response.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(response.Results))
	}

	if enResult, exists := response.Results["en"]; exists {
		if enResult.Progress != 100 {
			t.Errorf("expected progress 100, got %d", enResult.Progress)
		}
	} else {
		t.Error("expected 'en' result to exist")
	}
}
