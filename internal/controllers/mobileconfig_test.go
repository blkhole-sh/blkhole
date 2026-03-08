package controllers

import (
	"regexp"
	"testing"
)

func TestGenerateUUID(t *testing.T) {
	// regex for UUID v4
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	uuid, err := generateUUID()
	if err != nil {
		t.Fatalf("generateUUID() returned error: %v", err)
	}

	if !uuidRegex.MatchString(uuid) {
		t.Errorf("generateUUID() returned invalid UUID format: %s", uuid)
	}
}
