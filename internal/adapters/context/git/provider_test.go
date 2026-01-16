package git

import (
	"context"
	"testing"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

func TestProvider_Description(t *testing.T) {
	p := NewProvider(".")
	desc := p.Description()

	if desc.Name != "git" {
		t.Errorf("expected Name 'git', got '%s'", desc.Name)
	}
	if desc.Type != ports.ContextTypeGit {
		t.Errorf("expected Type ContextTypeGit, got '%s'", desc.Type)
	}
}

func TestProvider_GetContextItems_Status(t *testing.T) {
	// This test assumes we're running in a git repo
	p := NewProvider(".")
	items, err := p.GetContextItems(context.Background(), "status", ports.ContextExtras{})
	if err != nil {
		t.Skipf("skipping test - not in a git repo: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Name != "git-status" {
		t.Errorf("expected Name 'git-status', got '%s'", items[0].Name)
	}
}

func TestProvider_GetContextItems_Branch(t *testing.T) {
	p := NewProvider(".")
	items, err := p.GetContextItems(context.Background(), "branch", ports.ContextExtras{})
	if err != nil {
		t.Skipf("skipping test - not in a git repo: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Name != "git-branch" {
		t.Errorf("expected Name 'git-branch', got '%s'", items[0].Name)
	}
}

func TestProvider_GetContextItems_InvalidQuery(t *testing.T) {
	p := NewProvider(".")
	_, err := p.GetContextItems(context.Background(), "invalid-query", ports.ContextExtras{})
	if err == nil {
		t.Error("expected error for invalid query")
	}
}
