package logs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

func TestProvider_Description(t *testing.T) {
	p := NewProvider(".")
	desc := p.Description()

	if desc.Name != "logs" {
		t.Errorf("expected Name 'logs', got '%s'", desc.Name)
	}
	if desc.Type != ports.ContextTypeLogs {
		t.Errorf("expected Type ContextTypeLogs, got '%s'", desc.Type)
	}
}

func TestProvider_GetContextItems_LogFile(t *testing.T) {
	// Create temp log file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	content := `2026-01-16 10:00:00 INFO Starting application
2026-01-16 10:00:01 INFO Loading config
2026-01-16 10:00:02 ERROR Failed to connect to database
2026-01-16 10:00:03 WARN Retrying connection
2026-01-16 10:00:04 INFO Connection established
2026-01-16 10:00:05 ERROR timeout waiting for response`
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewProvider(tmpDir)
	items, err := p.GetContextItems(context.Background(), "test.log", ports.ContextExtras{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	// Should detect ERROR lines
	if !strings.Contains(items[0].Content, "DETECTED ISSUES") {
		t.Error("expected to detect issues in log")
	}
	if !strings.Contains(items[0].Content, "ERROR Failed to connect") {
		t.Error("expected to find database error")
	}
}

func TestProvider_GetContextItems_LastNLines(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Create log with 10 lines
	var lines []string
	for i := 1; i <= 10; i++ {
		lines = append(lines, "Line number "+string(rune('0'+i)))
	}
	if err := os.WriteFile(logFile, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewProvider(tmpDir)
	items, err := p.GetContextItems(context.Background(), "test.log:5", ports.ContextExtras{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	// Should only show last 5 lines description
	if !strings.Contains(items[0].Description, "5 lines") {
		t.Errorf("expected description to mention 5 lines, got: %s", items[0].Description)
	}
}

func TestProvider_GetContextItems_FileNotFound(t *testing.T) {
	p := NewProvider(".")
	_, err := p.GetContextItems(context.Background(), "nonexistent.log", ports.ContextExtras{})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
