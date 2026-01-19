package test

import (
	"os"
	"path/filepath"
	"testing"
)

// SetupTestEnv sets up test environment variables
func SetupTestEnv(t *testing.T) {
	t.Helper()
	// Set test environment variables if needed
}

// CleanupTestFiles removes test files
func CleanupTestFiles(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if path != "" {
			_ = os.Remove(path)
		}
	}
}

// GetTestFixturesDir returns the path to test fixtures directory
func GetTestFixturesDir() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "fixtures")
}
