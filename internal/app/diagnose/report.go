package diagnose

import (
	"fmt"
	"strings"
)

// ANSI colors
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// FormatReport formats the diagnosis result for terminal output
func FormatReport(result *DiagnoseResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n%s%s System Diagnostics Report%s\n", colorBold, colorCyan, colorReset))
	sb.WriteString("══════════════════════════════\n\n")

	// Errors first (most critical)
	if len(result.Errors) > 0 {
		sb.WriteString(fmt.Sprintf("%s%s🔴 CRITICAL ERRORS:%s\n", colorBold, colorRed, colorReset))
		for _, issue := range result.Errors {
			sb.WriteString(formatIssue(issue, colorRed))
		}
		sb.WriteString("\n")
	}

	// Warnings
	if len(result.Warnings) > 0 {
		sb.WriteString(fmt.Sprintf("%s%s⚠️  WARNINGS:%s\n", colorBold, colorYellow, colorReset))
		for _, issue := range result.Warnings {
			sb.WriteString(formatIssue(issue, colorYellow))
		}
		sb.WriteString("\n")
	}

	// OK checks
	if len(result.OK) > 0 {
		sb.WriteString(fmt.Sprintf("%s%s✅ OK:%s\n", colorBold, colorGreen, colorReset))
		for _, check := range result.OK {
			sb.WriteString(fmt.Sprintf("  • %s: %s\n", check.Description, check.Value))
		}
		sb.WriteString("\n")
	}

	// Suggested fixes
	fixes := collectFixes(result)
	if len(fixes) > 0 {
		sb.WriteString(fmt.Sprintf("%s💡 Suggestions:%s\n", colorBold, colorReset))
		for i, fix := range fixes {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, fix))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func formatIssue(issue Issue, color string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("  %s• %s%s", color, issue.Description, colorReset))
	if issue.Threshold != "" {
		sb.WriteString(fmt.Sprintf(" (threshold: %s)", issue.Threshold))
	}
	sb.WriteString("\n")
	if issue.FixCommand != "" {
		sb.WriteString(fmt.Sprintf("    → Fix: %s\n", issue.FixCommand))
	}
	return sb.String()
}

func collectFixes(result *DiagnoseResult) []string {
	var fixes []string
	seen := make(map[string]bool)

	for _, issue := range result.Errors {
		if issue.FixCommand != "" && !seen[issue.FixCommand] {
			fixes = append(fixes, issue.FixCommand)
			seen[issue.FixCommand] = true
		}
	}
	for _, issue := range result.Warnings {
		if issue.FixCommand != "" && !seen[issue.FixCommand] {
			fixes = append(fixes, issue.FixCommand)
			seen[issue.FixCommand] = true
		}
	}

	return fixes
}
