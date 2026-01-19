package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"github.com/sinouw/multilingual-video-processor/internal/api"
	"github.com/sinouw/multilingual-video-processor/internal/config"
	"github.com/sinouw/multilingual-video-processor/internal/storage"
	stt "github.com/sinouw/multilingual-video-processor/internal/stt"
	"github.com/sinouw/multilingual-video-processor/internal/translation"
	"github.com/sinouw/multilingual-video-processor/internal/tts"
	"github.com/sinouw/multilingual-video-processor/internal/utils"
	"github.com/sinouw/multilingual-video-processor/internal/validator"
	"github.com/sinouw/multilingual-video-processor/internal/video"
	"github.com/sinouw/multilingual-video-processor/pkg/models"
)

var (
	cfg           *config.Config
	storageClient *storage.GCSStorage
	jobStore      *api.InMemoryJobStore
	rateLimiter   *api.RateLimiter
)

func init() {
	var err error

	// Load configuration
	cfg, err = config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Set up logger
	opts := &slog.HandlerOptions{
		Level: cfg.GetLoggerLevel(),
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	// Initialize storage client
	ctx := context.Background()
	storageClient, err = storage.NewGCSStorage(ctx)
	if err != nil {
		slog.Error("Failed to initialize storage client", "error", err)
		os.Exit(1)
	}

	// Initialize job store with TTL
	jobStore = api.NewInMemoryJobStore(cfg.JobTTL)

	// Initialize rate limiter
	rateLimiter = api.NewRateLimiter(cfg.RateLimitRPM)

	slog.Info("Application initialized successfully")
}

// TranslateVideo is the main HTTP handler for video translation
func TranslateVideo(w http.ResponseWriter, r *http.Request) {
	// Handle CORS
	if r.Method == http.MethodOptions {
		handleCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set CORS headers
	handleCORS(w)

	// Route requests
	switch r.URL.Path {
	case "/health":
		api.HealthHandler(w, r)
		return
	case "/health/ready":
		api.ReadinessHandler(w, r)
		return
	case "/health/live":
		api.LivenessHandler(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/v1/status/") {
		api.StatusHandler(jobStore)(w, r)
		return
	}

	if r.URL.Path == "/v1/translate" || r.URL.Path == "/translate" {
		if r.Method == http.MethodPost {
			// Apply rate limiting middleware
			clientIP := api.GetClientIP(r)
			if !rateLimiter.Allow(clientIP) {
				api.ErrorResponse(w, http.StatusTooManyRequests, "rate limit exceeded", "")
				return
			}
			handleTranslate(w, r)
			return
		}
	}

	api.ErrorResponse(w, http.StatusNotFound, "endpoint not found", "")
}

func handleTranslate(w http.ResponseWriter, r *http.Request) {
	requestID := utils.GenerateUUID()

	slog.Info("Translation request received", "requestID", requestID)

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxRequestBodySize)

	// Parse request
	var req models.TranslateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to parse request", "error", err, "requestID", requestID)
		// Check if error is due to size limit
		if err.Error() == "http: request body too large" {
			api.ErrorResponse(w, http.StatusRequestEntityTooLarge, "request body too large", requestID)
		} else {
			api.ErrorResponse(w, http.StatusBadRequest, "invalid request body: "+err.Error(), requestID)
		}
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		slog.Error("Request validation failed", "error", err, "requestID", requestID)
		api.ErrorResponse(w, http.StatusBadRequest, err.Error(), requestID)
		return
	}

	if err := validator.ValidateTranslateRequest(&req, cfg); err != nil {
		slog.Error("Request validation failed", "error", err, "requestID", requestID)
		api.ErrorResponse(w, http.StatusBadRequest, err.Error(), requestID)
		return
	}

	// Generate job ID
	jobID := utils.GenerateUUID()

	// Initialize job status
	now := time.Now()
	jobStatus := &models.StatusResponse{
		JobID:     jobID,
		Status:    models.StatusProcessing,
		Results:   make(map[string]*models.LanguageResult),
		CreatedAt: &now,
		UpdatedAt: now,
	}

	jobStore.SetStatus(jobID, jobStatus)

	// Return immediate response with job ID
	response := models.TranslateResponse{
		JobID:  jobID,
		Status: models.StatusProcessing,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err, "requestID", requestID)
		return
	}

	// Start processing asynchronously (after response is sent)
	// Use background context with timeout since request context will be cancelled after response
	processCtx, processCancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer processCancel()
	go processTranslation(processCtx, jobID, &req, jobStatus)
}

func processTranslation(ctx context.Context, jobID string, req *models.TranslateRequest, jobStatus *models.StatusResponse) {
	slog.Info("Starting translation processing", "jobID", jobID)

	// Track all temporary files for cleanup
	tempFiles := []string{}
	defer func() {
		// Cleanup all temporary files
		for _, file := range tempFiles {
			if file != "" {
				if err := os.Remove(file); err != nil {
					// Log but don't fail if cleanup fails
					slog.Warn("Failed to cleanup temp file", "file", file, "error", err, "jobID", jobID)
				}
			}
		}
		slog.Info("Temp file cleanup completed", "jobID", jobID, "filesCleaned", len(tempFiles))
	}()

	// Check context cancellation
	select {
	case <-ctx.Done():
		updateJobError(jobID, "processing cancelled: "+ctx.Err().Error())
		return
	default:
	}

	// Parse video URL
	bucket, path, err := storage.ParseGCSURL(req.VideoURL)
	if err != nil {
		updateJobError(jobID, "failed to parse video URL: "+err.Error())
		return
	}

	// Download video
	slog.Info("Downloading video", "jobID", jobID, "bucket", bucket, "path", path)
	videoPath, err := storageClient.Download(ctx, bucket, path)
	if err != nil {
		if ctx.Err() != nil {
			updateJobError(jobID, "processing cancelled during download: "+ctx.Err().Error())
		} else {
			updateJobError(jobID, "failed to download video: "+err.Error())
		}
		return
	}
	tempFiles = append(tempFiles, videoPath)

	// Check context cancellation
	select {
	case <-ctx.Done():
		updateJobError(jobID, "processing cancelled: "+ctx.Err().Error())
		return
	default:
	}

	// Get video duration
	videoDuration, err := video.GetVideoDuration(ctx, videoPath)
	if err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() != nil {
			updateJobError(jobID, "processing cancelled during duration check: "+ctx.Err().Error())
		} else {
			updateJobError(jobID, "failed to get video duration: "+err.Error())
		}
		return
	}

	// Validate video duration
	if videoDuration > cfg.MaxVideoDuration.Seconds() {
		updateJobError(jobID, fmt.Sprintf("video duration exceeds maximum: %.2fs > %.2fs", videoDuration, cfg.MaxVideoDuration.Seconds()))
		return
	}

	// Extract audio
	slog.Info("Extracting audio", "jobID", jobID)
	audioPath, err := stt.ExtractAudioFromVideo(ctx, videoPath)
	if err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() != nil {
			updateJobError(jobID, "processing cancelled during audio extraction: "+ctx.Err().Error())
		} else {
			updateJobError(jobID, "failed to extract audio: "+err.Error())
		}
		return
	}
	tempFiles = append(tempFiles, audioPath)

	// Check context cancellation
	select {
	case <-ctx.Done():
		updateJobError(jobID, "processing cancelled: "+ctx.Err().Error())
		return
	default:
	}

	// Transcribe audio
	slog.Info("Transcribing audio", "jobID", jobID)
	transcription, err := stt.SpeechToText(ctx, audioPath, req.SourceLanguage)
	if err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() != nil {
			updateJobError(jobID, "transcription cancelled: "+ctx.Err().Error())
		} else {
			updateJobError(jobID, "failed to transcribe audio: "+err.Error())
		}
		return
	}

	originalText := transcription.Text
	sourceLanguage := transcription.Language
	if sourceLanguage == "" {
		sourceLanguage = req.SourceLanguage
		if sourceLanguage == "" {
			sourceLanguage = "auto"
		}
	}

	// Validate transcription result
	if originalText == "" {
		updateJobError(jobID, "transcription returned empty text")
		return
	}

	slog.Info("Transcription completed", "jobID", jobID, "textLength", len(originalText), "language", sourceLanguage)

	// Check context cancellation before starting language processing
	select {
	case <-ctx.Done():
		updateJobError(jobID, "processing cancelled: "+ctx.Err().Error())
		return
	default:
	}

	// Process each target language concurrently
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, cfg.MaxConcurrentTranslations)

	for _, targetLang := range req.TargetLanguages {
		// Check context cancellation before processing each language
		select {
		case <-ctx.Done():
			slog.Warn("Processing cancelled, stopping language processing", "jobID", jobID)
			// Mark remaining languages as failed
			for _, lang := range req.TargetLanguages {
				if _, exists := jobStatus.Results[lang]; !exists {
					jobStore.UpdateStatusSafely(jobID, func(status *models.StatusResponse) {
						if status.Results == nil {
							status.Results = make(map[string]*models.LanguageResult)
						}
						status.Results[lang] = &models.LanguageResult{
							Status: models.StatusFailed,
							Error:  "processing cancelled",
						}
						status.UpdatedAt = time.Now()
					})
				}
			}
			updateJobError(jobID, "processing cancelled: "+ctx.Err().Error())
			return
		default:
		}

		wg.Add(1)
		go func(lang string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			result := processLanguage(ctx, jobID, originalText, sourceLanguage, lang, videoPath, videoDuration, cfg.GCSOutputBucket)

			// Thread-safe update using UpdateStatusSafely
			jobStore.UpdateStatusSafely(jobID, func(status *models.StatusResponse) {
				if status.Results == nil {
					status.Results = make(map[string]*models.LanguageResult)
				}
				status.Results[lang] = result
				status.UpdatedAt = time.Now()
			})
		}(targetLang)
	}

	// Wait for all goroutines with context cancellation support
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines completed
	case <-ctx.Done():
		slog.Warn("Context cancelled while waiting for goroutines", "jobID", jobID)
		// Context is cancelled, but goroutines should finish quickly
		// Wait with timeout for goroutines to clean up
		timeout := time.NewTimer(5 * time.Second)
		defer timeout.Stop()
		select {
		case <-done:
			// Goroutines finished quickly
		case <-timeout.C:
			slog.Warn("Goroutines did not complete within timeout after cancellation", "jobID", jobID)
		}
	}

	// Check context cancellation after all languages processed
	select {
	case <-ctx.Done():
		updateJobError(jobID, "processing cancelled: "+ctx.Err().Error())
		return
	default:
	}

	// Update final status using thread-safe update
	var finalStatus models.TranslationStatus
	jobStore.UpdateStatusSafely(jobID, func(status *models.StatusResponse) {
		allCompleted := true
		anyFailed := false
		for _, result := range status.Results {
			if result.Status != models.StatusCompleted {
				allCompleted = false
				if result.Status == models.StatusFailed {
					anyFailed = true
				}
			}
		}

		if allCompleted {
			status.Status = models.StatusCompleted
			finalStatus = models.StatusCompleted
		} else if anyFailed {
			status.Status = models.StatusFailed
			finalStatus = models.StatusFailed
		}
		status.UpdatedAt = time.Now()
	})

	slog.Info("Translation processing completed", "jobID", jobID, "status", finalStatus)

	// Send webhook notification if configured
	if cfg.WebhookURL != "" {
		go func() {
			status, err := jobStore.GetStatus(jobID)
			if err == nil && status != nil {
				// Use background context for webhook since main context may be cancelled
				webhookCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if err := api.NotifyWebhook(webhookCtx, cfg.WebhookURL, status); err != nil {
					slog.Warn("Webhook notification failed", "error", err, "jobID", jobID)
					// Don't fail the job if webhook fails
				}
			}
		}()
	}
}

func processLanguage(ctx context.Context, jobID string, originalText string, sourceLanguage string, targetLanguage string, videoPath string, videoDuration float64, outputBucket string) *models.LanguageResult {
	result := &models.LanguageResult{
		Status:   models.StatusProcessing,
		Progress: 0,
	}

	slog.Info("Processing language", "jobID", jobID, "targetLanguage", targetLanguage)

	// Check context cancellation before translation
	select {
	case <-ctx.Done():
		result.Status = models.StatusFailed
		result.Error = "processing cancelled: " + ctx.Err().Error()
		result.Progress = 0
		return result
	default:
	}

	// Translate text
	result.Progress = 20
	translatedText, err := translation.TranslateText(ctx, originalText, sourceLanguage, targetLanguage)
	if err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() != nil {
			result.Status = models.StatusFailed
			result.Error = "translation cancelled: " + ctx.Err().Error()
		} else {
			result.Status = models.StatusFailed
			result.Error = "translation failed: " + err.Error()
		}
		result.Progress = 0
		slog.Error("Translation failed", "jobID", jobID, "targetLanguage", targetLanguage, "error", err)
		return result
	}

	result.Progress = 40

	// Check context cancellation before TTS generation
	select {
	case <-ctx.Done():
		result.Status = models.StatusFailed
		result.Error = "processing cancelled: " + ctx.Err().Error()
		result.Progress = 0
		return result
	default:
	}

	// Generate TTS audio
	audioPath, err := createTempFile(fmt.Sprintf("audio_%s_%s.mp3", jobID, targetLanguage))
	if err != nil {
		result.Status = models.StatusFailed
		result.Error = "failed to create temp file: " + err.Error()
		result.Progress = 0
		return result
	}
	defer os.Remove(audioPath)

	err = tts.GenerateTTS(ctx, translatedText, targetLanguage, videoDuration, audioPath)
	if err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() != nil {
			result.Status = models.StatusFailed
			result.Error = "TTS generation cancelled: " + ctx.Err().Error()
		} else {
			result.Status = models.StatusFailed
			result.Error = "TTS generation failed: " + err.Error()
		}
		result.Progress = 0
		return result
	}

	result.Progress = 60

	// Check context cancellation before audio sync
	select {
	case <-ctx.Done():
		result.Status = models.StatusFailed
		result.Error = "processing cancelled: " + ctx.Err().Error()
		result.Progress = 0
		return result
	default:
	}

	// Sync audio with video
	outputVideoPath, err := createTempFile(fmt.Sprintf("video_%s_%s.mp4", jobID, targetLanguage))
	if err != nil {
		result.Status = models.StatusFailed
		result.Error = "failed to create temp file: " + err.Error()
		result.Progress = 0
		return result
	}
	defer os.Remove(outputVideoPath)

	err = video.SyncAudioWithVideo(ctx, videoPath, audioPath, outputVideoPath)
	if err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() != nil {
			result.Status = models.StatusFailed
			result.Error = "audio sync cancelled: " + ctx.Err().Error()
		} else {
			result.Status = models.StatusFailed
			result.Error = "audio sync failed: " + err.Error()
		}
		result.Progress = 0
		return result
	}

	result.Progress = 80

	// Upload to GCS
	outputPath := fmt.Sprintf("translations/%s/%s.mp4", jobID, targetLanguage)
	err = storageClient.Upload(ctx, outputBucket, outputPath, outputVideoPath)
	if err != nil {
		result.Status = models.StatusFailed
		result.Error = "upload failed: " + err.Error()
		result.Progress = 0
		return result
	}

	result.Progress = 100
	result.Status = models.StatusCompleted
	result.VideoURL = storageClient.GetPublicURL(outputBucket, outputPath)
	result.TranslatedText = translatedText
	now := time.Now()
	result.ProcessedAt = &now

	slog.Info("Language processing completed", "jobID", jobID, "targetLanguage", targetLanguage)
	return result
}

func updateJobError(jobID string, errorMsg string) {
	jobStore.UpdateStatusSafely(jobID, func(status *models.StatusResponse) {
		status.Status = models.StatusFailed
		status.UpdatedAt = time.Now()
		// Add error to the first language result or create a generic error
		if len(status.Results) == 0 {
			status.Results = make(map[string]*models.LanguageResult)
			status.Results["error"] = &models.LanguageResult{
				Status: models.StatusFailed,
				Error:  errorMsg,
			}
		}
	})
	slog.Error("Job failed", "jobID", jobID, "error", errorMsg)

	// Send webhook notification if configured
	if cfg.WebhookURL != "" {
		go func() {
			status, err := jobStore.GetStatus(jobID)
			if err == nil && status != nil {
				// Use background context for webhook since main context may be cancelled
				webhookCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if err := api.NotifyWebhook(webhookCtx, cfg.WebhookURL, status); err != nil {
					slog.Warn("Webhook notification failed", "error", err, "jobID", jobID)
					// Don't fail the job if webhook fails
				}
			}
		}()
	}
}

func createTempFile(pattern string) (string, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	path := file.Name()
	file.Close()
	return path, nil
}

func handleCORS(w http.ResponseWriter) {
	origins := cfg.CORSOrigins
	if len(origins) == 0 || (len(origins) == 1 && origins[0] == "*") {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		// In production, validate against allowed origins
		w.Header().Set("Access-Control-Allow-Origin", origins[0])
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "3600")
}

func main() {
	// Register HTTP function
	funcframework.RegisterHTTPFunction("/", TranslateVideo)

	// Use PORT environment variable, or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	if err := funcframework.Start(port); err != nil {
		slog.Error("Failed to start function", "error", err)
		os.Exit(1)
	}
}
