package file

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

	if desc.Name != "file" {
		t.Errorf("expected Name 'file', got '%s'", desc.Name)
	}
	if desc.Type != ports.ContextTypeFile {
		t.Errorf("expected Type ContextTypeFile, got '%s'", desc.Type)
	}
}

func TestProvider_GetContextItems_File(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewProvider(tmpDir)
	items, err := p.GetContextItems(context.Background(), "test.txt", ports.ContextExtras{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if !strings.Contains(items[0].Content, "line1") {
		t.Error("expected content to contain 'line1'")
	}
	if !strings.Contains(items[0].Content, "line3") {
		t.Error("expected content to contain 'line3'")
	}
}

func TestProvider_GetContextItems_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(""), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	p := NewProvider(tmpDir)
	items, err := p.GetContextItems(context.Background(), ".", ports.ContextExtras{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if !strings.Contains(items[0].Content, "file1.txt") {
		t.Error("expected content to contain 'file1.txt'")
	}
	if !strings.Contains(items[0].Content, "subdir") {
		t.Error("expected content to contain 'subdir'")
	}
}

func TestProvider_GetContextItems_NotFound(t *testing.T) {
	p := NewProvider(".")
	_, err := p.GetContextItems(context.Background(), "nonexistent-file-12345.txt", ports.ContextExtras{})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
