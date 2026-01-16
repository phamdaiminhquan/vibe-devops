package ports

import (
	"context"
	"encoding/json"
)

// ToolPolicy defines the permission level for tool execution
type ToolPolicy string

const (
	// PolicyAllowed - Tool can run without confirmation
	PolicyAllowed ToolPolicy = "allowed"
	// PolicyWithPermission - Tool requires user confirmation
	PolicyWithPermission ToolPolicy = "allowedWithPermission"
	// PolicyDenied - Tool is not allowed to run
	PolicyDenied ToolPolicy = "denied"
)

// ToolDefinition contains metadata about a tool
type ToolDefinition struct {
	// Name is the unique identifier for the tool (used in AI function calls)
	Name string `json:"name"`
	// DisplayTitle is human-readable name for UI
	DisplayTitle string `json:"displayTitle"`
	// Description explains what the tool does (for AI context)
	Description string `json:"description"`
	// WouldLikeTo describes intent: "would like to read file..."
	WouldLikeTo string `json:"wouldLikeTo,omitempty"`
	// IsCurrently describes action: "is currently reading..."
	IsCurrently string `json:"isCurrently,omitempty"`
	// HasAlready describes completion: "has already read..."
	HasAlready string `json:"hasAlready,omitempty"`
	// ReadOnly indicates the tool has no side effects
	ReadOnly bool `json:"readOnly"`
	// InputSchema is JSON Schema for input validation
	InputSchema string `json:"inputSchema"`
	// DefaultPolicy is the default permission level
	DefaultPolicy ToolPolicy `json:"defaultPolicy"`
	// Group categorizes the tool (e.g., "filesystem", "system", "network")
	Group string `json:"group,omitempty"`
}

// ToolExtras provides context and callbacks for tool execution
type ToolExtras struct {
	// WorkDir is the working directory for the tool
	WorkDir string
	// OnPartialOutput is called for streaming output (optional)
	OnPartialOutput func(output PartialOutput)
	// OnConfirm is called when tool needs user confirmation (optional)
	OnConfirm func(message string) bool
}

// PartialOutput represents a chunk of streaming output
type PartialOutput struct {
	// Content is the partial result content
	Content string
	// Status describes the current state (e.g., "Reading file...", "Completed")
	Status string
	// IsError indicates if this output represents an error
	IsError bool
}

// ToolResult contains the output of a tool execution
type ToolResult struct {
	// Content is the main result content
	Content string `json:"content"`
	// Status describes the final state
	Status string `json:"status,omitempty"`
	// IsError indicates if the execution failed
	IsError bool `json:"isError,omitempty"`
}

// Tool is a capability that an agent can invoke
// Tools can be read-only (safe) or write (require confirmation)
type Tool interface {
	// Definition returns the tool's metadata
	Definition() ToolDefinition

	// EvaluatePolicy determines the policy based on input
	// This allows dynamic policy based on args (e.g., rm -rf / should be denied)
	EvaluatePolicy(input json.RawMessage) ToolPolicy

	// Run executes the tool with the given input and extras
	// Returns a ToolResult with content and status
	Run(ctx context.Context, input json.RawMessage, extras ToolExtras) (ToolResult, error)
}

// ToolRegistry manages available tools
type ToolRegistry interface {
	// Register adds a tool to the registry
	Register(tool Tool) error
	// Get returns a tool by name
	Get(name string) (Tool, bool)
	// List returns all registered tools
	List() []Tool
	// ListByGroup returns tools in a specific group
	ListByGroup(group string) []Tool
}
