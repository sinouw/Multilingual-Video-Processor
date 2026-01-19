package models

import "time"

// TranslationStatus represents the status of a translation job
type TranslationStatus string

const (
	StatusIdle       TranslationStatus = "idle"
	StatusProcessing TranslationStatus = "processing"
	StatusCompleted  TranslationStatus = "completed"
	StatusFailed     TranslationStatus = "failed"
)

// TranslateResponse represents the response from the translation API
type TranslateResponse struct {
	JobID   string                     `json:"jobId"`
	Status  TranslationStatus          `json:"status"`
	Results map[string]*LanguageResult `json:"results,omitempty"`
	Error   string                     `json:"error,omitempty"`
}

// LanguageResult represents the result for a single target language
type LanguageResult struct {
	Status         TranslationStatus `json:"status"`
	VideoURL       string            `json:"videoUrl,omitempty"`
	TranslatedText string            `json:"translatedText,omitempty"`
	Progress       int               `json:"progress,omitempty"` // 0-100
	Error          string            `json:"error,omitempty"`
	ProcessedAt    *time.Time        `json:"processedAt,omitempty"`
}

// StatusResponse represents the response from the status endpoint
type StatusResponse struct {
	JobID     string                     `json:"jobId"`
	Status    TranslationStatus          `json:"status"`
	Results   map[string]*LanguageResult `json:"results,omitempty"`
	CreatedAt *time.Time                 `json:"createdAt,omitempty"`
	UpdatedAt time.Time                  `json:"updatedAt,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"requestId,omitempty"`
}
