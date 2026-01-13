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

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

type grepInput struct {
	Pattern    string `json:"pattern"`
	Path       string `json:"path"`
	MaxMatches int    `json:"maxMatches"`
}

type GrepTool struct {
	baseDir string
}

func NewGrepTool(baseDir string) *GrepTool {
	return &GrepTool{baseDir: baseDir}
}

func (t *GrepTool) Name() string { return "grep" }

func (t *GrepTool) Description() string {
	return "Search for a regex pattern in text files under a path (read-only)."
}

func (t *GrepTool) InputSchema() string {
	return `{"pattern":"string (regex)","path":"string (optional, default .)","maxMatches":"int (optional, default 50)"}`
}

func (t *GrepTool) Run(ctx context.Context, input json.RawMessage) (string, error) {
	_ = ctx

	var in grepInput
	if len(input) > 0 {
		_ = json.Unmarshal(input, &in)
	}
	if strings.TrimSpace(in.Pattern) == "" {
		return "", fmt.Errorf("pattern is required")
	}
	if in.MaxMatches <= 0 {
		in.MaxMatches = 50
	}
	if in.MaxMatches > 200 {
		in.MaxMatches = 200
	}

	re, err := regexp.Compile(in.Pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex: %w", err)
	}

	root, err := resolvePath(t.baseDir, in.Path)
	if err != nil {
		return "", err
	}

	baseAbs, err := filepath.Abs(t.baseDir)
	if err != nil {
		return "", err
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
		return "", err
	}

	if matches == 0 {
		b.WriteString("(no matches)\n")
	}
	return strings.TrimSpace(b.String()), nil
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
