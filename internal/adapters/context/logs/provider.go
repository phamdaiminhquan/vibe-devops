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

	// Detect log format from first non-empty line
	logFormat := FormatPlain
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			logFormat = DetectFormat(line)
			break
		}
	}

	// Analyze for issues using new highlighter
	issues := AnalyzeLines(lines)
	levelCounts := CountByLevel(lines)

	// Build content
	var content strings.Builder
	content.WriteString(fmt.Sprintf("=== Log File: %s (last %d lines) ===\n\n", filePath, len(lines)))

	// Show summary
	content.WriteString("Summary:\n")
	for _, level := range []LogLevel{LevelFatal, LevelError, LevelWarn, LevelInfo} {
		if count := levelCounts[level]; count > 0 {
			content.WriteString(fmt.Sprintf("  • %s: %d\n", LevelString(level), count))
		}
	}
	content.WriteString("\n")

	// Show detected issues
	if len(issues) > 0 {
		content.WriteString("⚠️ DETECTED ISSUES:\n")

		// Group by category
		issuesByCategory := make(map[string][]DetectedIssue)
		for _, issue := range issues {
			issuesByCategory[issue.Category] = append(issuesByCategory[issue.Category], issue)
		}

		for category, catIssues := range issuesByCategory {
			content.WriteString(fmt.Sprintf("\n  [%s] (%d)\n", category, len(catIssues)))
			for _, issue := range catIssues {
				if len(catIssues) <= 5 {
					content.WriteString(fmt.Sprintf("    Line %d: %s\n", issue.Line, issue.Content))
				}
			}
			if len(catIssues) > 5 {
				content.WriteString(fmt.Sprintf("    ... and %d more\n", len(catIssues)-5))
			}
		}
		content.WriteString("\n")
	} else {
		content.WriteString("✅ No obvious errors detected in log snippet.\n\n")
	}

	// Log content with highlighting
	content.WriteString("=== Log Content ===\n")

	for i, line := range lines {
		lineNum := fmt.Sprintf("%4d: ", i+1)

		// Format based on detected format
		switch logFormat {
		case FormatJSON:
			if entry, err := ParseJSONLine(line); err == nil {
				content.WriteString(lineNum + FormatParsedEntry(entry, false) + "\n")
			} else {
				content.WriteString(lineNum + line + "\n")
			}
		case FormatLogfmt:
			entry := ParseLogfmtLine(line)
			content.WriteString(lineNum + FormatParsedEntry(entry, false) + "\n")
		default:
			// Plain text - just show with level indicator
			level := DetectLevel(line)
			if level == LevelError || level == LevelFatal {
				content.WriteString(lineNum + "❌ " + line + "\n")
			} else if level == LevelWarn {
				content.WriteString(lineNum + "⚠️ " + line + "\n")
			} else {
				content.WriteString(lineNum + line + "\n")
			}
		}
	}

	formatName := "plain"
	if logFormat == FormatJSON {
		formatName = "JSON"
	} else if logFormat == FormatLogfmt {
		formatName = "logfmt"
	}

	return []ports.ContextItem{
		{
			Name:        filepath.Base(filePath),
			Description: fmt.Sprintf("Log analysis: %d lines, %d issues, format: %s", len(lines), len(issues), formatName),
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
