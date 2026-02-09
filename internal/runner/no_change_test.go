package runner

import (
	"regexp"
	"strings"
	"testing"
)

func TestDryRun_NoChanges(t *testing.T) {
	prefix := regexp.MustCompile(`\s*[#;/]* selfup `)

	t.Run("Empty input", func(t *testing.T) {
		input := ``
		result, err := DryRun(strings.NewReader(input), prefix, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ChangedCount != 0 {
			t.Errorf("expected 0 changes for empty input, got %d", result.ChangedCount)
		}
		if result.Total != 0 {
			t.Errorf("expected 0 total definitions, got %d", result.Total)
		}
	})

	t.Run("No selfup definitions", func(t *testing.T) {
		input := "line1\nline2"
		result, err := DryRun(strings.NewReader(input), prefix, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ChangedCount != 0 {
			t.Errorf("expected 0 changes, got %d", result.ChangedCount)
		}
		if result.Total != 0 {
			t.Errorf("expected 0 total definitions, got %d", result.Total)
		}
	})

	t.Run("Selfup definition exists but no change needed", func(t *testing.T) {
		input := `tool: 1.0 # selfup { "extract": "[0-9.]+", "replacer": ["echo", "1.0"] }`
		result, err := DryRun(strings.NewReader(input), prefix, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ChangedCount != 0 {
			t.Errorf("expected 0 changes, got %d", result.ChangedCount)
		}
		if result.Total != 1 {
			t.Errorf("expected 1 total definition, got %d", result.Total)
		}
	})
}
