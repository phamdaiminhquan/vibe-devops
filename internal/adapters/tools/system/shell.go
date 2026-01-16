package system

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/tools/definitions"
	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

type safeShellInput struct {
	Command string `json:"command"`
}

// SafeShellTool implements the safe_shell tool
type SafeShellTool struct {
	allowedCmds []string
}

// NewSafeShellTool creates a new SafeShellTool
func NewSafeShellTool() *SafeShellTool {
	return &SafeShellTool{
		allowedCmds: []string{
			"ps", "netstat", "ss", "curl", "df", "free", "uptime", "id", "whoami", "date",
			"tasklist", "Get-Process", "Get-Service",
		},
	}
}

// Definition returns the tool metadata
func (t *SafeShellTool) Definition() ports.ToolDefinition {
	return definitions.SafeShell
}

// EvaluatePolicy checks if command is whitelisted
func (t *SafeShellTool) EvaluatePolicy(input json.RawMessage) ports.ToolPolicy {
	var in safeShellInput
	if err := json.Unmarshal(input, &in); err != nil {
		return ports.PolicyWithPermission
	}

	cmdStr := strings.TrimSpace(in.Command)
	for _, allowedCmd := range t.allowedCmds {
		if strings.HasPrefix(strings.ToLower(cmdStr), strings.ToLower(allowedCmd)+" ") ||
			strings.EqualFold(cmdStr, allowedCmd) {
			return ports.PolicyAllowed
		}
	}
	return ports.PolicyWithPermission
}

// Run executes the safe_shell tool
func (t *SafeShellTool) Run(ctx context.Context, input json.RawMessage, extras ports.ToolExtras) (ports.ToolResult, error) {
	var in safeShellInput
	if err := json.Unmarshal(input, &in); err != nil {
		return ports.ToolResult{IsError: true, Content: fmt.Sprintf("invalid input: %v", err)}, err
	}

	cmdStr := strings.TrimSpace(in.Command)
	if cmdStr == "" {
		return ports.ToolResult{IsError: true, Content: "command is required"}, fmt.Errorf("command is required")
	}

	// Check policy
	policy := t.EvaluatePolicy(input)
	if policy == ports.PolicyWithPermission {
		// Ask user for confirmation
		if extras.OnConfirm != nil && !extras.OnConfirm(fmt.Sprintf("Execute command: %s ?", cmdStr)) {
			return ports.ToolResult{
				Content: fmt.Sprintf("Command '%s' was rejected by user", cmdStr),
				Status:  "rejected",
				IsError: true,
			}, fmt.Errorf("command rejected by user")
		}
	}

	// Stream partial output if callback provided
	if extras.OnPartialOutput != nil {
		extras.OnPartialOutput(ports.PartialOutput{
			Content: fmt.Sprintf("Executing: %s", cmdStr),
			Status:  "executing",
		})
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
		return ports.ToolResult{
			Content: fmt.Sprintf("Exit Code: %v\nOutput:\n%s", err, output),
			Status:  "failed",
			IsError: true,
		}, nil // Return as result for Agent to analyze
	}

	// Truncate if too long
	if len(output) > 4000 {
		output = output[:4000] + "\n...(truncated)"
	}

	return ports.ToolResult{
		Content: output,
		Status:  "completed",
	}, nil
}

var _ ports.Tool = (*SafeShellTool)(nil)
