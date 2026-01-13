package fs

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

type readFileInput struct {
	Path      string `json:"path"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	MaxBytes  int    `json:"maxBytes"`
}

type ReadFileTool struct {
	baseDir string
}

func NewReadFileTool(baseDir string) *ReadFileTool {
	return &ReadFileTool{baseDir: baseDir}
}

func (t *ReadFileTool) Name() string { return "read_file" }

func (t *ReadFileTool) Description() string {
	return "Read a text file (read-only), optionally by line range."
}

func (t *ReadFileTool) InputSchema() string {
	return `{"path":"string","startLine":"int (optional, 1-based)","endLine":"int (optional, inclusive)","maxBytes":"int (optional, default 65536)"}`
}

func (t *ReadFileTool) Run(ctx context.Context, input json.RawMessage) (string, error) {
	_ = ctx

	var in readFileInput
	if len(input) > 0 {
		_ = json.Unmarshal(input, &in)
	}
	if in.MaxBytes <= 0 {
		in.MaxBytes = 64 * 1024
	}
	if in.MaxBytes > 256*1024 {
		in.MaxBytes = 256 * 1024
	}
	if in.StartLine < 0 {
		in.StartLine = 0
	}
	if in.EndLine < 0 {
		in.EndLine = 0
	}

	abs, err := resolvePath(t.baseDir, in.Path)
	if err != nil {
		return "", err
	}

	f, err := os.Open(abs)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	start := in.StartLine
	end := in.EndLine
	if start == 0 {
		start = 1
	}
	if end == 0 {
		end = start + 200 - 1
	}
	if end < start {
		end = start
	}
	if (end - start) > 400 {
		end = start + 400
	}

	scanner := bufio.NewScanner(f)
	// Allow moderately long lines.
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 512*1024)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s (lines %d-%d)\n", abs, start, end))

	lineNo := 0
	bytesOut := 0
	for scanner.Scan() {
		lineNo++
		if lineNo < start {
			continue
		}
		if lineNo > end {
			break
		}
		line := scanner.Text()
		chunk := fmt.Sprintf("%6d: %s\n", lineNo, line)
		bytesOut += len(chunk)
		if bytesOut > in.MaxBytes {
			b.WriteString("... (truncated)\n")
			break
		}
		b.WriteString(chunk)
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.TrimSpace(b.String()), nil
}

var _ ports.Tool = (*ReadFileTool)(nil)
