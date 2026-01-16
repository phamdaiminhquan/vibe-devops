package tools

import (
	"sync"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

// Registry is the default implementation of ToolRegistry
type Registry struct {
	mu    sync.RWMutex
	tools map[string]ports.Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]ports.Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool ports.Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Definition().Name] = tool
	return nil
}

// Get returns a tool by name
func (r *Registry) Get(name string) (ports.Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tools
func (r *Registry) List() []ports.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ports.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}

// ListByGroup returns tools in a specific group
func (r *Registry) ListByGroup(group string) []ports.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []ports.Tool
	for _, t := range r.tools {
		if t.Definition().Group == group {
			result = append(result, t)
		}
	}
	return result
}

// Definitions returns all tool definitions (for AI prompt building)
func (r *Registry) Definitions() []ports.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ports.ToolDefinition, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t.Definition())
	}
	return result
}

var _ ports.ToolRegistry = (*Registry)(nil)
