package ports

import (
	"context"

	"github.com/phamdaiminhquan/vibe-devops/internal/domain/chat"
)

// Provider is the outbound port for an AI model provider.
// It is chat-first, but still supports the current single-prompt workflow.
type Provider interface {
	Name() string
	IsConfigured(ctx context.Context) error
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
	Close() error
}

type GenerateRequest struct {
	// Prompt is the raw text prompt for providers that are prompt-based.
	// For chat-based usage, Messages can be used instead.
	Prompt   string
	Messages []chat.Message
}

type GenerateResponse struct {
	Text string
}
