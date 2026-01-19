package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// GCSStorage implements Storage interface for Google Cloud Storage
type GCSStorage struct {
	client *storage.Client
}

// NewGCSStorage creates a new GCS storage client
func NewGCSStorage(ctx context.Context) (*GCSStorage, error) {
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	var client *storage.Client
	var err error

	if credentialsPath != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
		if err != nil {
			slog.Warn("Failed to create GCS client with credentials file, trying default", "error", err)
			client, err = storage.NewClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to create GCS client: %w", err)
			}
		}
	} else {
		client, err = storage.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCS client: %w", err)
		}
	}

	return &GCSStorage{client: client}, nil
}

// Close closes the storage client
func (s *GCSStorage) Close() error {
	return s.client.Close()
}

// Download downloads a file from GCS and saves it to a temporary local file
// Returns the path to the temporary file
func (s *GCSStorage) Download(ctx context.Context, bucket, path string) (string, error) {
	slog.Info("Downloading from GCS", "bucket", bucket, "path", path)

	obj := s.client.Bucket(bucket).Object(path)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create reader: %w", err)
	}
	defer reader.Close()

	// Create temporary file
	tmpDir := os.TempDir()
	fileName := filepath.Base(path)
	if fileName == "" || fileName == "." {
		fileName = "downloaded_file"
	}
	tmpPath := filepath.Join(tmpDir, fmt.Sprintf("download_%d_%s", os.Getpid(), fileName))

	file, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	// Check context cancellation before copy
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("download cancelled: %w", ctx.Err())
	default:
	}

	// Copy data with context awareness
	// Use io.CopyBuffer for better control and context checking
	copyDone := make(chan error, 1)
	go func() {
		_, err := io.CopyBuffer(file, reader, make([]byte, 32*1024)) // 32KB buffer
		copyDone <- err
	}()

	select {
	case err := <-copyDone:
		if err != nil {
			os.Remove(tmpPath) // Clean up on error
			return "", fmt.Errorf("failed to copy data: %w", err)
		}
	case <-ctx.Done():
		// Context cancelled during copy
		reader.Close() // Close reader to stop copy
		file.Close()
		os.Remove(tmpPath) // Clean up
		return "", fmt.Errorf("download cancelled during copy: %w", ctx.Err())
	}

	// Verify copy completed successfully
	if ctx.Err() != nil {
		return "", fmt.Errorf("download cancelled: %w", ctx.Err())
	}

	slog.Info("Download completed", "localPath", tmpPath)
	return tmpPath, nil
}

// Upload uploads a file from local path to GCS
func (s *GCSStorage) Upload(ctx context.Context, bucket, path string, localPath string) error {
	slog.Info("Uploading to GCS", "bucket", bucket, "path", path, "localPath", localPath)

	// Open local file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Upload to GCS
	obj := s.client.Bucket(bucket).Object(path)
	writer := obj.NewWriter(ctx)
	defer writer.Close()

	// Check context cancellation before copy
	select {
	case <-ctx.Done():
		return fmt.Errorf("upload cancelled: %w", ctx.Err())
	default:
	}

	// Copy data with context awareness
	copyDone := make(chan error, 1)
	go func() {
		_, err := io.CopyBuffer(writer, file, make([]byte, 32*1024)) // 32KB buffer
		copyDone <- err
	}()

	select {
	case err := <-copyDone:
		if err != nil {
			return fmt.Errorf("failed to upload file: %w", err)
		}
	case <-ctx.Done():
		// Context cancelled during copy
		writer.Close() // Close writer to stop copy
		file.Close()
		// Try to delete the object that was being written
		obj := s.client.Bucket(bucket).Object(path)
		obj.Delete(context.Background()) // Use background context for cleanup
		return fmt.Errorf("upload cancelled during copy: %w", ctx.Err())
	}

	// Verify copy completed successfully
	if ctx.Err() != nil {
		return fmt.Errorf("upload cancelled: %w", ctx.Err())
	}

	slog.Info("Upload completed", "bucket", bucket, "path", path)
	return nil
}

// GetPublicURL returns a public URL for a GCS file
func (s *GCSStorage) GetPublicURL(bucket, path string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket, path)
}

// Delete deletes a file from GCS
func (s *GCSStorage) Delete(ctx context.Context, bucket, path string) error {
	slog.Info("Deleting from GCS", "bucket", bucket, "path", path)

	obj := s.client.Bucket(bucket).Object(path)
	err := obj.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	slog.Info("Delete completed", "bucket", bucket, "path", path)
	return nil
}

// Exists checks if a file exists in GCS
func (s *GCSStorage) Exists(ctx context.Context, bucket, path string) (bool, error) {
	obj := s.client.Bucket(bucket).Object(path)
	_, err := obj.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return true, nil
}

// ParseGCSURL parses a GCS URL (gs://bucket/path or https://storage.googleapis.com/bucket/path)
// Returns bucket and path
func ParseGCSURL(url string) (bucket, path string, err error) {
	if strings.HasPrefix(url, "gs://") {
		url = url[5:] // Remove "gs://"
		parts := strings.SplitN(url, "/", 2)
		if len(parts) < 2 {
			return "", "", fmt.Errorf("invalid GCS URL: %s", url)
		}
		return parts[0], parts[1], nil
	}

	if strings.HasPrefix(url, "https://storage.googleapis.com/") {
		url = url[31:] // Remove "https://storage.googleapis.com/" (31 chars)
		parts := strings.SplitN(url, "/", 2)
		if len(parts) < 2 {
			return "", "", fmt.Errorf("invalid GCS URL: %s", url)
		}
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unsupported URL format: %s (expected gs:// or https://)", url)
}
