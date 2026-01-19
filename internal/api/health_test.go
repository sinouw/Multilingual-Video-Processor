package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sinouw/multilingual-video-processor/pkg/models"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", response.Status)
	}

	if response.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", response.Version)
	}

	if response.Timestamp == "" {
		t.Error("expected timestamp to be set")
	}

	// Verify timestamp is valid RFC3339
	if _, err := time.Parse(time.RFC3339, response.Timestamp); err != nil {
		t.Errorf("timestamp is not valid RFC3339: %v", err)
	}
}

func TestReadinessHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	ReadinessHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "ready" {
		t.Errorf("expected status 'ready', got '%s'", response.Status)
	}

	if response.Timestamp == "" {
		t.Error("expected timestamp to be set")
	}
}

func TestLivenessHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()

	LivenessHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "alive" {
		t.Errorf("expected status 'alive', got '%s'", response.Status)
	}

	if response.Timestamp == "" {
		t.Error("expected timestamp to be set")
	}
}

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		requestID  string
	}{
		{
			name:       "Bad Request",
			statusCode: http.StatusBadRequest,
			message:    "invalid request",
			requestID:  "test-request-id",
		},
		{
			name:       "Internal Server Error",
			statusCode: http.StatusInternalServerError,
			message:    "internal error",
			requestID:  "test-request-id-2",
		},
		{
			name:       "Not Found",
			statusCode: http.StatusNotFound,
			message:    "not found",
			requestID:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			ErrorResponse(w, tt.statusCode, tt.message, tt.requestID)

			if w.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, w.Code)
			}

			var response models.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response.Message != tt.message {
				t.Errorf("expected message '%s', got '%s'", tt.message, response.Message)
			}

			if response.RequestID != tt.requestID {
				t.Errorf("expected requestID '%s', got '%s'", tt.requestID, response.RequestID)
			}
		})
	}
}
