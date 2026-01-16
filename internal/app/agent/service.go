package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

type Service struct {
	provider        ports.Provider
	tools           []ports.Tool
	logger          *slog.Logger
	maxSteps        int
	contextRegistry ports.ContextProviderRegistry
}

func NewService(provider ports.Provider, tools []ports.Tool, logger *slog.Logger, maxSteps int) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	if maxSteps <= 0 {
		maxSteps = 15
	}
	return &Service{provider: provider, tools: tools, logger: logger, maxSteps: maxSteps}
}

// WithContextRegistry adds a context provider registry to the service
func (s *Service) WithContextRegistry(registry ports.ContextProviderRegistry) *Service {
	s.contextRegistry = registry
	return s
}

type SuggestRequest struct {
	UserRequest string
	GOOS        string
	// Transcript allows continuing an ongoing agent run (e.g., after executing a proposed command).
	// If empty, the service will start a new transcript.
	Transcript []string

	// OnProgress is called when the agent provides intermediate feedback
	OnProgress func(StepInfo)

	// OnToken is called when a token is generated (streaming)
	OnToken func(token string)
}

type StepInfo struct {
	Step    int
	Type    string // "thinking", "tool_call", "tool_output"
	Message string
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

	// Resolve @mentions to context items (only on first step)
	contextItems := s.resolveContextMentions(ctx, req.UserRequest)

	s.logger.InfoContext(ctx, "agent start", "request", req.UserRequest, "max_steps", s.maxSteps, "context_items", len(contextItems))

	for step := 0; step < s.maxSteps; step++ {
		// Callback: Thinking
		if req.OnProgress != nil {
			req.OnProgress(StepInfo{Step: step + 1, Type: "thinking", Message: "Analyzing request..."})
		}

		prompt := buildAgentPrompt(req.GOOS, req.UserRequest, transcript, s.tools, contextItems)
		s.logger.DebugContext(ctx, "agent generate", "provider", s.provider.Name(), "step", step+1)

		var responseText string

		if req.OnToken != nil {
			// Smart streaming: buffer tokens, parse JSON, stream only thought/explanation
			responseText = s.smartStreamGenerate(ctx, prompt, req.OnToken, step)
			if responseText == "" {
				// Fallback to non-streaming if smart stream failed
				resp, err := s.provider.Generate(ctx, ports.GenerateRequest{Prompt: prompt})
				if err != nil {
					s.logger.ErrorContext(ctx, "agent generate failed", "error", err, "step", step+1)
					return SuggestResponse{}, fmt.Errorf("agent generation failed at step %d: %w", step+1, err)
				}
				responseText = resp.Text
			}
		} else {
			resp, err := s.provider.Generate(ctx, ports.GenerateRequest{Prompt: prompt})
			if err != nil {
				s.logger.ErrorContext(ctx, "agent generate failed", "error", err, "step", step+1)
				return SuggestResponse{}, fmt.Errorf("agent generation failed at step %d: %w", step+1, err)
			}
			responseText = resp.Text
		}

		action, err := ParseAction(responseText)
		if err != nil {
			s.logger.WarnContext(ctx, "agent parse error", "error", err, "response", responseText)
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
			// Callback: Tool Call
			if req.OnProgress != nil {
				msg := fmt.Sprintf("Using tool: %s", action.Tool)
				if action.Thought != "" {
					msg = fmt.Sprintf("[%s] %s", action.Tool, action.Thought)
				}
				req.OnProgress(StepInfo{Step: step + 1, Type: "tool_call", Message: msg})
			}

			// Execute tool and continue loop
			toolOutput := s.executeTool(ctx, action, toolsByName)

			// Callback: Tool Output
			if req.OnProgress != nil {
				// Truncate output for UI if too long
				displayOut := toolOutput
				if len(displayOut) > 100 {
					displayOut = displayOut[:100] + "..."
				}
				req.OnProgress(StepInfo{Step: step + 1, Type: "tool_done", Message: "Result: " + displayOut})
			}

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
	return SuggestResponse{
		StepsUsed:  s.maxSteps,
		Transcript: transcript,
	}, fmt.Errorf("agent exceeded max steps (%d) without returning a command", s.maxSteps)
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

	// Create tool extras with empty callbacks for now
	extras := ports.ToolExtras{
		WorkDir: "",
	}

	result, err := tool.Run(ctx, action.Input, extras)
	if err != nil {
		s.logger.ErrorContext(ctx, "tool execution failed", "tool", toolName, "error", err)
		return fmt.Sprintf("ERROR: %v", err)
	}
	if result.IsError {
		s.logger.WarnContext(ctx, "tool returned error result", "tool", toolName)
		return "ERROR: " + result.Content
	}
	s.logger.DebugContext(ctx, "tool execution success", "tool", toolName)
	return result.Content
}

func (s *Service) mapToolsByName() map[string]ports.Tool {
	m := make(map[string]ports.Tool, len(s.tools))
	for _, t := range s.tools {
		if t != nil {
			m[t.Definition().Name] = t
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

// resolveContextMentions parses @mentions from user input and resolves them to context items
func (s *Service) resolveContextMentions(ctx context.Context, input string) []ports.ContextItem {
	if s.contextRegistry == nil {
		return nil
	}

	mentions := ParseContextMentions(input)
	if len(mentions) == 0 {
		return nil
	}

	var items []ports.ContextItem
	for _, m := range mentions {
		provider, ok := s.contextRegistry.Get(m.Provider)
		if !ok {
			s.logger.WarnContext(ctx, "unknown context provider", "provider", m.Provider)
			continue
		}

		extras := ports.ContextExtras{
			WorkDir:   ".",
			FullInput: input,
		}

		providerItems, err := provider.GetContextItems(ctx, m.Query, extras)
		if err != nil {
			s.logger.WarnContext(ctx, "context provider error", "provider", m.Provider, "query", m.Query, "error", err)
			continue
		}

		items = append(items, providerItems...)
	}

	return items
}

// smartStreamGenerate buffers streaming tokens, parses JSON, and streams only thought/explanation
// This provides a clean UX by not showing raw JSON structure to the user
func (s *Service) smartStreamGenerate(ctx context.Context, prompt string, onToken func(string), step int) string {
	streamCh, err := s.provider.StreamGenerate(ctx, ports.GenerateRequest{Prompt: prompt})
	if err != nil {
		s.logger.ErrorContext(ctx, "smart stream failed", "error", err, "step", step+1)
		return "" // Signal to use fallback
	}

	var fullBuffer strings.Builder
	var lastStreamedLen int

	// Track which fields we're currently inside for streaming
	inThought := false
	inExplanation := false
	thoughtStart := -1
	explanationStart := -1

	for chunk := range streamCh {
		if chunk.Error != nil {
			s.logger.ErrorContext(ctx, "smart stream chunk error", "error", chunk.Error)
			return "" // Signal to use fallback
		}

		fullBuffer.WriteString(chunk.Content)
		currentText := fullBuffer.String()

		// Skip thought streaming - it's already shown via onProgress callback
		// Only look for thought end to know when to start looking for explanation
		if !inThought && thoughtStart == -1 {
			if idx := strings.Index(currentText, `"thought":`); idx != -1 {
				afterKey := currentText[idx+10:]
				if qIdx := strings.Index(afterKey, `"`); qIdx != -1 {
					thoughtStart = idx + 10 + qIdx + 1
					inThought = true
				}
			}
		}

		if inThought && thoughtStart != -1 {
			// Just track when thought ends, don't stream it
			subset := currentText[thoughtStart:]
			endIdx := findUnescapedQuote(subset)
			if endIdx != -1 {
				inThought = false
				// Reset for explanation
			}
		}

		// Try to find and stream "explanation" content
		if !inExplanation && explanationStart == -1 && !inThought {
			if idx := strings.Index(currentText, `"explanation":`); idx != -1 {
				afterKey := currentText[idx+14:] // len(`"explanation":`) = 14
				if qIdx := strings.Index(afterKey, `"`); qIdx != -1 {
					explanationStart = idx + 14 + qIdx + 1
					inExplanation = true
					lastStreamedLen = 0
					onToken("\n") // Newline before explanation
				}
			}
		}

		if inExplanation && explanationStart != -1 {
			subset := currentText[explanationStart:]
			endIdx := findUnescapedQuote(subset)
			if endIdx != -1 {
				// Explanation complete
				explanationContent := subset[:endIdx]
				if len(explanationContent) > lastStreamedLen {
					// Unescape the content
					unescaped := unescapeJSON(explanationContent[lastStreamedLen:])
					onToken(unescaped)
				}
				inExplanation = false
			} else {
				// Still streaming explanation
				if len(subset) > lastStreamedLen {
					unescaped := unescapeJSON(subset[lastStreamedLen:])
					onToken(unescaped)
					lastStreamedLen = len(subset)
				}
			}
		}
	}

	return fullBuffer.String()
}

// findUnescapedQuote finds the first unescaped quote in a string
func findUnescapedQuote(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			// Check if escaped
			backslashes := 0
			for j := i - 1; j >= 0 && s[j] == '\\'; j-- {
				backslashes++
			}
			if backslashes%2 == 0 {
				return i
			}
		}
	}
	return -1
}

// unescapeJSON handles common JSON escape sequences
func unescapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\\`, `\`)
	return s
}
