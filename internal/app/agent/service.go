package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

type Service struct {
	provider ports.Provider
	tools    []ports.Tool
	logger   *slog.Logger
	maxSteps int
}

func NewService(provider ports.Provider, tools []ports.Tool, logger *slog.Logger, maxSteps int) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	if maxSteps <= 0 {
		maxSteps = 5
	}
	return &Service{provider: provider, tools: tools, logger: logger, maxSteps: maxSteps}
}

type SuggestRequest struct {
	UserRequest string
	GOOS        string
	// Transcript allows continuing an ongoing agent run (e.g., after executing a proposed command).
	// If empty, the service will start a new transcript.
	Transcript []string
}

type SuggestResponse struct {
	Command     string
	Explanation string
	StepsUsed   int
	Transcript  []string
}

func (s *Service) SuggestCommand(ctx context.Context, req SuggestRequest) (SuggestResponse, error) {
	if strings.TrimSpace(req.UserRequest) == "" {
		return SuggestResponse{}, fmt.Errorf("empty request")
	}

	toolsByName := make(map[string]ports.Tool, len(s.tools))
	for _, t := range s.tools {
		if t == nil {
			continue
		}
		toolsByName[t.Name()] = t
	}

	transcript := make([]string, 0, 32)
	if len(req.Transcript) > 0 {
		transcript = append(transcript, req.Transcript...)
	} else {
		transcript = append(transcript,
			"USER_REQUEST: "+req.UserRequest,
			"GOOS: "+strings.TrimSpace(req.GOOS),
		)
	}

	for step := 0; step < s.maxSteps; step++ {
		prompt := buildAgentPrompt(req.GOOS, req.UserRequest, transcript, s.tools)
		s.logger.Debug("agent generate", "provider", s.provider.Name(), "step", step+1)

		resp, err := s.provider.Generate(ctx, ports.GenerateRequest{Prompt: prompt})
		if err != nil {
			return SuggestResponse{}, err
		}

		action, err := ParseAction(resp.Text)
		if err != nil {
			return SuggestResponse{}, fmt.Errorf("agent protocol parse error: %w", err)
		}

		switch action.Type {
		case ActionTypeDone:
			cmd := strings.TrimSpace(action.Command)
			if cmd == "" {
				return SuggestResponse{}, fmt.Errorf("agent returned empty command")
			}
			return SuggestResponse{Command: cmd, Explanation: strings.TrimSpace(action.Explanation), StepsUsed: step, Transcript: transcript}, nil

		case ActionTypeAnswer:
			return SuggestResponse{Command: "", Explanation: strings.TrimSpace(action.Explanation), StepsUsed: step, Transcript: transcript}, nil

		case ActionTypeTool:
			toolName := strings.TrimSpace(action.Tool)
			tool, ok := toolsByName[toolName]
			if !ok {
				return SuggestResponse{}, fmt.Errorf("agent requested unknown tool: %s", toolName)
			}
			out, err := tool.Run(ctx, action.Input)
			if err != nil {
				out = fmt.Sprintf("ERROR: %v", err)
			}
			transcript = append(transcript,
				fmt.Sprintf("TOOL_CALL: %s %s", toolName, strings.TrimSpace(string(action.Input))),
				fmt.Sprintf("TOOL_OUTPUT: %s", strings.TrimSpace(out)),
			)
			continue

		default:
			return SuggestResponse{}, fmt.Errorf("unsupported action type: %s", action.Type)
		}
	}

	return SuggestResponse{}, fmt.Errorf("agent exceeded max steps (%d) without returning a command", s.maxSteps)
}
