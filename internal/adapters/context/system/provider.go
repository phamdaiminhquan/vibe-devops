package system

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

// Provider provides system information as context
type Provider struct{}

// NewProvider creates a new system context provider
func NewProvider() *Provider {
	return &Provider{}
}

// Description returns the provider's metadata
func (p *Provider) Description() ports.ContextProviderDescription {
	return ports.ContextProviderDescription{
		Name:         "system",
		DisplayTitle: "System Info",
		Description:  "System information like OS, env vars. Use @system os|env|cwd",
		Type:         ports.ContextTypeSystem,
	}
}

// GetContextItems retrieves system-related context
func (p *Provider) GetContextItems(ctx context.Context, query string, extras ports.ContextExtras) ([]ports.ContextItem, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		query = "os"
	}

	switch query {
	case "os":
		return p.getOSInfo()
	case "env":
		return p.getEnvVars()
	case "cwd":
		return p.getCwd(extras)
	case "all":
		return p.getAll(extras)
	default:
		return nil, fmt.Errorf("unknown system query: %s. Use: os, env, cwd, all", query)
	}
}

func (p *Provider) getOSInfo() ([]ports.ContextItem, error) {
	hostname, _ := os.Hostname()

	content := fmt.Sprintf(`OS: %s
Architecture: %s
Hostname: %s
Go Version: %s
NumCPU: %d
`, runtime.GOOS, runtime.GOARCH, hostname, runtime.Version(), runtime.NumCPU())

	return []ports.ContextItem{
		{
			Name:        "system-os",
			Description: "Operating system information",
			Content:     content,
		},
	}, nil
}

func (p *Provider) getEnvVars() ([]ports.ContextItem, error) {
	// Only return safe, relevant env vars
	safeVars := []string{
		"PATH", "SHELL", "HOME", "USER", "LANG",
		"PWD", "OLDPWD", "TERM",
		"GOPATH", "GOROOT",
		"NODE_ENV", "PYTHON",
	}

	var content strings.Builder
	for _, key := range safeVars {
		if val := os.Getenv(key); val != "" {
			// Truncate long values
			if len(val) > 200 {
				val = val[:200] + "..."
			}
			content.WriteString(fmt.Sprintf("%s=%s\n", key, val))
		}
	}

	return []ports.ContextItem{
		{
			Name:        "system-env",
			Description: "Selected environment variables",
			Content:     content.String(),
		},
	}, nil
}

func (p *Provider) getCwd(extras ports.ContextExtras) ([]ports.ContextItem, error) {
	cwd := extras.WorkDir
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	return []ports.ContextItem{
		{
			Name:        "system-cwd",
			Description: "Current working directory",
			Content:     cwd,
		},
	}, nil
}

func (p *Provider) getAll(extras ports.ContextExtras) ([]ports.ContextItem, error) {
	var items []ports.ContextItem

	osItems, _ := p.getOSInfo()
	items = append(items, osItems...)

	cwdItems, _ := p.getCwd(extras)
	items = append(items, cwdItems...)

	return items, nil
}

var _ ports.ContextProvider = (*Provider)(nil)
