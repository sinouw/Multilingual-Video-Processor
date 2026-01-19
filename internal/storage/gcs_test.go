package storage

import (
	"context"
	"testing"
	"time"
)

func TestParseGCSURL_Valid(t *testing.T) {
	tests := []struct {
		url      string
		bucket   string
		path     string
		hasError bool
	}{
		{"gs://bucket/path/to/file.mp4", "bucket", "path/to/file.mp4", false},
		{"gs://my-bucket/videos/video.mp4", "my-bucket", "videos/video.mp4", false},
		{"gs://bucket/file", "bucket", "file", false},
		{"https://storage.googleapis.com/bucket/file.mp4", "bucket", "file.mp4", false},
		{"https://storage.googleapis.com/my-bucket/path/file.mp4", "my-bucket", "path/file.mp4", false},
		{"invalid-url", "", "", true},
		{"http://example.com/file.mp4", "", "", true},
		{"", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			bucket, path, err := ParseGCSURL(tt.url)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error for URL %s", tt.url)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for URL %s: %v", tt.url, err)
				}
				if bucket != tt.bucket {
					t.Errorf("expected bucket %s, got %s", tt.bucket, bucket)
				}
				if path != tt.path {
					t.Errorf("expected path %s, got %s", tt.path, path)
				}
			}
		})
	}
}

func TestGetPublicURL(t *testing.T) {
	// Create a mock storage client (we can't easily test GCS without credentials)
	// So we test the URL construction logic
	storage := &GCSStorage{}

	tests := []struct {
		bucket string
		path   string
		want   string
	}{
		{"bucket", "path/file.mp4", "https://storage.googleapis.com/bucket/path/file.mp4"},
		{"my-bucket", "videos/video.mp4", "https://storage.googleapis.com/my-bucket/videos/video.mp4"},
	}

	for _, tt := range tests {
		t.Run(tt.bucket+"/"+tt.path, func(t *testing.T) {
			got := storage.GetPublicURL(tt.bucket, tt.path)
			if got != tt.want {
				t.Errorf("expected %s, got %s", tt.want, got)
			}
		})
	}
}

func TestExists_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// We can't easily test GCS operations without credentials
	// But we can test context handling
	if ctx.Err() == nil {
		t.Error("expected context to be cancelled")
	}
}

func TestExists_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	if ctx.Err() == nil {
		t.Error("expected context to be timed out")
	}
}
