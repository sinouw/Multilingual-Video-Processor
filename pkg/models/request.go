package models

// TranslateRequest represents the request body for video translation
type TranslateRequest struct {
	VideoURL        string   `json:"videoUrl"`                 // GCS URL or HTTPS URL of the video
	TargetLanguages []string `json:"targetLanguages"`          // Languages to translate to (e.g., ["en", "ar", "de"])
	SourceLanguage  string   `json:"sourceLanguage,omitempty"` // Optional source language hint (empty for auto-detect)
}

// Validate performs basic validation on the request
func (r *TranslateRequest) Validate() error {
	if r.VideoURL == "" {
		return ErrMissingVideoURL
	}

	if len(r.TargetLanguages) == 0 {
		return ErrMissingTargetLanguages
	}

	return nil
}

// Common validation errors
var (
	ErrMissingVideoURL        = &ValidationError{Message: "videoUrl is required"}
	ErrMissingTargetLanguages = &ValidationError{Message: "at least one target language is required"}
)

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
