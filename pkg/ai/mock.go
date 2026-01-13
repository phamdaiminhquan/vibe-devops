package ai

import "fmt"

// MockProvider is a mock implementation of the Provider interface for testing
type MockProvider struct {
	name       string
	configured bool
}

// NewMockProvider creates a new mock AI provider
func NewMockProvider() *MockProvider {
	return &MockProvider{
		name:       "mock",
		configured: true,
	}
}

// GetCompletion returns a mock response
func (m *MockProvider) GetCompletion(prompt string) (string, error) {
	if !m.configured {
		return "", fmt.Errorf("provider not configured")
	}
	
	// Return a simple mock response
	return fmt.Sprintf("Mock AI response to: %s", prompt), nil
}

// GetName returns the provider name
func (m *MockProvider) GetName() string {
	return m.name
}

// IsConfigured checks if the provider is configured
func (m *MockProvider) IsConfigured() bool {
	return m.configured
}
