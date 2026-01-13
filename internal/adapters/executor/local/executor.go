package local

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

type Executor struct {
	goos string
}

func New() *Executor {
	return &Executor{goos: runtime.GOOS}
}

func NewForOS(goos string) *Executor {
	if goos == "" {
		goos = runtime.GOOS
	}
	return &Executor{goos: goos}
}

func (e *Executor) Run(ctx context.Context, spec ports.ExecSpec) (ports.ExecResult, error) {
	if spec.DryRun {
		return ports.ExecResult{ExitCode: 0, Duration: 0}, nil
	}

	shell := spec.Shell
	if shell.Name == "" {
		shell = defaultShell(e.goos)
	}

	execCtx := ctx
	var cancel context.CancelFunc
	if spec.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, spec.Timeout)
		defer cancel()
	}

	start := time.Now()
	cmd := exec.CommandContext(execCtx, shell.Name, append(shell.Args, spec.Command)...)

	if spec.Stdin != nil {
		cmd.Stdin = spec.Stdin
	} else {
		cmd.Stdin = os.Stdin
	}
	if spec.Stdout != nil {
		cmd.Stdout = spec.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}
	if spec.Stderr != nil {
		cmd.Stderr = spec.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}

	err := cmd.Run()
	dur := time.Since(start)
	if err == nil {
		return ports.ExecResult{ExitCode: 0, Duration: dur}, nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		return ports.ExecResult{ExitCode: exitErr.ExitCode(), Duration: dur}, fmt.Errorf("command failed with exit code %d", exitErr.ExitCode())
	}
	return ports.ExecResult{ExitCode: -1, Duration: dur}, fmt.Errorf("failed to execute command: %w", err)
}

func defaultShell(goos string) ports.ShellSpec {
	if goos == "windows" {
		return ports.ShellSpec{Name: "powershell", Args: []string{"-Command"}}
	}
	return ports.ShellSpec{Name: "sh", Args: []string{"-c"}}
}

var _ ports.Executor = (*Executor)(nil)
