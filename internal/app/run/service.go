package run

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

type Service struct {
	provider ports.Provider
	logger   *slog.Logger
}

func NewService(provider ports.Provider, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{provider: provider, logger: logger}
}

type SuggestRequest struct {
	UserRequest string
	GOOS        string
}

func (s *Service) SuggestCommand(ctx context.Context, req SuggestRequest) (string, error) {
	if strings.TrimSpace(req.UserRequest) == "" {
		return "", fmt.Errorf("empty request")
	}

	prompt := buildPrompt(req.GOOS, req.UserRequest)
	s.logger.Debug("generating command", "provider", s.provider.Name())

	resp, err := s.provider.Generate(ctx, ports.GenerateRequest{Prompt: prompt})
	if err != nil {
		return "", err
	}

	cmd := sanitizeAIResponse(resp.Text)
	if strings.HasPrefix(cmd, "Error:") {
		return "", fmt.Errorf("AI returned an error: %s", cmd)
	}

	return cmd, nil
}
