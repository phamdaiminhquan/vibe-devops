package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

// Provider provides git-related context
type Provider struct {
	baseDir string
}

// NewProvider creates a new git context provider
func NewProvider(baseDir string) *Provider {
	return &Provider{baseDir: baseDir}
}

// Description returns the provider's metadata
func (p *Provider) Description() ports.ContextProviderDescription {
	return ports.ContextProviderDescription{
		Name:         "git",
		DisplayTitle: "Git Context",
		Description:  "Git information like status, diff, log. Use @git status|diff|log|branch",
		Type:         ports.ContextTypeGit,
	}
}

// GetContextItems retrieves git-related context
func (p *Provider) GetContextItems(ctx context.Context, query string, extras ports.ContextExtras) ([]ports.ContextItem, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		query = "status"
	}

	workDir := p.baseDir
	if extras.WorkDir != "" {
		workDir = extras.WorkDir
	}

	switch query {
	case "status":
		return p.getStatus(ctx, workDir)
	case "diff":
		return p.getDiff(ctx, workDir)
	case "log":
		return p.getLog(ctx, workDir)
	case "branch":
		return p.getBranch(ctx, workDir)
	default:
		return nil, fmt.Errorf("unknown git query: %s. Use: status, diff, log, branch", query)
	}
}

func (p *Provider) runGit(ctx context.Context, workDir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git %s failed: %w\n%s", args[0], err, string(out))
	}
	return string(out), nil
}

func (p *Provider) getStatus(ctx context.Context, workDir string) ([]ports.ContextItem, error) {
	out, err := p.runGit(ctx, workDir, "status", "--porcelain", "-b")
	if err != nil {
		return nil, err
	}

	return []ports.ContextItem{
		{
			Name:        "git-status",
			Description: "Current git status",
			Content:     out,
		},
	}, nil
}

func (p *Provider) getDiff(ctx context.Context, workDir string) ([]ports.ContextItem, error) {
	out, err := p.runGit(ctx, workDir, "diff", "--stat", "HEAD")
	if err != nil {
		// Try without HEAD for new repos
		out, err = p.runGit(ctx, workDir, "diff", "--stat")
		if err != nil {
			return nil, err
		}
	}

	return []ports.ContextItem{
		{
			Name:        "git-diff",
			Description: "Git diff summary",
			Content:     out,
		},
	}, nil
}

func (p *Provider) getLog(ctx context.Context, workDir string) ([]ports.ContextItem, error) {
	out, err := p.runGit(ctx, workDir, "log", "--oneline", "-n", "10")
	if err != nil {
		return nil, err
	}

	return []ports.ContextItem{
		{
			Name:        "git-log",
			Description: "Recent git commits (last 10)",
			Content:     out,
		},
	}, nil
}

func (p *Provider) getBranch(ctx context.Context, workDir string) ([]ports.ContextItem, error) {
	out, err := p.runGit(ctx, workDir, "branch", "-a")
	if err != nil {
		return nil, err
	}

	return []ports.ContextItem{
		{
			Name:        "git-branch",
			Description: "Git branches",
			Content:     out,
		},
	}, nil
}

var _ ports.ContextProvider = (*Provider)(nil)
