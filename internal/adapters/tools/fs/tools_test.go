package fs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

func TestReadFileTool_Definition(t *testing.T) {
	tool := NewReadFileTool(".")
	def := tool.Definition()

	if def.Name != "read_file" {
		t.Errorf("expected Name 'read_file', got '%s'", def.Name)
	}
	if !def.ReadOnly {
		t.Error("expected ReadOnly to be true")
	}
	if def.DefaultPolicy != ports.PolicyAllowed {
		t.Errorf("expected PolicyAllowed, got '%s'", def.DefaultPolicy)
	}
}

func TestReadFileTool_Run(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\nline4\nline5"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewReadFileTool(tmpDir)

	input, _ := json.Marshal(map[string]any{
		"path":      "test.txt",
		"startLine": 2,
		"endLine":   4,
	})

	result, err := tool.Run(context.Background(), input, ports.ToolExtras{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected no error, got: %s", result.Content)
	}

	// Should contain lines 2-4
	if !strings.Contains(result.Content, "line2") {
		t.Error("expected content to contain 'line2'")
	}
	if !strings.Contains(result.Content, "line4") {
		t.Error("expected content to contain 'line4'")
	}
}

func TestListDirTool_Definition(t *testing.T) {
	tool := NewListDirTool(".")
	def := tool.Definition()

	if def.Name != "list_dir" {
		t.Errorf("expected Name 'list_dir', got '%s'", def.Name)
	}
	if !def.ReadOnly {
		t.Error("expected ReadOnly to be true")
	}
}

func TestListDirTool_Run(t *testing.T) {
	// Create temp dir with files
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte(""), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	tool := NewListDirTool(tmpDir)

	input, _ := json.Marshal(map[string]any{
		"path": ".",
	})

	result, err := tool.Run(context.Background(), input, ports.ToolExtras{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected no error, got: %s", result.Content)
	}

	if !strings.Contains(result.Content, "file1.txt") {
		t.Error("expected content to contain 'file1.txt'")
	}
	if !strings.Contains(result.Content, "subdir") {
		t.Error("expected content to contain 'subdir'")
	}
}

func TestGrepTool_Definition(t *testing.T) {
	tool := NewGrepTool(".")
	def := tool.Definition()

	if def.Name != "grep" {
		t.Errorf("expected Name 'grep', got '%s'", def.Name)
	}
	if !def.ReadOnly {
		t.Error("expected ReadOnly to be true")
	}
}

func TestGrepTool_Run(t *testing.T) {
	// Create temp file with searchable content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "search.txt")
	content := "hello world\nfoo bar\nhello again\ngoodbye"
	os.WriteFile(testFile, []byte(content), 0644)

	tool := NewGrepTool(tmpDir)

	input, _ := json.Marshal(map[string]any{
		"pattern": "hello",
		"path":    ".",
	})

	result, err := tool.Run(context.Background(), input, ports.ToolExtras{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected no error, got: %s", result.Content)
	}

	// Should find 2 matches
	if !strings.Contains(result.Content, "hello world") {
		t.Error("expected to find 'hello world'")
	}
	if !strings.Contains(result.Content, "hello again") {
		t.Error("expected to find 'hello again'")
	}
}

func TestToolPolicy_ReadOnly(t *testing.T) {
	readFileTool := NewReadFileTool(".")
	listDirTool := NewListDirTool(".")
	grepTool := NewGrepTool(".")

	// All read-only tools should return PolicyAllowed
	if readFileTool.EvaluatePolicy(nil) != ports.PolicyAllowed {
		t.Error("ReadFileTool should return PolicyAllowed")
	}
	if listDirTool.EvaluatePolicy(nil) != ports.PolicyAllowed {
		t.Error("ListDirTool should return PolicyAllowed")
	}
	if grepTool.EvaluatePolicy(nil) != ports.PolicyAllowed {
		t.Error("GrepTool should return PolicyAllowed")
	}
}
