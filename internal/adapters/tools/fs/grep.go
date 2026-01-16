package fs

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/tools/definitions"
	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

type grepInput struct {
	Pattern    string `json:"pattern"`
	Path       string `json:"path"`
	MaxMatches int    `json:"maxMatches"`
}

// GrepTool implements the grep tool
type GrepTool struct {
	baseDir string
}

// NewGrepTool creates a new GrepTool
func NewGrepTool(baseDir string) *GrepTool {
	return &GrepTool{baseDir: baseDir}
}

// Definition returns the tool metadata
func (t *GrepTool) Definition() ports.ToolDefinition {
	return definitions.Grep
}

// EvaluatePolicy always returns allowed for read-only operations
func (t *GrepTool) EvaluatePolicy(_ json.RawMessage) ports.ToolPolicy {
	return ports.PolicyAllowed
}

// Run executes the grep tool
func (t *GrepTool) Run(ctx context.Context, input json.RawMessage, extras ports.ToolExtras) (ports.ToolResult, error) {
	_ = ctx

	var in grepInput
	if len(input) > 0 {
		_ = json.Unmarshal(input, &in)
	}
	if strings.TrimSpace(in.Pattern) == "" {
		return ports.ToolResult{IsError: true, Content: "pattern is required"}, fmt.Errorf("pattern is required")
	}
	if in.MaxMatches <= 0 {
		in.MaxMatches = 50
	}
	if in.MaxMatches > 200 {
		in.MaxMatches = 200
	}

	re, err := regexp.Compile(in.Pattern)
	if err != nil {
		return ports.ToolResult{IsError: true, Content: fmt.Sprintf("invalid regex: %v", err)}, err
	}

	root, err := resolvePath(t.baseDir, in.Path)
	if err != nil {
		return ports.ToolResult{IsError: true, Content: err.Error()}, err
	}

	baseAbs, err := filepath.Abs(t.baseDir)
	if err != nil {
		return ports.ToolResult{IsError: true, Content: err.Error()}, err
	}

	// Stream partial output if callback provided
	if extras.OnPartialOutput != nil {
		extras.OnPartialOutput(ports.PartialOutput{
			Content: fmt.Sprintf("Searching for %q in %s...", in.Pattern, root),
			Status:  "searching",
		})
	}

	matches := 0
	var b strings.Builder
	b.WriteString(fmt.Sprintf("grep %q under %s\n", in.Pattern, root))

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}

		name := d.Name()
		if d.IsDir() {
			if name == ".git" || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() > 1_000_000 {
			return nil
		}

		rel, _ := filepath.Rel(baseAbs, path)
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer func() { _ = f.Close() }()

		// Quick binary detection: if first chunk contains NUL, skip.
		buf := make([]byte, 4096)
		n, _ := f.Read(buf)
		if bytesContainsNUL(buf[:n]) {
			return nil
		}
		_, _ = f.Seek(0, io.SeekStart)

		scanner := bufio.NewScanner(f)
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()
			if re.FindStringIndex(line) != nil {
				b.WriteString(fmt.Sprintf("%s:%d: %s\n", filepath.ToSlash(rel), lineNo, strings.TrimSpace(line)))
				matches++
				if matches >= in.MaxMatches {
					return io.EOF
				}
			}
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return ports.ToolResult{IsError: true, Content: err.Error()}, err
	}

	if matches == 0 {
		b.WriteString("(no matches)\n")
	}

	return ports.ToolResult{
		Content: strings.TrimSpace(b.String()),
		Status:  fmt.Sprintf("found %d matches", matches),
	}, nil
}

func bytesContainsNUL(b []byte) bool {
	for _, c := range b {
		if c == 0 {
			return true
		}
	}
	return false
}

var _ ports.Tool = (*GrepTool)(nil)
