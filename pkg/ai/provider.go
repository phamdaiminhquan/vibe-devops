package ai

// Provider defines the interface for AI service providers
type Provider interface {
	// GetCompletion sends a prompt to the AI and returns the response
	GetCompletion(prompt string) (string, error)
	
	// GetName returns the name of the AI provider
	GetName() string
	
	// IsConfigured checks if the provider is properly configured
	IsConfigured() bool
}

// Config holds configuration for AI providers
type Config struct {
	// ProviderType specifies which AI provider to use (e.g., "openai", "anthropic", "local")
	ProviderType string
	
	// APIKey is the authentication key for the AI service
	APIKey string
	
	// Model specifies which model to use (e.g., "gpt-4", "claude-3")
	Model string
	
	// Endpoint is the API endpoint URL (optional, for custom endpoints)
	Endpoint string
	
	// MaxTokens limits the response length
	MaxTokens int
	
	// Temperature controls randomness in responses (0.0 - 1.0)
	Temperature float64
}

// Request represents a request to the AI provider
type Request struct {
	// Prompt is the user's input/question
	Prompt string
	
	// Context provides additional context for the AI
	Context string
	
	// SystemPrompt sets the AI's behavior/role
	SystemPrompt string
}

// Response represents a response from the AI provider
type Response struct {
	// Content is the AI's response text
	Content string
	
	// Model is the model that generated the response
	Model string
	
	// TokensUsed indicates how many tokens were consumed
	TokensUsed int
	
	// Provider is the name of the provider that generated the response
	Provider string
}
