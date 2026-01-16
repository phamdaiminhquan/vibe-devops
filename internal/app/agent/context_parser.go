package agent

import (
	"regexp"
	"strings"
)

// ContextMention represents a parsed @mention in user input
type ContextMention struct {
	// Provider is the context provider name (e.g., "file", "git", "system")
	Provider string
	// Query is the argument after the provider (e.g., "path/to/file" for @file)
	Query string
	// Raw is the original matched string
	Raw string
}

// mentionPattern matches @provider query patterns
// Examples: @file path/to/file.txt, @git status, @system os
// Only captures the provider name and the first argument (path or command)
var mentionPattern = regexp.MustCompile(`@(\w+)\s+(\S+)`)

// ParseContextMentions extracts all @mentions from user input
func ParseContextMentions(input string) []ContextMention {
	matches := mentionPattern.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return nil
	}

	mentions := make([]ContextMention, 0, len(matches))
	for _, m := range matches {
		if len(m) >= 3 {
			mentions = append(mentions, ContextMention{
				Provider: strings.ToLower(m[1]),
				Query:    strings.TrimSpace(m[2]),
				Raw:      m[0],
			})
		}
	}
	return mentions
}

// StripContextMentions removes @mentions from user input
// Returns the cleaned input
func StripContextMentions(input string) string {
	return strings.TrimSpace(mentionPattern.ReplaceAllString(input, ""))
}
