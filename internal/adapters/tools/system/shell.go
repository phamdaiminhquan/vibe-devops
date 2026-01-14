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
	OnConfirm   func(string) bool
}

func NewSafeShellTool(onConfirm func(string) bool) *SafeShellTool {
	return &SafeShellTool{
		allowedCmds: []string{
			"ps", "netstat", "ss", "curl", "df", "free", "uptime", "id", "whoami", "date",
			"tasklist", "Get-Process", "Get-Service",
		},
		OnConfirm: onConfirm,
	}
}

func (t *SafeShellTool) Name() string { return "safe_shell" }

func (t *SafeShellTool) Description() string {
	return "Execute safe system commands. If a command is not whitelisted, the user will be asked for permission."
}

func (t *SafeShellTool) InputSchema() string {
	return `{"command":"string"}`
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
	allowed := false
	for _, allowedCmd := range t.allowedCmds {
		// Case-insensitive check for Windows friendliness
		if strings.HasPrefix(strings.ToLower(cmdStr), strings.ToLower(allowedCmd)+" ") || strings.EqualFold(cmdStr, allowedCmd) {
			allowed = true
			break
		}
	}

	// If not allowed by default, ask user
	if !allowed {
		if t.OnConfirm != nil && t.OnConfirm(cmdStr) {
			allowed = true
		}
	}

	if !allowed {
		return "", fmt.Errorf("command '%s' is not allowed and was rejected by user", cmdStr)
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
