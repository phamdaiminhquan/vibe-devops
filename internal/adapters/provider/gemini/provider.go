package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const defaultPlaceholderKey = "YOUR_GEMINI_API_KEY_HERE"

type Provider struct {
	apiKey string
	model  string

	client *genai.Client
	gm     *genai.GenerativeModel
}

func New(apiKey, model string) (*Provider, error) {
	p := &Provider{apiKey: apiKey, model: model}
	if err := p.IsConfigured(context.Background()); err != nil {
		return nil, err
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("gemini model is not configured")
	}

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	p.client = client
	p.gm = client.GenerativeModel(model)
	return p, nil
}

func (p *Provider) Name() string { return "gemini" }

func (p *Provider) IsConfigured(ctx context.Context) error {
	if strings.TrimSpace(p.apiKey) == "" || p.apiKey == defaultPlaceholderKey {
		return fmt.Errorf("gemini API key is not configured")
	}
	return nil
}

func (p *Provider) Generate(ctx context.Context, req ports.GenerateRequest) (ports.GenerateResponse, error) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" && len(req.Messages) > 0 {
		var b strings.Builder
		for _, m := range req.Messages {
			b.WriteString(string(m.Role))
			b.WriteString(": ")
			b.WriteString(m.Content)
			b.WriteString("\n")
		}
		prompt = strings.TrimSpace(b.String())
	}

	if prompt == "" {
		return ports.GenerateResponse{}, fmt.Errorf("empty prompt")
	}

	// support overriding model/temperature if needed, but for now stick to defaults or configured

	resp, err := p.gm.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return ports.GenerateResponse{}, fmt.Errorf("failed to get completion from gemini: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return ports.GenerateResponse{}, fmt.Errorf("gemini returned no content")
	}

	if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		return ports.GenerateResponse{Text: string(txt)}, nil
	}

	return ports.GenerateResponse{}, fmt.Errorf("gemini response was not text")
}

func (p *Provider) StreamGenerate(ctx context.Context, req ports.GenerateRequest) (<-chan ports.StreamChunk, error) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" && len(req.Messages) > 0 {
		var b strings.Builder
		for _, m := range req.Messages {
			b.WriteString(string(m.Role))
			b.WriteString(": ")
			b.WriteString(m.Content)
			b.WriteString("\n")
		}
		prompt = strings.TrimSpace(b.String())
	}

	if prompt == "" {
		return nil, fmt.Errorf("empty prompt")
	}

	ch := make(chan ports.StreamChunk)

	iter := p.gm.GenerateContentStream(ctx, genai.Text(prompt))

	go func() {
		defer close(ch)
		for {
			resp, err := iter.Next()
			if err != nil {
				if err == iterator.Done {
					return
				}
				ch <- ports.StreamChunk{Error: err}
				return
			}

			if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
				if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
					ch <- ports.StreamChunk{Content: string(txt)}
				}
			}
		}
	}()

	return ch, nil
}

func (p *Provider) Close() error {
	if p.client == nil {
		return nil
	}
	return p.client.Close()
}

var _ ports.Provider = (*Provider)(nil)
