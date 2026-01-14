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

	toolsByName := s.mapToolsByName()
	transcript := s.initializeTranscript(req)

	s.logger.InfoContext(ctx, "agent start", "request", req.UserRequest, "max_steps", s.maxSteps)

	for step := 0; step < s.maxSteps; step++ {
		prompt := buildAgentPrompt(req.GOOS, req.UserRequest, transcript, s.tools)
		s.logger.DebugContext(ctx, "agent generate", "provider", s.provider.Name(), "step", step+1)

		resp, err := s.provider.Generate(ctx, ports.GenerateRequest{Prompt: prompt})
		if err != nil {
			s.logger.ErrorContext(ctx, "agent generate failed", "error", err, "step", step+1)
			return SuggestResponse{}, fmt.Errorf("agent generation failed at step %d: %w", step+1, err)
		}

		action, err := ParseAction(resp.Text)
		if err != nil {
			s.logger.WarnContext(ctx, "agent parse error", "error", err, "response", resp.Text)
			return SuggestResponse{}, fmt.Errorf("agent protocol parse error: %w", err)
		}

		s.logger.InfoContext(ctx, "agent action", "type", action.Type, "step", step+1)

		switch action.Type {
		case ActionTypeDone:
			cmd := strings.TrimSpace(action.Command)
			if cmd == "" {
				return SuggestResponse{}, fmt.Errorf("agent returned empty command")
			}
			s.logger.InfoContext(ctx, "agent done", "command", cmd)
			return SuggestResponse{
				Command:     cmd,
				Explanation: strings.TrimSpace(action.Explanation),
				StepsUsed:   step + 1,
				Transcript:  transcript,
			}, nil

		case ActionTypeAnswer:
			s.logger.InfoContext(ctx, "agent answer", "explanation", action.Explanation)
			return SuggestResponse{
				Command:     "",
				Explanation: strings.TrimSpace(action.Explanation),
				StepsUsed:   step + 1,
				Transcript:  transcript,
			}, nil

		case ActionTypeTool:
			// Execute tool and continue loop
			toolOutput := s.executeTool(ctx, action, toolsByName)
			transcript = append(transcript,
				fmt.Sprintf("TOOL_CALL: %s %s", action.Tool, strings.TrimSpace(string(action.Input))),
				fmt.Sprintf("TOOL_OUTPUT: %s", strings.TrimSpace(toolOutput)),
			)
			continue

		default:
			return SuggestResponse{}, fmt.Errorf("unsupported action type: %s", action.Type)
		}
	}

	s.logger.WarnContext(ctx, "agent max steps exceeded", "max_steps", s.maxSteps)
	return SuggestResponse{}, fmt.Errorf("agent exceeded max steps (%d) without returning a command", s.maxSteps)
}

func (s *Service) executeTool(ctx context.Context, action Action, toolsByName map[string]ports.Tool) string {
	toolName := strings.TrimSpace(action.Tool)
	tool, ok := toolsByName[toolName]
	if !ok {
		msg := fmt.Sprintf("agent requested unknown tool: %s", toolName)
		s.logger.WarnContext(ctx, "unknown tool", "tool", toolName)
		return "ERROR: " + msg
	}

	s.logger.DebugContext(ctx, "tool execution start", "tool", toolName)
	out, err := tool.Run(ctx, action.Input)
	if err != nil {
		s.logger.ErrorContext(ctx, "tool execution failed", "tool", toolName, "error", err)
		return fmt.Sprintf("ERROR: %v", err)
	}
	s.logger.DebugContext(ctx, "tool execution success", "tool", toolName)
	return out
}

func (s *Service) mapToolsByName() map[string]ports.Tool {
	m := make(map[string]ports.Tool, len(s.tools))
	for _, t := range s.tools {
		if t != nil {
			m[t.Name()] = t
		}
	}
	return m
}

func (s *Service) initializeTranscript(req SuggestRequest) []string {
	transcript := make([]string, 0, 32)
	if len(req.Transcript) > 0 {
		transcript = append(transcript, req.Transcript...)
	} else {
		transcript = append(transcript,
			"USER_REQUEST: "+req.UserRequest,
			"GOOS: "+strings.TrimSpace(req.GOOS),
		)
	}
	return transcript
}
