package logs

import (
	"regexp"
	"strings"
)

// ANSI color codes
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorYellow  = "\033[33m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"
	colorBold    = "\033[1m"
	colorMagenta = "\033[35m"
)

// LogLevel represents the severity of a log line
type LogLevel int

const (
	LevelUnknown LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// Level detection patterns
var levelPatterns = map[LogLevel]*regexp.Regexp{
	LevelFatal: regexp.MustCompile(`(?i)\b(fatal|panic|critical|emerg)\b`),
	LevelError: regexp.MustCompile(`(?i)\b(error|err|fail|failed|failure|exception)\b`),
	LevelWarn:  regexp.MustCompile(`(?i)\b(warn|warning)\b`),
	LevelInfo:  regexp.MustCompile(`(?i)\b(info|notice)\b`),
	LevelDebug: regexp.MustCompile(`(?i)\b(debug|trace|verbose)\b`),
}

// Issue patterns for detection
var issuePatterns = []struct {
	Pattern     *regexp.Regexp
	Description string
	Category    string
}{
	{regexp.MustCompile(`(?i)(timeout|timed out)`), "Timeout detected", "performance"},
	{regexp.MustCompile(`(?i)(connection refused|connection reset|ECONNREFUSED)`), "Connection issue", "network"},
	{regexp.MustCompile(`(?i)(oom|out of memory|memory exhausted|ENOMEM)`), "Memory issue", "resource"},
	{regexp.MustCompile(`(?i)(permission denied|access denied|forbidden|EACCES)`), "Permission issue", "security"},
	{regexp.MustCompile(`(?i)(not found|404|no such file|ENOENT)`), "Not found", "filesystem"},
	{regexp.MustCompile(`(?i)(disk full|no space left|ENOSPC)`), "Disk full", "resource"},
	{regexp.MustCompile(`(?i)(segmentation fault|segfault|SIGSEGV)`), "Crash detected", "crash"},
	{regexp.MustCompile(`(?i)(killed|OOMKilled|exit code [1-9])`), "Process killed", "crash"},
}

// DetectedIssue represents an issue found in logs
type DetectedIssue struct {
	Line        int
	Content     string
	Level       LogLevel
	Description string
	Category    string
}

// HighlightLine adds ANSI colors to a log line based on level
func HighlightLine(line string) string {
	level := DetectLevel(line)
	return colorizeByLevel(line, level)
}

// DetectLevel determines the log level of a line
func DetectLevel(line string) LogLevel {
	// Check in order of severity (highest first)
	for _, level := range []LogLevel{LevelFatal, LevelError, LevelWarn, LevelInfo, LevelDebug} {
		if pattern, ok := levelPatterns[level]; ok {
			if pattern.MatchString(line) {
				return level
			}
		}
	}
	return LevelUnknown
}

func colorizeByLevel(line string, level LogLevel) string {
	switch level {
	case LevelFatal:
		return colorBold + colorRed + line + colorReset
	case LevelError:
		return colorRed + line + colorReset
	case LevelWarn:
		return colorYellow + line + colorReset
	case LevelInfo:
		return colorCyan + line + colorReset
	case LevelDebug:
		return colorGray + line + colorReset
	default:
		return line
	}
}

// LevelString returns string representation of level
func LevelString(level LogLevel) string {
	switch level {
	case LevelFatal:
		return "FATAL"
	case LevelError:
		return "ERROR"
	case LevelWarn:
		return "WARN"
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	default:
		return "UNKNOWN"
	}
}

// AnalyzeLines scans lines for issues and returns detected problems
func AnalyzeLines(lines []string) []DetectedIssue {
	var issues []DetectedIssue

	for i, line := range lines {
		level := DetectLevel(line)
		
		// Check for specific issue patterns
		for _, p := range issuePatterns {
			if p.Pattern.MatchString(line) {
				issues = append(issues, DetectedIssue{
					Line:        i + 1,
					Content:     truncate(line, 150),
					Level:       level,
					Description: p.Description,
					Category:    p.Category,
				})
				break // Only report first matching pattern per line
			}
		}

		// Also capture error/fatal lines without specific patterns
		if level == LevelError || level == LevelFatal {
			found := false
			for _, issue := range issues {
				if issue.Line == i+1 {
					found = true
					break
				}
			}
			if !found {
				issues = append(issues, DetectedIssue{
					Line:        i + 1,
					Content:     truncate(line, 150),
					Level:       level,
					Description: LevelString(level) + " log entry",
					Category:    "error",
				})
			}
		}
	}

	return issues
}

// SummarizeIssues groups issues by category
func SummarizeIssues(issues []DetectedIssue) map[string]int {
	summary := make(map[string]int)
	for _, issue := range issues {
		summary[issue.Category]++
	}
	return summary
}

// CountByLevel counts issues by log level
func CountByLevel(lines []string) map[LogLevel]int {
	counts := make(map[LogLevel]int)
	for _, line := range lines {
		level := DetectLevel(line)
		counts[level]++
	}
	return counts
}

func truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
