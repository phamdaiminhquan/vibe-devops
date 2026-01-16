package ollama

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ollama/ollama/api"
	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

type Provider struct {
	client *api.Client
	model  string
}

// New creates a new Ollama provider
// host example: "http://localhost:11434"
func New(host, model string) (*Provider, error) {
	if host == "" {
		host = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3"
	}

	u, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("invalid ollama host: %w", err)
	}

	client := api.NewClient(u, http.DefaultClient)
	return &Provider{
		client: client,
		model:  model,
	}, nil
}

func (p *Provider) Name() string {
	return "ollama"
}

func (p *Provider) IsConfigured(ctx context.Context) error {
	// Ping or List models to check connection
	_, err := p.client.List(ctx)
	if err != nil {
		return fmt.Errorf("ollama connection failed: %w", err)
	}
	return nil
}

func (p *Provider) Generate(ctx context.Context, req ports.GenerateRequest) (ports.GenerateResponse, error) {
	model := p.model
	if req.Model != "" {
		model = req.Model
	}

	messages := p.convertToOllamaMessages(req)

	var responseText strings.Builder

	// Use Chat API (supports messages)
	reqChat := &api.ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   new(bool), // false
	}
	*reqChat.Stream = false

	err := p.client.Chat(ctx, reqChat, func(resp api.ChatResponse) error {
		responseText.WriteString(resp.Message.Content)
		return nil
	})

	if err != nil {
		return ports.GenerateResponse{}, fmt.Errorf("ollama generate error: %w", err)
	}

	return ports.GenerateResponse{
		Text: responseText.String(),
	}, nil
}

func (p *Provider) StreamGenerate(ctx context.Context, req ports.GenerateRequest) (<-chan ports.StreamChunk, error) {
	model := p.model
	if req.Model != "" {
		model = req.Model
	}

	messages := p.convertToOllamaMessages(req)

	ch := make(chan ports.StreamChunk)

	reqChat := &api.ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   new(bool), // true
	}
	*reqChat.Stream = true

	go func() {
		defer close(ch)

		err := p.client.Chat(ctx, reqChat, func(resp api.ChatResponse) error {
			if resp.Message.Content != "" {
				ch <- ports.StreamChunk{Content: resp.Message.Content}
			}
			if resp.Done {
				// isLast handled by channel close
			}
			return nil
		})

		if err != nil {
			ch <- ports.StreamChunk{Error: err}
		}
	}()

	return ch, nil
}

func (p *Provider) convertToOllamaMessages(req ports.GenerateRequest) []api.Message {
	var messages []api.Message

	if len(req.Messages) > 0 {
		for _, m := range req.Messages {
			role := "user"
			switch m.Role {
			case "system":
				role = "system"
			case "assistant":
				role = "assistant"
			}
			messages = append(messages, api.Message{
				Role:    role,
				Content: m.Content,
			})
		}
	} else if req.Prompt != "" {
		messages = append(messages, api.Message{
			Role:    "user",
			Content: req.Prompt,
		})
	}
	return messages
}

func (p *Provider) Close() error {
	return nil
}

var _ ports.Provider = (*Provider)(nil)
