package ai

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"github.com/phamdaiminhquan/vibe-devops/pkg/config"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GeminiProvider implements the Provider interface for Google's Gemini models.
type GeminiProvider struct {
	client *genai.GenerativeModel
	cfg    config.GeminiConfig
}

// NewGeminiProvider creates a new instance of the GeminiProvider.
// It takes a GeminiConfig struct and returns a configured provider.
func NewGeminiProvider(cfg config.GeminiConfig) (*GeminiProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("gemini API key is not configured")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := client.GenerativeModel(cfg.Model)

	return &GeminiProvider{
		client: model,
		cfg:    cfg,
	}, nil
}

// GetName returns the name of the provider.
func (p *GeminiProvider) GetName() string {
	return "gemini"
}

// GetCompletion sends a prompt to the Gemini API and returns the response.
func (p *GeminiProvider) GetCompletion(prompt string) (string, error) {
	ctx := context.Background()
	resp, err := p.client.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to get completion from gemini: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned no content")
	}

	// Extract text from the first candidate
	if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		return string(txt), nil
	}

	return "", fmt.Errorf("gemini response was not text")
}

// IsConfigured checks if the provider has the necessary configuration.
func (p *GeminiProvider) IsConfigured() bool {
	return p.cfg.APIKey != "" && p.cfg.APIKey != "YOUR_GEMINI_API_KEY_HERE"
}

// GetGeminiModels returns a list of available Gemini models for the given API key.
func GetGeminiModels(apiKey string) ([]string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}
	defer client.Close()

	iter := client.ListModels(ctx)
	var models []string
	for {
		m, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list models: %w", err)
		}

		// Check if the model supports generateContent
		if m.SupportedGenerationMethods != nil {
			for _, method := range m.SupportedGenerationMethods {
				if method == "generateContent" {
					models = append(models, m.Name)
					break
				}
			}
		}
	}
	return models, nil
}
