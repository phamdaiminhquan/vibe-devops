package context

import (
	"sync"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

// Registry is the default implementation of ContextProviderRegistry
type Registry struct {
	mu        sync.RWMutex
	providers map[string]ports.ContextProvider
}

// NewRegistry creates a new context provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]ports.ContextProvider),
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(provider ports.ContextProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[provider.Description().Name] = provider
	return nil
}

// Get returns a provider by name
func (r *Registry) Get(name string) (ports.ContextProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, ok := r.providers[name]
	return provider, ok
}

// List returns all registered providers
func (r *Registry) List() []ports.ContextProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ports.ContextProvider, 0, len(r.providers))
	for _, p := range r.providers {
		result = append(result, p)
	}
	return result
}

// ListByType returns providers of a specific type
func (r *Registry) ListByType(providerType ports.ContextProviderType) []ports.ContextProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []ports.ContextProvider
	for _, p := range r.providers {
		if p.Description().Type == providerType {
			result = append(result, p)
		}
	}
	return result
}

// Descriptions returns all provider descriptions (for prompt building)
func (r *Registry) Descriptions() []ports.ContextProviderDescription {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ports.ContextProviderDescription, 0, len(r.providers))
	for _, p := range r.providers {
		result = append(result, p.Description())
	}
	return result
}

var _ ports.ContextProviderRegistry = (*Registry)(nil)
