package system

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

func TestProvider_Description(t *testing.T) {
	p := NewProvider()
	desc := p.Description()

	if desc.Name != "system" {
		t.Errorf("expected Name 'system', got '%s'", desc.Name)
	}
	if desc.Type != ports.ContextTypeSystem {
		t.Errorf("expected Type ContextTypeSystem, got '%s'", desc.Type)
	}
}

func TestProvider_GetContextItems_OS(t *testing.T) {
	p := NewProvider()
	items, err := p.GetContextItems(context.Background(), "os", ports.ContextExtras{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	// Should contain OS info
	if !strings.Contains(items[0].Content, runtime.GOOS) {
		t.Errorf("expected content to contain OS '%s'", runtime.GOOS)
	}
	if !strings.Contains(items[0].Content, runtime.GOARCH) {
		t.Errorf("expected content to contain architecture '%s'", runtime.GOARCH)
	}
}

func TestProvider_GetContextItems_Cwd(t *testing.T) {
	p := NewProvider()
	items, err := p.GetContextItems(context.Background(), "cwd", ports.ContextExtras{WorkDir: "/test/path"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Content != "/test/path" {
		t.Errorf("expected '/test/path', got '%s'", items[0].Content)
	}
}

func TestProvider_GetContextItems_All(t *testing.T) {
	p := NewProvider()
	items, err := p.GetContextItems(context.Background(), "all", ports.ContextExtras{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return multiple items
	if len(items) < 2 {
		t.Errorf("expected at least 2 items for 'all', got %d", len(items))
	}
}

func TestProvider_GetContextItems_InvalidQuery(t *testing.T) {
	p := NewProvider()
	_, err := p.GetContextItems(context.Background(), "invalid", ports.ContextExtras{})
	if err == nil {
		t.Error("expected error for invalid query")
	}
}
