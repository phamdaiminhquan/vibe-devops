package system

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

func TestSafeShellTool_Definition(t *testing.T) {
	tool := NewSafeShellTool()
	def := tool.Definition()

	if def.Name != "safe_shell" {
		t.Errorf("expected Name 'safe_shell', got '%s'", def.Name)
	}
	if def.ReadOnly {
		t.Error("expected ReadOnly to be false for shell tool")
	}
	if def.DefaultPolicy != ports.PolicyWithPermission {
		t.Errorf("expected PolicyWithPermission, got '%s'", def.DefaultPolicy)
	}
}

func TestSafeShellTool_EvaluatePolicy_Whitelisted(t *testing.T) {
	tool := NewSafeShellTool()

	// Test whitelisted commands
	whitelisted := []string{"ps", "whoami", "date", "uptime"}
	for _, cmd := range whitelisted {
		input, _ := json.Marshal(map[string]string{"command": cmd})
		policy := tool.EvaluatePolicy(input)
		if policy != ports.PolicyAllowed {
			t.Errorf("expected '%s' to be PolicyAllowed, got '%s'", cmd, policy)
		}
	}
}

func TestSafeShellTool_EvaluatePolicy_NotWhitelisted(t *testing.T) {
	tool := NewSafeShellTool()

	// Test non-whitelisted commands
	notWhitelisted := []string{"rm -rf /", "sudo reboot", "shutdown"}
	for _, cmd := range notWhitelisted {
		input, _ := json.Marshal(map[string]string{"command": cmd})
		policy := tool.EvaluatePolicy(input)
		if policy != ports.PolicyWithPermission {
			t.Errorf("expected '%s' to be PolicyWithPermission, got '%s'", cmd, policy)
		}
	}
}

func TestSafeShellTool_Run_WhitelistedCommand(t *testing.T) {
	tool := NewSafeShellTool()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "whoami"
	} else {
		cmd = "whoami"
	}

	input, _ := json.Marshal(map[string]string{"command": cmd})

	result, err := tool.Run(context.Background(), input, ports.ToolExtras{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// whoami should return current user
	if result.IsError {
		t.Errorf("expected no error, got: %s", result.Content)
	}
	if strings.TrimSpace(result.Content) == "" {
		t.Error("expected non-empty output from whoami")
	}
}

func TestSafeShellTool_Run_EmptyCommand(t *testing.T) {
	tool := NewSafeShellTool()

	input, _ := json.Marshal(map[string]string{"command": ""})

	_, err := tool.Run(context.Background(), input, ports.ToolExtras{})
	if err == nil {
		t.Error("expected error for empty command")
	}
}
