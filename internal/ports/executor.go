package ports

import (
	"context"
	"io"
	"time"
)

// Executor is the outbound port for executing commands.
// Implementations can target local shell, SSH, Docker, etc.
type Executor interface {
	Run(ctx context.Context, spec ExecSpec) (ExecResult, error)
}

type ShellSpec struct {
	Name string
	Args []string
}

type ExecSpec struct {
	Command string
	Shell   ShellSpec

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Timeout time.Duration
	DryRun  bool
}

type ExecResult struct {
	ExitCode int
	Duration time.Duration
}
