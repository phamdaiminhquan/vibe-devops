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
	b.WriteString("- Use {\"type\":\"tool\",\"thought\":\"user-friendly status\",\"tool\":...,\"input\":{...}} to call a tool.\n")
	b.WriteString("- Use {\"type\":\"done\",\"command\":...,\"explanation\":...} when you want to propose a command to run.\n")
	b.WriteString("- Use {\"type\":\"answer\",\"explanation\":...} when you can answer WITHOUT a command OR need to clarify user intent (e.g. 'What is be?').\n")
	b.WriteString("- 'thought' is REQUIRED for tools. It must be a short, friendly status message for the user (e.g., 'Checking backend folder...').\n")
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
		def := t.Definition()
		b.WriteString("- ")
		b.WriteString(def.Name)
		b.WriteString(": ")
		b.WriteString(def.Description)
		b.WriteString(" Input schema: ")
		b.WriteString(def.InputSchema)
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
	b.WriteString("EFFICIENCY TIP: You can run complex shell commands! Instead of 3 separate calls (e.g. check dir, then ps, then netstat), use ONE safe_shell call with joined commands (e.g. 'ls -F && ps aux | grep app && netstat -tulpn'). Save your steps.\n")

	return b.String()
}
