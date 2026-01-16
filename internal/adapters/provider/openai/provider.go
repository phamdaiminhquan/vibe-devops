package openai

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
	gopenai "github.com/sashabaranov/go-openai"
)

// Provider implements ports.Provider for OpenAI
type Provider struct {
	client *gopenai.Client
	model  string
}

// New creates a new OpenAI provider
func New(apiKey, model string) *Provider {
	if model == "" {
		model = gopenai.GPT4o
	}
	config := gopenai.DefaultConfig(apiKey)
	client := gopenai.NewClientWithConfig(config)
	return &Provider{
		client: client,
		model:  model,
	}
}

func (p *Provider) Name() string {
	return "openai"
}

func (p *Provider) IsConfigured(ctx context.Context) error {
	// Simple check: do we have a client?
	// Proper check would be making a simple API call, but that might be expensive/slow
	if p.client == nil {
		return fmt.Errorf("openai client not initialized")
	}
	return nil
}

func (p *Provider) Generate(ctx context.Context, req ports.GenerateRequest) (ports.GenerateResponse, error) {
	messages := p.convertToOpenAIMessages(req)

	model := p.model
	if req.Model != "" {
		model = req.Model
	}

	resp, err := p.client.CreateChatCompletion(
		ctx,
		gopenai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	)

	if err != nil {
		return ports.GenerateResponse{}, fmt.Errorf("openai generate error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return ports.GenerateResponse{}, fmt.Errorf("openai returned no choices")
	}

	return ports.GenerateResponse{
		Text: resp.Choices[0].Message.Content,
	}, nil
}

func (p *Provider) StreamGenerate(ctx context.Context, req ports.GenerateRequest) (<-chan ports.StreamChunk, error) {
	messages := p.convertToOpenAIMessages(req)

	model := p.model
	if req.Model != "" {
		model = req.Model
	}

	stream, err := p.client.CreateChatCompletionStream(
		ctx,
		gopenai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("openai stream creation error: %w", err)
	}

	ch := make(chan ports.StreamChunk)

	go func() {
		defer stream.Close()
		defer close(ch)

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				ch <- ports.StreamChunk{Error: err}
				return
			}

			if len(response.Choices) > 0 {
				content := response.Choices[0].Delta.Content
				if content != "" {
					ch <- ports.StreamChunk{Content: content}
				}
			}
		}
	}()

	return ch, nil
}

func (p *Provider) convertToOpenAIMessages(req ports.GenerateRequest) []gopenai.ChatCompletionMessage {
	var messages []gopenai.ChatCompletionMessage

	// If Messages are provided, use them
	if len(req.Messages) > 0 {
		for _, m := range req.Messages {
			role := gopenai.ChatMessageRoleUser
			switch m.Role {
			case "system":
				role = gopenai.ChatMessageRoleSystem
			case "assistant":
				role = gopenai.ChatMessageRoleAssistant
			}
			messages = append(messages, gopenai.ChatCompletionMessage{
				Role:    role,
				Content: m.Content,
			})
		}
	} else if req.Prompt != "" {
		// Fallback to single prompt
		messages = append(messages, gopenai.ChatCompletionMessage{
			Role:    gopenai.ChatMessageRoleUser,
			Content: req.Prompt,
		})
	}

	return messages
}

func (p *Provider) Close() error {
	return nil
}

var _ ports.Provider = (*Provider)(nil)
