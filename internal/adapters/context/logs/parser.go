package logs

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// LogFormat represents the detected log format
type LogFormat int

const (
	FormatPlain LogFormat = iota
	FormatJSON
	FormatLogfmt
)

// ParsedLogEntry represents a parsed structured log entry
type ParsedLogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Fields    map[string]interface{}
	Raw       string
}

// logfmt pattern: key=value or key="value with spaces"
var logfmtPattern = regexp.MustCompile(`(\w+)=(?:"([^"]+)"|(\S+))`)

// Common timestamp patterns
var timestampPatterns = []string{
	"2006-01-02T15:04:05.000Z",
	"2006-01-02T15:04:05Z",
	"2006-01-02 15:04:05.000",
	"2006-01-02 15:04:05",
	"2006/01/02 15:04:05",
	"Jan 02 15:04:05",
	"02/Jan/2006:15:04:05",
}

// DetectFormat determines if a line is JSON, logfmt, or plain text
func DetectFormat(line string) LogFormat {
	line = strings.TrimSpace(line)

	// Check for JSON
	if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(line), &js); err == nil {
			return FormatJSON
		}
	}

	// Check for logfmt (at least 2 key=value pairs)
	matches := logfmtPattern.FindAllStringSubmatch(line, -1)
	if len(matches) >= 2 {
		return FormatLogfmt
	}

	return FormatPlain
}

// ParseJSONLine parses a JSON log line
func ParseJSONLine(line string) (*ParsedLogEntry, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return nil, err
	}

	entry := &ParsedLogEntry{
		Fields: data,
		Raw:    line,
	}

	// Extract common fields
	// Level
	for _, key := range []string{"level", "lvl", "severity", "loglevel"} {
		if val, ok := data[key]; ok {
			entry.Level = levelFromString(fmt.Sprintf("%v", val))
			delete(data, key)
			break
		}
	}

	// Message
	for _, key := range []string{"msg", "message", "text", "log"} {
		if val, ok := data[key]; ok {
			entry.Message = fmt.Sprintf("%v", val)
			delete(data, key)
			break
		}
	}

	// Timestamp
	for _, key := range []string{"time", "timestamp", "ts", "@timestamp", "datetime"} {
		if val, ok := data[key]; ok {
			entry.Timestamp = parseTimestamp(fmt.Sprintf("%v", val))
			delete(data, key)
			break
		}
	}

	return entry, nil
}

// ParseLogfmtLine parses a logfmt line (key=value pairs)
func ParseLogfmtLine(line string) *ParsedLogEntry {
	entry := &ParsedLogEntry{
		Fields: make(map[string]interface{}),
		Raw:    line,
	}

	matches := logfmtPattern.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		key := match[1]
		value := match[2]
		if value == "" {
			value = match[3]
		}

		// Check for common keys
		switch strings.ToLower(key) {
		case "level", "lvl", "severity":
			entry.Level = levelFromString(value)
		case "msg", "message":
			entry.Message = value
		case "time", "timestamp", "ts":
			entry.Timestamp = parseTimestamp(value)
		default:
			entry.Fields[key] = value
		}
	}

	return entry
}

// FormatParsedEntry formats a parsed entry for display
func FormatParsedEntry(entry *ParsedLogEntry, useColor bool) string {
	var sb strings.Builder

	// Timestamp
	if !entry.Timestamp.IsZero() {
		sb.WriteString(entry.Timestamp.Format("15:04:05"))
		sb.WriteString(" ")
	}

	// Level with color
	levelStr := LevelString(entry.Level)
	if useColor {
		levelStr = colorizeLevel(levelStr, entry.Level)
	}
	sb.WriteString(fmt.Sprintf("[%-5s] ", levelStr))

	// Message
	if entry.Message != "" {
		sb.WriteString(entry.Message)
	}

	// Extra fields
	if len(entry.Fields) > 0 {
		sb.WriteString(" ")
		keys := make([]string, 0, len(entry.Fields))
		for k := range entry.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i, k := range keys {
			if i > 0 {
				sb.WriteString(" ")
			}
			if useColor {
				sb.WriteString(colorGray)
			}
			sb.WriteString(fmt.Sprintf("%s=%v", k, entry.Fields[k]))
			if useColor {
				sb.WriteString(colorReset)
			}
		}
	}

	return sb.String()
}

// PrettyPrintJSON formats JSON log with indentation and color
func PrettyPrintJSON(line string, useColor bool) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return line // Return as-is if not valid JSON
	}

	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return line
	}

	result := string(pretty)

	// Add color for keys if enabled
	if useColor {
		// Color JSON keys
		result = regexp.MustCompile(`"(\w+)":`).ReplaceAllString(result, colorCyan+`"$1":`+colorReset)
	}

	return result
}

func levelFromString(s string) LogLevel {
	s = strings.ToLower(s)
	switch s {
	case "fatal", "panic", "critical", "emerg":
		return LevelFatal
	case "error", "err":
		return LevelError
	case "warn", "warning":
		return LevelWarn
	case "info", "notice":
		return LevelInfo
	case "debug", "trace", "verbose":
		return LevelDebug
	default:
		return LevelUnknown
	}
}

func colorizeLevel(levelStr string, level LogLevel) string {
	switch level {
	case LevelFatal:
		return colorBold + colorRed + levelStr + colorReset
	case LevelError:
		return colorRed + levelStr + colorReset
	case LevelWarn:
		return colorYellow + levelStr + colorReset
	case LevelInfo:
		return colorCyan + levelStr + colorReset
	case LevelDebug:
		return colorGray + levelStr + colorReset
	default:
		return levelStr
	}
}

func parseTimestamp(s string) time.Time {
	for _, layout := range timestampPatterns {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}

	// Try parsing as Unix timestamp
	var ts int64
	if _, err := fmt.Sscanf(s, "%d", &ts); err == nil {
		if ts > 1e12 { // Milliseconds
			return time.Unix(ts/1000, (ts%1000)*1e6)
		}
		return time.Unix(ts, 0)
	}

	return time.Time{}
}
