package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

type listDirInput struct {
	Path       string `json:"path"`
	MaxEntries int    `json:"maxEntries"`
}

type ListDirTool struct {
	baseDir string
}

func NewListDirTool(baseDir string) *ListDirTool {
	return &ListDirTool{baseDir: baseDir}
}

func (t *ListDirTool) Name() string { return "list_dir" }

func (t *ListDirTool) Description() string {
	return "List entries in a directory (read-only)."
}

func (t *ListDirTool) InputSchema() string {
	return `{"path":"string","maxEntries":"int (optional, default 200)"}`
}

func (t *ListDirTool) Run(ctx context.Context, input json.RawMessage) (string, error) {
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
		return "", err
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return "", err
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

	return strings.TrimSpace(b.String()), nil
}

var _ ports.Tool = (*ListDirTool)(nil)
