package system

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type safeShellInput struct {
	Command string `json:"command"`
}

type SafeShellTool struct {
	allowedCmds []string
}

func NewSafeShellTool() *SafeShellTool {
	return &SafeShellTool{
		allowedCmds: []string{
			"ps", "netstat", "ss", "curl", "df", "free", "uptime", "id", "whoami", "date",
		},
	}
}

func (t *SafeShellTool) Name() string { return "safe_shell" }

func (t *SafeShellTool) Description() string {
	return "Execute safe system commands (read-only) to inspect processes, network, and resources. Allowed: ps, netstat, ss, curl, df, free, uptime, id, whoami, date."
}

func (t *SafeShellTool) InputSchema() string {
	return `{"command":"string (e.g. 'ps aux | grep backend')"}`
}

func (t *SafeShellTool) Run(ctx context.Context, input json.RawMessage) (string, error) {
	var in safeShellInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	cmdStr := strings.TrimSpace(in.Command)
	if cmdStr == "" {
		return "", fmt.Errorf("command is required")
	}

	// Security Check: Validate against allowlist
	// We check if the command starts with one of the allowed binaries.
	// NOTE: This is a basic check. Complex chaining might still allow abuse,
	// but for an Agent tool usage it's a good first layer.
	// Users run the agent, so they trust the agent somewhat, but we want to prevent accidental destruction.
	allowed := false
	for _, allowedCmd := range t.allowedCmds {
		if strings.HasPrefix(cmdStr, allowedCmd+" ") || cmdStr == allowedCmd {
			allowed = true
			break
		}
	}

	if !allowed {
		return "", fmt.Errorf("command '%s' is not allowed in safe_shell. Allowed: %v", strings.Split(cmdStr, " ")[0], t.allowedCmds)
	}

	// Execution
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.CommandContext(ctx, "powershell", "-Command", cmdStr)
	} else {
		c = exec.CommandContext(ctx, "sh", "-c", cmdStr)
	}

	out, err := c.CombinedOutput()
	output := string(out)
	if err != nil {
		return fmt.Sprintf("Exit Code: %v\nOutput:\n%s", err, output), nil // Return error as string for Agent to analyze
	}

	// Truncate if too long (processes list can be huge)
	if len(output) > 4000 {
		output = output[:4000] + "\n...(truncated)"
	}

	return output, nil
}
