package bootstrap

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/provider/gemini"
	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/sessionstore/jsonfile"
	"github.com/phamdaiminhquan/vibe-devops/internal/app/session"
	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
	"github.com/phamdaiminhquan/vibe-devops/pkg/config"
)

// ApplicationContext holds dependencies for the application
type ApplicationContext struct {
	Config   *config.Config
	Provider ports.Provider
	Logger   *slog.Logger
}

// SessionConfig holds configuration for session management
type SessionConfig struct {
	Name      string
	Scope     string
	Resume    bool
	NoSession bool
	Budget    session.Budget
}

// Initialize loads config and sets up dependencies
func Initialize(ctx context.Context) (*ApplicationContext, error) {
	// 1. Load config
	cfg, err := config.Load(".")
	if err != nil {
		return nil, fmt.Errorf("could not load configuration from .vibe.yaml. Please run 'vibe init' first. Error: %w", err)
	}

	// 2. Setup Logger - use quiet logger for clean CLI output
	// Set VIBE_DEBUG=1 to enable debug logging
	var logger *slog.Logger
	if os.Getenv("VIBE_DEBUG") == "1" {
		logger = slog.Default()
	} else {
		// Discard all logs for clean CLI output
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	// 3. Instantiate AI provider
	var provider ports.Provider
	switch cfg.AI.Provider {
	case "gemini":
		p, err := gemini.New(cfg.AI.Gemini.APIKey, cfg.AI.Gemini.Model)
		if err != nil {
			return nil, err
		}
		provider = p
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", cfg.AI.Provider)
	}

	return &ApplicationContext{
		Config:   cfg,
		Provider: provider,
		Logger:   logger,
	}, nil
}

// InitializeSessionService creates the session service based on config
func InitializeSessionService(provider ports.Provider, cfg SessionConfig) *session.Service {
	if cfg.NoSession {
		return session.NewService(provider, nil, nil, cfg.Budget)
	}

	home, _ := os.UserHomeDir()
	globalDir := filepath.Join(home, ".vibe")
	projectDir := filepath.Join(".", ".vibe")

	// Create stores
	projectStore := jsonfile.New(projectDir)
	globalStore := jsonfile.New(globalDir)

	// Based on scope, decide which stores the service should actually write to (or read from)
	// The Service logic handles "Scope" during Load/Update, but we pass both stores if available.
	// For now, simpler is better: pass both stores, the service manages logic.
	return session.NewService(provider, projectStore, globalStore, cfg.Budget)
}
