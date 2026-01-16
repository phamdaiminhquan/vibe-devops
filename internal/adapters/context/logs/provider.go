package logs

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

// Provider provides log file analysis as context
type Provider struct {
	baseDir string
}

// NewProvider creates a new logs context provider
func NewProvider(baseDir string) *Provider {
	return &Provider{baseDir: baseDir}
}

// Description returns the provider's metadata
func (p *Provider) Description() ports.ContextProviderDescription {
	return ports.ContextProviderDescription{
		Name:         "logs",
		DisplayTitle: "Log Analysis",
		Description:  "Read and analyze log files. Use @logs path/to/file.log or @logs path/to/file.log:100 (last N lines)",
		Type:         ports.ContextTypeLogs,
	}
}

// Common error patterns for highlighting
var errorPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(error|err|fatal|fail|failed|failure|exception|panic|critical)\b`),
	regexp.MustCompile(`(?i)\b(warning|warn)\b`),
	regexp.MustCompile(`(?i)\b(timeout|timed out|connection refused|connection reset)\b`),
	regexp.MustCompile(`(?i)(exit\s+code|exitcode|status)\s*[=:]?\s*[1-9]\d*`),
	regexp.MustCompile(`(?i)(oom|out of memory|memory exhausted)`),
	regexp.MustCompile(`(?i)(permission denied|access denied|forbidden)`),
	regexp.MustCompile(`(?i)(not found|404|no such file)`),
}

// GetContextItems retrieves log content with analysis
func (p *Provider) GetContextItems(ctx context.Context, query string, extras ports.ContextExtras) ([]ports.ContextItem, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("log file path is required. Usage: @logs path/to/file.log or @logs path/to/file.log:100")
	}

	// Parse query: support format "path/to/file.log:100" for last N lines
	filePath := query
	lastN := 100 // Default: last 100 lines

	if idx := strings.LastIndex(query, ":"); idx > 0 {
		// Check if after : is a number
		suffix := query[idx+1:]
		if n := parseLineLimit(suffix); n > 0 {
			filePath = query[:idx]
			lastN = n
		}
	}

	// Resolve path
	absPath := filePath
	if !filepath.IsAbs(filePath) {
		baseDir := p.baseDir
		if extras.WorkDir != "" {
			baseDir = extras.WorkDir
		}
		absPath = filepath.Join(baseDir, filePath)
	}

	// Read log file
	lines, err := readLastNLines(absPath, lastN)
	if err != nil {
		return nil, err
	}

	// Analyze for errors
	var issues []string
	for i, line := range lines {
		for _, pattern := range errorPatterns {
			if pattern.MatchString(line) {
				issues = append(issues, fmt.Sprintf("Line %d: %s", i+1, strings.TrimSpace(line)))
				break
			}
		}
	}

	// Build content
	var content strings.Builder
	content.WriteString(fmt.Sprintf("=== Log File: %s (last %d lines) ===\n\n", filePath, len(lines)))

	// Show detected issues first
	if len(issues) > 0 {
		content.WriteString("⚠️ DETECTED ISSUES:\n")
		for _, issue := range issues {
			if len(issue) > 200 {
				issue = issue[:200] + "..."
			}
			content.WriteString("  • " + issue + "\n")
		}
		content.WriteString("\n")
	} else {
		content.WriteString("✅ No obvious errors detected in log snippet.\n\n")
	}

	content.WriteString("=== Log Content ===\n")
	for i, line := range lines {
		content.WriteString(fmt.Sprintf("%4d: %s\n", i+1, line))
	}

	return []ports.ContextItem{
		{
			Name:        filepath.Base(filePath),
			Description: fmt.Sprintf("Log analysis: %d lines, %d issues detected", len(lines), len(issues)),
			Content:     content.String(),
			URI:         "file://" + absPath,
		},
	}, nil
}

func parseLineLimit(s string) int {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil || n <= 0 {
		return 0
	}
	if n > 1000 {
		return 1000 // Cap at 1000 lines
	}
	return n
}

func readLastNLines(filePath string, n int) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open log file: %w", err)
	}
	defer f.Close()

	// Read all lines (for simplicity, could optimize for large files)
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Return last N lines
	if len(lines) <= n {
		return lines, nil
	}
	return lines[len(lines)-n:], nil
}

var _ ports.ContextProvider = (*Provider)(nil)
