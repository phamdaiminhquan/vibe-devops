package session

import (
	"context"
	"fmt"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

type Scope string

const (
	ScopeNone    Scope = "none"
	ScopeProject Scope = "project"
	ScopeGlobal  Scope = "global"
	ScopeBoth    Scope = "both"
)

type Budget struct {
	MaxRecentLines int
	MaxRecentChars int
}

type Service struct {
	provider     ports.Provider
	projectStore ports.SessionStore
	globalStore  ports.SessionStore
	budget       Budget
}

func NewService(provider ports.Provider, projectStore, globalStore ports.SessionStore, budget Budget) *Service {
	if budget.MaxRecentLines <= 0 {
		budget.MaxRecentLines = 40
	}
	if budget.MaxRecentChars <= 0 {
		budget.MaxRecentChars = 8000
	}
	return &Service{
		provider:     provider,
		projectStore: projectStore,
		globalStore:  globalStore,
		budget:       budget,
	}
}

type CombinedContext struct {
	GlobalSummary  string
	ProjectSummary string
	Recent         []string
}

func (s *Service) LoadCombined(scope Scope, sessionName string) (CombinedContext, error) {
	var out CombinedContext

	if scope == ScopeNone {
		return out, nil
	}

	if (scope == ScopeGlobal || scope == ScopeBoth) && s.globalStore != nil {
		st, err := s.globalStore.Load(sessionName)
		if err != nil {
			return out, err
		}
		out.GlobalSummary = strings.TrimSpace(st.Summary)
		out.Recent = append(out.Recent, st.Recent...)
	}

	if (scope == ScopeProject || scope == ScopeBoth) && s.projectStore != nil {
		st, err := s.projectStore.Load(sessionName)
		if err != nil {
			return out, err
		}
		out.ProjectSummary = strings.TrimSpace(st.Summary)
		out.Recent = append(out.Recent, st.Recent...)
	}

	out.Recent = trimRecent(out.Recent, s.budget.MaxRecentLines, s.budget.MaxRecentChars)
	return out, nil
}

func (s *Service) BuildSeedTranscript(ctx CombinedContext, userRequest, goos string) []string {
	var t []string
	if ctx.GlobalSummary != "" {
		t = append(t, "GLOBAL_SESSION_SUMMARY: "+ctx.GlobalSummary)
	}
	if ctx.ProjectSummary != "" {
		t = append(t, "PROJECT_SESSION_SUMMARY: "+ctx.ProjectSummary)
	}
	if len(ctx.Recent) > 0 {
		t = append(t, "SESSION_RECENT:")
		t = append(t, ctx.Recent...)
	}
	// Always include fresh request markers for consistency.
	t = append(t,
		"USER_REQUEST: "+strings.TrimSpace(userRequest),
		"GOOS: "+strings.TrimSpace(goos),
	)
	return t
}

// UpdateBoth persists session state to project + global stores (Option C).
// For global scope, we store only compact signals by summarizing aggressively.
func (s *Service) UpdateBoth(ctx context.Context, sessionName string, newLines []string) error {
	newLines = redactLines(newLines)

	if s.projectStore != nil {
		st, err := s.projectStore.Load(sessionName)
		if err != nil {
			return err
		}
		st, err = s.mergeAndMaybeSummarize(ctx, st, newLines, "project")
		if err != nil {
			return err
		}
		if err := s.projectStore.Save(sessionName, st); err != nil {
			return err
		}
	}

	if s.globalStore != nil {
		st, err := s.globalStore.Load(sessionName)
		if err != nil {
			return err
		}
		// Global store: keep fewer recent lines and summarize harder.
		prevBudget := s.budget
		s.budget.MaxRecentLines = min(prevBudget.MaxRecentLines, 12)
		s.budget.MaxRecentChars = min(prevBudget.MaxRecentChars, 2000)
		st, err = s.mergeAndMaybeSummarize(ctx, st, newLines, "global")
		// restore
		s.budget = prevBudget
		if err != nil {
			return err
		}
		if err := s.globalStore.Save(sessionName, st); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) mergeAndMaybeSummarize(ctx context.Context, st *ports.SessionState, newLines []string, label string) (*ports.SessionState, error) {
	if st == nil {
		st = &ports.SessionState{Version: 1}
	}
	combined := append([]string{}, st.Recent...)
	combined = append(combined, newLines...)

	combined = trimRecent(combined, 200, 20000)
	st.Recent = trimRecent(combined, s.budget.MaxRecentLines, s.budget.MaxRecentChars)

	// If we had to drop a lot, ask the model to update the rolling summary.
	if len(combined) > len(st.Recent) && s.provider != nil {
		toSummarize := combined
		if len(st.Recent) > 0 {
			toSummarize = combined[:max(0, len(combined)-len(st.Recent))]
		}
		updated, err := s.summarize(ctx, st.Summary, toSummarize, label)
		if err == nil {
			st.Summary = strings.TrimSpace(updated)
		}
	}

	return st, nil
}

func (s *Service) summarize(ctx context.Context, existing string, lines []string, label string) (string, error) {
	if len(lines) == 0 {
		return existing, nil
	}

	var b strings.Builder
	b.WriteString("You are a summarizer for a CLI agent session. Update the rolling summary with the new evidence.\n")
	b.WriteString("Keep it compact and actionable (max ~12 lines). Prefer stable facts, errors, decisions, and next steps.\n")
	b.WriteString("Do NOT include secrets. If something looks like a key/token, replace with [REDACTED].\n")
	b.WriteString("Output plain text only.\n\n")
	b.WriteString("SESSION_SCOPE: ")
	b.WriteString(label)
	b.WriteString("\n\n")
	b.WriteString("EXISTING_SUMMARY:\n")
	b.WriteString(strings.TrimSpace(existing))
	b.WriteString("\n\n")
	b.WriteString("NEW_EVENTS:\n")
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		b.WriteString("- ")
		b.WriteString(ln)
		b.WriteString("\n")
	}

	resp, err := s.provider.Generate(ctx, ports.GenerateRequest{Prompt: b.String()})
	if err != nil {
		return existing, fmt.Errorf("summarize failed: %w", err)
	}
	return strings.TrimSpace(resp.Text), nil
}

func trimRecent(lines []string, maxLines, maxChars int) []string {
	if maxLines <= 0 {
		maxLines = 40
	}
	if maxChars <= 0 {
		maxChars = 8000
	}

	// Keep most recent lines.
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}

	// Enforce char budget (from most recent backwards).
	chars := 0
	start := 0
	for i := len(lines) - 1; i >= 0; i-- {
		chars += len(lines[i]) + 1
		if chars > maxChars {
			start = i + 1
			break
		}
	}
	if start > 0 && start < len(lines) {
		lines = lines[start:]
	}
	return lines
}

func redactLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		l := ln
		lower := strings.ToLower(l)
		if strings.Contains(lower, "apikey") || strings.Contains(lower, "api-key") || strings.Contains(lower, "token") {
			l = "[REDACTED_LINE]"
		}
		out = append(out, l)
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
