package ports

import (
	"context"
)

// ContextItem represents a piece of context that can be provided to the AI
type ContextItem struct {
	// Name is a short identifier for this context item
	Name string `json:"name"`
	// Description explains what this context contains
	Description string `json:"description,omitempty"`
	// Content is the actual context content
	Content string `json:"content"`
	// URI is an optional reference to the source
	URI string `json:"uri,omitempty"`
	// Hidden indicates if this should be sent to AI but not shown to user
	Hidden bool `json:"hidden,omitempty"`
	// Icon is an optional icon for UI display
	Icon string `json:"icon,omitempty"`
}

// ContextProviderDescription describes a context provider's capabilities
type ContextProviderDescription struct {
	// Name is the unique identifier for this provider (used in @mentions)
	Name string `json:"name"`
	// DisplayTitle is human-readable name for UI
	DisplayTitle string `json:"displayTitle"`
	// Description explains what context this provider offers
	Description string `json:"description"`
	// Type indicates the provider category
	Type ContextProviderType `json:"type"`
}

// ContextProviderType categorizes context providers
type ContextProviderType string

const (
	// ContextTypeFile provides file-based context
	ContextTypeFile ContextProviderType = "file"
	// ContextTypeGit provides git-related context
	ContextTypeGit ContextProviderType = "git"
	// ContextTypeSystem provides system information context
	ContextTypeSystem ContextProviderType = "system"
	// ContextTypeLogs provides log analysis context
	ContextTypeLogs ContextProviderType = "logs"
	// ContextTypeCustom for user-defined providers
	ContextTypeCustom ContextProviderType = "custom"
)

// ContextExtras provides additional context and dependencies for providers
type ContextExtras struct {
	// WorkDir is the current working directory
	WorkDir string
	// Provider allows context providers to use AI for smart context
	Provider Provider
	// SelectedText is any text the user has selected
	SelectedText string
	// FullInput is the user's full input/query
	FullInput string
}

// ContextProvider is the interface for providing context to the AI
type ContextProvider interface {
	// Description returns the provider's metadata
	Description() ContextProviderDescription

	// GetContextItems retrieves context items based on query
	// The query can be a file path, search term, or provider-specific format
	GetContextItems(ctx context.Context, query string, extras ContextExtras) ([]ContextItem, error)
}

// ContextProviderRegistry manages available context providers
type ContextProviderRegistry interface {
	// Register adds a provider to the registry
	Register(provider ContextProvider) error
	// Get returns a provider by name
	Get(name string) (ContextProvider, bool)
	// List returns all registered providers
	List() []ContextProvider
	// ListByType returns providers of a specific type
	ListByType(providerType ContextProviderType) []ContextProvider
}
