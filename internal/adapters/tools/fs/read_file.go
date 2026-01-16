package fs

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/tools/definitions"
	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

type readFileInput struct {
	Path      string `json:"path"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	MaxBytes  int    `json:"maxBytes"`
}

// ReadFileTool implements the read_file tool
type ReadFileTool struct {
	baseDir string
}

// NewReadFileTool creates a new ReadFileTool
func NewReadFileTool(baseDir string) *ReadFileTool {
	return &ReadFileTool{baseDir: baseDir}
}

// Definition returns the tool metadata
func (t *ReadFileTool) Definition() ports.ToolDefinition {
	return definitions.ReadFile
}

// EvaluatePolicy always returns allowed for read-only operations
func (t *ReadFileTool) EvaluatePolicy(_ json.RawMessage) ports.ToolPolicy {
	return ports.PolicyAllowed
}

// Run executes the read_file tool
func (t *ReadFileTool) Run(ctx context.Context, input json.RawMessage, extras ports.ToolExtras) (ports.ToolResult, error) {
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
		return ports.ToolResult{IsError: true, Content: err.Error()}, err
	}

	f, err := os.Open(abs)
	if err != nil {
		return ports.ToolResult{IsError: true, Content: err.Error()}, err
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

	// Stream partial output if callback provided
	if extras.OnPartialOutput != nil {
		extras.OnPartialOutput(ports.PartialOutput{
			Content: fmt.Sprintf("Reading %s...", abs),
			Status:  "reading",
		})
	}

	scanner := bufio.NewScanner(f)
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
		return ports.ToolResult{IsError: true, Content: err.Error()}, err
	}

	return ports.ToolResult{
		Content: strings.TrimSpace(b.String()),
		Status:  "completed",
	}, nil
}

var _ ports.Tool = (*ReadFileTool)(nil)
