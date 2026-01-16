package definitions

import "github.com/phamdaiminhquan/vibe-devops/internal/ports"

// SafeShell defines the safe_shell tool metadata
var SafeShell = ports.ToolDefinition{
	Name:         "safe_shell",
	DisplayTitle: "Run Shell Command",
	Description:  "Execute a shell command. Whitelisted commands run automatically, others require confirmation.",
	WouldLikeTo:  "run the following command",
	IsCurrently:  "executing command",
	HasAlready:   "executed the command",
	ReadOnly:     false, // Shell commands can have side effects
	InputSchema: `{
		"type": "object",
		"properties": {
			"command": {
				"type": "string",
				"description": "The shell command to execute"
			}
		},
		"required": ["command"]
	}`,
	DefaultPolicy: ports.PolicyWithPermission, // Requires user confirmation by default
	Group:         "system",
}

// SystemInfo defines the system_info tool metadata (future tool)
var SystemInfo = ports.ToolDefinition{
	Name:         "system_info",
	DisplayTitle: "System Information",
	Description:  "Get system information including OS, CPU, memory, and disk usage.",
	WouldLikeTo:  "get system information",
	IsCurrently:  "gathering system info",
	HasAlready:   "gathered system information",
	ReadOnly:     true,
	InputSchema: `{
		"type": "object",
		"properties": {
			"subsystem": {
				"type": "string",
				"enum": ["all", "cpu", "memory", "disk", "network"],
				"description": "Which subsystem to query (default: all)"
			}
		}
	}`,
	DefaultPolicy: ports.PolicyAllowed,
	Group:         "system",
}
