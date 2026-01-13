package ai

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"github.com/phamdaiminhquan/vibe-devops/pkg/config"
	"google.golang.org/api/option"
)

// GeminiProvider implements the AI provider interface for Google Gemini.
type GeminiProvider struct {
	Cfg *config.Config
}

// NewGeminiProvider creates a new instance of the Gemini provider.
func NewGeminiProvider(cfg *config.Config) *GeminiProvider {
	return &GeminiProvider{Cfg: cfg}
}

// GetName returns the name of the provider.
func (p *GeminiProvider) GetName() string {
	return "gemini"
}

// IsConfigured checks if the provider has the necessary configuration.
func (p *GeminiProvider) IsConfigured() bool {
	return p.Cfg.AI.Gemini.APIKey != "" && p.Cfg.AI.Gemini.APIKey != "YOUR_API_KEY_HERE"
}

// GetCompletion gets a completion from the Gemini API.
func (p *GeminiProvider) GetCompletion(prompt string) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("gemini provider is not configured. Please add your API key to .vibe.yaml")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(p.Cfg.AI.Gemini.APIKey))
	if err != nil {
		return "", fmt.Errorf("failed to create a new Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(p.Cfg.AI.Gemini.Model)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("received an empty response from the API")
	}

	// Concatenate all parts of the response.
	var result string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			result += string(txt)
		}
	}

	return result, nil
}
