package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/tools/definitions"
	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

type listDirInput struct {
	Path       string `json:"path"`
	MaxEntries int    `json:"maxEntries"`
}

// ListDirTool implements the list_dir tool
type ListDirTool struct {
	baseDir string
}

// NewListDirTool creates a new ListDirTool
func NewListDirTool(baseDir string) *ListDirTool {
	return &ListDirTool{baseDir: baseDir}
}

// Definition returns the tool metadata
func (t *ListDirTool) Definition() ports.ToolDefinition {
	return definitions.ListDir
}

// EvaluatePolicy always returns allowed for read-only operations
func (t *ListDirTool) EvaluatePolicy(_ json.RawMessage) ports.ToolPolicy {
	return ports.PolicyAllowed
}

// Run executes the list_dir tool
func (t *ListDirTool) Run(ctx context.Context, input json.RawMessage, extras ports.ToolExtras) (ports.ToolResult, error) {
	_ = ctx

	var in listDirInput
	if len(input) > 0 {
		_ = json.Unmarshal(input, &in)
	}
	if in.MaxEntries <= 0 {
		in.MaxEntries = 200
	}
	if in.MaxEntries > 500 {
		in.MaxEntries = 500
	}

	abs, err := resolvePath(t.baseDir, in.Path)
	if err != nil {
		return ports.ToolResult{IsError: true, Content: err.Error()}, err
	}

	// Stream partial output if callback provided
	if extras.OnPartialOutput != nil {
		extras.OnPartialOutput(ports.PartialOutput{
			Content: fmt.Sprintf("Listing %s...", abs),
			Status:  "reading",
		})
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return ports.ToolResult{IsError: true, Content: err.Error()}, err
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	if len(entries) > in.MaxEntries {
		entries = entries[:in.MaxEntries]
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s\n", abs))
	for _, e := range entries {
		kind := "file"
		if e.IsDir() {
			kind = "dir"
		}
		b.WriteString(fmt.Sprintf("- [%s] %s\n", kind, e.Name()))
	}

	return ports.ToolResult{
		Content: strings.TrimSpace(b.String()),
		Status:  "completed",
	}, nil
}

var _ ports.Tool = (*ListDirTool)(nil)
