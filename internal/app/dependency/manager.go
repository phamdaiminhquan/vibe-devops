package dependency

import (
	"context"
	"os/exec"
)

// Status represents the state of a dependency
type Status string

const (
	StatusInstalled Status = "installed"
	StatusMissing   Status = "missing"
	StatusError     Status = "error"
)

// Dependency represents a required external tool
type Dependency struct {
	Name        string
	Binary      string
	CheckArgs   []string // e.g., ["--version"]
	InstallHint string
	Critical    bool // If true, finding this missing is a serious warning
}

// Result holds the check result
type Result struct {
	Dependency Dependency
	Status     Status
	Version    string
	Error      error
}

// Manager handles dependency checks
type Manager struct {
	dependencies []Dependency
}

// NewManager creates a manager with default dependencies
func NewManager() *Manager {
	return &Manager{
		dependencies: []Dependency{
			{
				Name:        "Git",
				Binary:      "git",
				CheckArgs:   []string{"--version"},
				InstallHint: "Install from https://git-scm.com/downloads",
				Critical:    true,
			},
			{
				Name:        "Docker",
				Binary:      "docker",
				CheckArgs:   []string{"--version"},
				InstallHint: "Install Docker Desktop: https://www.docker.com/products/docker-desktop",
				Critical:    true,
			},
		},
	}
}

// VerifyAll checks all registered dependencies
func (m *Manager) VerifyAll(ctx context.Context) []Result {
	results := make([]Result, 0, len(m.dependencies))
	for _, dep := range m.dependencies {
		results = append(results, m.checkOne(ctx, dep))
	}
	return results
}

func (m *Manager) checkOne(ctx context.Context, dep Dependency) Result {
	path, err := exec.LookPath(dep.Binary)
	if err != nil {
		return Result{Dependency: dep, Status: StatusMissing}
	}

	// Try running version check
	cmd := exec.CommandContext(ctx, path, dep.CheckArgs...)
	out, err := cmd.Output()
	if err != nil {
		return Result{Dependency: dep, Status: StatusError, Error: err}
	}

	return Result{Dependency: dep, Status: StatusInstalled, Version: string(out)}
}
