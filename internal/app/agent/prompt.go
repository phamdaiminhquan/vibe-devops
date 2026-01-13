package agent

import (
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

func buildAgentPrompt(goos, userRequest string, transcript []string, tools []ports.Tool) string {
	var b strings.Builder
	b.WriteString("You are Vibe, a CLI assistant that proposes ONE shell command for the user to run.\n")
	b.WriteString("You MAY request safe read-only tools to inspect the workspace before proposing a command.\n")
	b.WriteString("\n")
	b.WriteString("CRITICAL OUTPUT RULES:\n")
	b.WriteString("- Output EXACTLY ONE JSON object. No markdown, no code fences, no extra text.\n")
	b.WriteString("- Use {\"type\":\"tool\",\"tool\":...,\"input\":{...}} to call a tool.\n")
	b.WriteString("- Use {\"type\":\"done\",\"command\":...,\"explanation\":...} when ready.\n")
	b.WriteString("- command MUST be a single-line command string (no surrounding backticks).\n")
	b.WriteString("\n")
	b.WriteString("Environment:\n")
	b.WriteString("- GOOS: ")
	b.WriteString(strings.TrimSpace(goos))
	b.WriteString("\n\n")

	b.WriteString("Available tools:\n")
	for _, t := range tools {
		if t == nil {
			continue
		}
		b.WriteString("- ")
		b.WriteString(t.Name())
		b.WriteString(": ")
		b.WriteString(t.Description())
		b.WriteString(" Input schema: ")
		b.WriteString(t.InputSchema())
		b.WriteString("\n")
	}
	if len(tools) == 0 {
		b.WriteString("(none)\n")
	}
	b.WriteString("\n")

	b.WriteString("Task:\n")
	b.WriteString(userRequest)
	b.WriteString("\n\n")

	b.WriteString("Transcript (most recent last):\n")
	for _, line := range transcript {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	b.WriteString("Reminder: If the task can be solved without tools, return type=done immediately.\n")
	b.WriteString("If you use tools, keep tool calls minimal and stop once you have enough info.\n")

	return b.String()
}
