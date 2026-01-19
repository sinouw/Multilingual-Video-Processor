package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/sinouw/multilingual-video-processor/pkg/models"
)

// JobStatusStore defines the interface for storing and retrieving job status
type JobStatusStore interface {
	GetStatus(jobID string) (*models.StatusResponse, error)
	SetStatus(jobID string, status *models.StatusResponse)
	UpdateStatusSafely(jobID string, updater func(*models.StatusResponse)) error
}

// StatusHandler handles job status requests
func StatusHandler(store JobStatusStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Extract job ID from path
		jobID := r.URL.Path[len("/v1/status/"):]
		if jobID == "" {
			ErrorResponse(w, http.StatusBadRequest, "job ID is required", "")
			return
		}

		slog.Info("Status request", "jobID", jobID)

		status, err := store.GetStatus(jobID)
		if err != nil {
			slog.Error("Failed to get job status", "error", err, "jobID", jobID)
			ErrorResponse(w, http.StatusNotFound, "job not found", jobID)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(status)
	}
}

// In-memory job store (for single-instance deployments)
// In production, use a persistent store like Redis, Firestore, or Cloud SQL
type InMemoryJobStore struct {
	mu     sync.RWMutex
	jobs   map[string]*jobEntry
	jobTTL time.Duration
}

// jobEntry wraps a job status with metadata
type jobEntry struct {
	status    *models.StatusResponse
	createdAt time.Time
}

// NewInMemoryJobStore creates a new in-memory job store
func NewInMemoryJobStore(jobTTL time.Duration) *InMemoryJobStore {
	store := &InMemoryJobStore{
		jobs:   make(map[string]*jobEntry),
		jobTTL: jobTTL,
	}
	// Start cleanup goroutine
	go store.startCleanup()
	return store
}

// SetStatus sets the status for a job (thread-safe)
func (s *InMemoryJobStore) SetStatus(jobID string, status *models.StatusResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	// Set CreatedAt if not already set
	if status.CreatedAt == nil {
		status.CreatedAt = &now
	}

	s.jobs[jobID] = &jobEntry{
		status:    status,
		createdAt: now,
	}
}

// GetStatus retrieves the status for a job (thread-safe)
func (s *InMemoryJobStore) GetStatus(jobID string) (*models.StatusResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.jobs[jobID]
	if !exists {
		return nil, &StatusNotFoundError{JobID: jobID}
	}

	// Check if job has expired
	if s.jobTTL > 0 && time.Since(entry.createdAt) > s.jobTTL {
		return nil, &StatusNotFoundError{JobID: jobID}
	}

	return entry.status, nil
}

// UpdateStatusSafely updates a job status using an updater function (thread-safe)
func (s *InMemoryJobStore) UpdateStatusSafely(jobID string, updater func(*models.StatusResponse)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.jobs[jobID]
	if !exists {
		return &StatusNotFoundError{JobID: jobID}
	}

	// Check if job has expired
	if s.jobTTL > 0 && time.Since(entry.createdAt) > s.jobTTL {
		return &StatusNotFoundError{JobID: jobID}
	}

	// Apply updater function
	updater(entry.status)
	entry.status.UpdatedAt = time.Now()

	return nil
}

// CleanupExpiredJobs removes expired jobs from the store
func (s *InMemoryJobStore) CleanupExpiredJobs() {
	if s.jobTTL <= 0 {
		return // No TTL, skip cleanup
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for jobID, entry := range s.jobs {
		if now.Sub(entry.createdAt) > s.jobTTL {
			delete(s.jobs, jobID)
			slog.Info("Removed expired job", "jobID", jobID, "age", now.Sub(entry.createdAt))
		}
	}
}

// startCleanup starts a background goroutine that periodically cleans up expired jobs
func (s *InMemoryJobStore) startCleanup() {
	if s.jobTTL <= 0 {
		return // No TTL, skip cleanup
	}

	ticker := time.NewTicker(s.jobTTL / 2) // Clean up every half TTL period
	defer ticker.Stop()

	for range ticker.C {
		s.CleanupExpiredJobs()
	}
}

// StatusNotFoundError represents a job not found error
type StatusNotFoundError struct {
	JobID string
}

func (e *StatusNotFoundError) Error() string {
	return "job not found: " + e.JobID
}
