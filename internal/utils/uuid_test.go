package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestGenerateUUID(t *testing.T) {
	uuidStr := GenerateUUID()

	// Verify it's a valid UUID
	_, err := uuid.Parse(uuidStr)
	if err != nil {
		t.Errorf("GenerateUUID() returned invalid UUID: %v", err)
	}

	// Verify it's not empty
	if uuidStr == "" {
		t.Error("GenerateUUID() returned empty string")
	}
}

func TestGenerateUUIDUniqueness(t *testing.T) {
	uuid1 := GenerateUUID()
	uuid2 := GenerateUUID()

	if uuid1 == uuid2 {
		t.Error("GenerateUUID() returned duplicate UUIDs")
	}
}
