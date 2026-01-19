package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/sinouw/multilingual-video-processor/pkg/models"
)

// HealthHandler handles health check requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ReadinessHandler handles readiness probe requests
func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	// Add readiness checks here (e.g., database connection, external service availability)
	response := models.HealthResponse{
		Status:    "ready",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// LivenessHandler handles liveness probe requests
func LivenessHandler(w http.ResponseWriter, r *http.Request) {
	response := models.HealthResponse{
		Status:    "alive",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ErrorResponse sends an error response
func ErrorResponse(w http.ResponseWriter, statusCode int, message string, requestID string) {
	slog.Error("Request error", "statusCode", statusCode, "message", message, "requestID", requestID)

	response := models.ErrorResponse{
		Error:     http.StatusText(statusCode),
		Message:   message,
		RequestID: requestID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode error response", "error", err)
		fmt.Fprintf(w, `{"error":"%s"}`, http.StatusText(statusCode))
	}
}
