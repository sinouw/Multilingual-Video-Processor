package storage

import (
	"context"
	"io"
)

// Storage defines the interface for storage operations
type Storage interface {
	// Download downloads a file from storage and returns a reader
	Download(ctx context.Context, bucket, path string) (io.ReadCloser, error)

	// Upload uploads a file to storage
	Upload(ctx context.Context, bucket, path string, reader io.Reader) error

	// GetPublicURL returns a public URL for a stored file
	GetPublicURL(bucket, path string) string

	// Delete deletes a file from storage
	Delete(ctx context.Context, bucket, path string) error

	// Exists checks if a file exists in storage
	Exists(ctx context.Context, bucket, path string) (bool, error)
}
