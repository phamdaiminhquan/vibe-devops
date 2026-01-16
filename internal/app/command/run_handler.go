package command

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	ctxregistry "github.com/phamdaiminhquan/vibe-devops/internal/adapters/context"
	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/context/file"
	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/context/git"
	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/context/logs"
	ctxsystem "github.com/phamdaiminhquan/vibe-devops/internal/adapters/context/system"
	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/executor/local"
	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/tools/fs"
	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/tools/system"
	"github.com/phamdaiminhquan/vibe-devops/internal/app/agent"
	"github.com/phamdaiminhquan/vibe-devops/internal/app/bootstrap"
	"github.com/phamdaiminhquan/vibe-devops/internal/app/dependency"
	"github.com/phamdaiminhquan/vibe-devops/internal/app/run"
	"github.com/phamdaiminhquan/vibe-devops/internal/app/session"
	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

// RunFlags contains command configuration flags
type RunFlags struct {
	AgentMode           bool
	AgentMaxSteps       int
	SelfHeal            bool
	SelfHealMaxAttempts int
}

// RunHandler encapsulates the logic for the 'run' command
type RunHandler struct {
	Ctx   *bootstrap.ApplicationContext
	Sess  *session.Service
	Flags RunFlags
	Dep   *dependency.Manager
}

// NewRunHandler creates a new handler instance
func NewRunHandler(ctx *bootstrap.ApplicationContext, sess *session.Service, flags RunFlags) *RunHandler {
	return &RunHandler{
		Ctx:   ctx,
		Sess:  sess,
		Flags: flags,
		Dep:   dependency.NewManager(),
	}
}

// Handle executes the run logic
func (h *RunHandler) Handle(ctx context.Context, input string) error {
	// 0. Proactive Dependency Check
	h.checkDependencies(ctx)

	if h.Flags.AgentMode {
		return h.runAgentMode(ctx, input)
	}
	return h.runSingleShotMode(ctx, input)
}

func (h *RunHandler) checkDependencies(ctx context.Context) {
	results := h.Dep.VerifyAll(ctx)
	missingCount := 0

	for _, res := range results {
		if res.Status != dependency.StatusInstalled {
			missingCount++
		}
	}

	if missingCount == 0 {
		return
	}

	fmt.Println("\n⚠️  Dependency Check Warning:")
	for _, res := range results {
		if res.Status == dependency.StatusInstalled {
			continue
		}

		icon := "❌"
		if res.Status == dependency.StatusError {
			icon = "⚠️"
		}

		fmt.Printf("  %s %s: Not found or error\n", icon, res.Dependency.Name)
		if res.Dependency.InstallHint != "" {
			fmt.Printf("      👉 Fix: %s\n", res.Dependency.InstallHint)
		}
	}
	fmt.Println()
	// Non-blocking for now, just warn.
}

func (h *RunHandler) runAgentMode(ctx context.Context, input string) error {
	// Simple spinner/status
	onProgress := func(step agent.StepInfo) {
		switch step.Type {
		case "thinking":
			fmt.Printf("\r\033[K[VIBE] ⏳ Thinking... ")
		case "tool_call":
			fmt.Printf("\r\033[K[VIBE] 🛠  %s\n", step.Message)
		case "tool_done":
			// Optional: print result preview or just keep quiet to let next thinking overwrite
			// fmt.Printf("\r\033[K[VIBE] ✅ Completed. \n")
		}
	}

	var agentTranscript []string

	// Seed transcript from session if available
	if h.Sess != nil {
		combined, err := h.Sess.LoadCombined(session.ScopeBoth, "default") // TODO: Parameterize session name
		if err == nil {
			agentTranscript = h.Sess.BuildSeedTranscript(combined, input, runtime.GOOS)
		}
	}

	// Tool confirmation callback
	toolConfirm := func(cmd string) bool {
		fmt.Printf("\n[VIBE] Agent wants to run command:\n")
		fmt.Printf("   \033[1;33m%s\033[0m\n", cmd)
		fmt.Print("   Allow this one-time execution? (y/N) ")
		reader := bufio.NewReader(os.Stdin)
		in, _ := reader.ReadString('\n')
		return strings.ToLower(strings.TrimSpace(in)) == "y"
	}

	// Note: toolConfirm callback is now passed via ToolExtras during agent execution
	_ = toolConfirm // Callback will be used in future when agent supports OnConfirm

	tools := []ports.Tool{
		fs.NewListDirTool("."),
		fs.NewReadFileTool("."),
		fs.NewGrepTool("."),
		system.NewSafeShellTool(),
	}

	// Create context provider registry for @mentions
	contextRegistry := ctxregistry.NewRegistry()
	contextRegistry.Register(file.NewProvider("."))
	contextRegistry.Register(git.NewProvider("."))
	contextRegistry.Register(ctxsystem.NewProvider())
	contextRegistry.Register(logs.NewProvider("."))

	ag := agent.NewService(h.Ctx.Provider, tools, h.Ctx.Logger, h.Flags.AgentMaxSteps).
		WithContextRegistry(contextRegistry)

	// Loop to allow extending steps
	for {
		resp, err := ag.SuggestCommand(ctx, agent.SuggestRequest{
			UserRequest: input,
			GOOS:        runtime.GOOS,
			Transcript:  agentTranscript,
			OnProgress:  onProgress,
		})

		fmt.Printf("\r\033[K") // Clear spinner

		if err != nil {
			// Check for max steps error
			if strings.Contains(err.Error(), "agent exceeded max steps") {
				fmt.Printf("\n[VIBE] Agent stopped after %d steps to avoid infinite loops.\n", h.Flags.AgentMaxSteps)
				fmt.Printf("   Latest thought: It likely needs more time or is stuck.\n")
				fmt.Print("   Do you want to give it 10 more steps? (y/N) ")

				reader := bufio.NewReader(os.Stdin)
				in, _ := reader.ReadString('\n')
				if strings.ToLower(strings.TrimSpace(in)) == "y" {
					fmt.Println("🔄 Extending session...")
					agentTranscript = resp.Transcript // Resume from where we left off
					continue
				}
			}

			// Normal error handling
			errMsg := err.Error()
			if strings.Contains(errMsg, "API key not valid") || strings.Contains(errMsg, "API_KEY_INVALID") {
				fmt.Println("\n❌ Error: Invalid AI Provider API Key.")
				fmt.Println("👉 To fix this, run:")
				fmt.Println("   vibe config api-key \"YOUR_API_KEY\"")
				return fmt.Errorf("check your API key")
			}
			return fmt.Errorf("AI completion failed: %w", err)
		}

		if strings.TrimSpace(resp.Explanation) != "" {
			// New friendly format
			fmt.Printf("\n[VIBE] %s\n", resp.Explanation)
		}

		return h.executeAndHeal(ctx, resp.Command, resp.Transcript, input)
	}
}

func (h *RunHandler) runSingleShotMode(ctx context.Context, input string) error {
	runner := run.NewService(h.Ctx.Provider, h.Ctx.Logger)
	fmt.Println("🤖 Calling AI to generate command...")
	fmt.Println("ℹ️  Note: Vibe single-shot mode.")

	cmd, err := runner.SuggestCommand(ctx, run.SuggestRequest{UserRequest: input, GOOS: runtime.GOOS})
	if err != nil {
		return fmt.Errorf("AI completion failed: %w", err)
	}

	return h.executeAndHeal(ctx, cmd, nil, input)
}

func (h *RunHandler) executeAndHeal(ctx context.Context, cmd string, transcript []string, originalRequest string) error {
	// Auto-Confirmation for safe read-only commands
	if isSafeCommand(cmd) {
		fmt.Printf("\n✨ Vibe auto-executing safe command:\n  \033[1;36m%s\033[0m\n", cmd)
	} else {
		// Ask for confirmation for others
		if !h.askConfirmation(cmd) {
			return nil
		}
	}

	// Initial Execution
	fmt.Println("🚀 Executing command...")
	exec := local.NewForOS(runtime.GOOS)
	var stdoutBuf, stderrBuf strings.Builder
	spec := ports.ExecSpec{Command: cmd, Stdout: &stdoutBuf, Stderr: &stderrBuf}

	res, err := exec.Run(ctx, spec)

	// Output Feedback
	if err == nil && res.ExitCode == 0 {
		fmt.Println("\n✅ Command executed successfully.")
	} else {
		fmt.Printf("\n❌ Command failed (exit code %d).\n", res.ExitCode)
		if stderrBuf.Len() > 0 {
			fmt.Println("\n--- stderr ---")
			fmt.Println(stderrBuf.String())
		}
	}

	// Self-Heal Check
	shouldSelfHeal := h.Flags.AgentMode && h.Flags.SelfHeal && (looksLikeDiagnosticQuestion(originalRequest) || err != nil || res.ExitCode != 0)

	if !shouldSelfHeal {
		// Persist simple run
		if h.Sess != nil && len(transcript) > 0 {
			_ = h.Sess.UpdateBoth(ctx, "default", transcript)
		}
		if res.ExitCode != 0 {
			return fmt.Errorf("command failed with exit code %d", res.ExitCode)
		}
		return err
	}

	// Start Self-Healing Loop
	attempts := h.Flags.SelfHealMaxAttempts
	if attempts <= 0 {
		attempts = 3
	}

	if len(transcript) == 0 {
		transcript = []string{"USER_REQUEST: " + originalRequest, "GOOS: " + strings.TrimSpace(runtime.GOOS)}
	}
	transcript = append(transcript, "EXEC_COMMAND: "+cmd)
	transcript = append(transcript,
		fmt.Sprintf("EXEC_RESULT: exit_code=%d", res.ExitCode),
		"EXEC_STDOUT_TAIL: "+tailString(stdoutBuf.String(), 4000),
		"EXEC_STDERR_TAIL: "+tailString(stderrBuf.String(), 4000),
		"INSTRUCTION: Based on the execution result above, either answer the user's question (type=answer) or propose the next best command (type=done).",
	)

	// Tool setup for agent
	tools := []ports.Tool{
		fs.NewListDirTool("."),
		fs.NewReadFileTool("."),
		fs.NewGrepTool("."),
	}
	ag := agent.NewService(h.Ctx.Provider, tools, h.Ctx.Logger, h.Flags.AgentMaxSteps)

	for i := 0; i < attempts; i++ {
		resp, err := ag.SuggestCommand(ctx, agent.SuggestRequest{
			UserRequest: originalRequest,
			GOOS:        runtime.GOOS,
			Transcript:  transcript,
		})
		if err != nil {
			return fmt.Errorf("AI completion failed (self-heal): %w", err)
		}
		transcript = resp.Transcript

		if strings.TrimSpace(resp.Explanation) != "" {
			fmt.Println("\n🧠 Agent analysis:")
			fmt.Println(resp.Explanation)
		}

		if strings.TrimSpace(resp.Command) == "" {
			break
		}

		// Ask again
		if !h.askConfirmation(resp.Command) {
			break
		}

		// Execute again
		fmt.Println("🚀 Executing command...")
		stdoutBuf.Reset()
		stderrBuf.Reset()
		spec := ports.ExecSpec{Command: resp.Command, Stdout: &stdoutBuf, Stderr: &stderrBuf}
		res, err = exec.Run(ctx, spec)

		if err != nil || res.ExitCode != 0 {
			fmt.Printf("\n❌ Command failed (exit code %d).\n", res.ExitCode)
			if stderrBuf.Len() > 0 {
				fmt.Println("\n--- stderr ---")
				fmt.Println(stderrBuf.String())
			}
			transcript = append(transcript,
				"EXEC_COMMAND: "+resp.Command,
				fmt.Sprintf("EXEC_RESULT: exit_code=%d", res.ExitCode),
				"EXEC_STDOUT_TAIL: "+tailString(stdoutBuf.String(), 4000),
				"EXEC_STDERR_TAIL: "+tailString(stderrBuf.String(), 4000),
			)
			continue
		}

		fmt.Println("\n✅ Command executed successfully.")
		transcript = append(transcript,
			"EXEC_COMMAND: "+resp.Command,
			"EXEC_RESULT: exit_code=0",
			"EXEC_STDOUT_TAIL: "+tailString(stdoutBuf.String(), 4000),
			"EXEC_STDERR_TAIL: "+tailString(stderrBuf.String(), 4000),
			"INSTRUCTION: Continue until you can answer (type=answer) or stop if no more steps.",
		)
	}

	// Final Persist using session scope (simplification: assume 'default' name)
	if h.Sess != nil {
		_ = h.Sess.UpdateBoth(ctx, "default", transcript)
	}

	return nil
}

func (h *RunHandler) askConfirmation(cmd string) bool {
	fmt.Printf("\n✨ Vibe suggests the following command:\n\n")
	fmt.Printf("  \033[1;36m%s\033[0m\n\n", cmd)
	fmt.Print("Do you want to execute it? (y/N) ")

	reader := bufio.NewReader(os.Stdin)
	in, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(in)) != "y" {
		fmt.Println("Execution cancelled.")
		return false
	}
	return true
}

// Helpers

func looksLikeDiagnosticQuestion(s string) bool {
	s = strings.ToLower(s)
	return strings.Contains(s, "giải thích") || strings.Contains(s, "giai thich") ||
		strings.Contains(s, "tại sao") || strings.Contains(s, "tai sao") ||
		strings.Contains(s, "why") || strings.Contains(s, "debug") || strings.Contains(s, "not run") ||
		strings.Contains(s, "không chạy") || strings.Contains(s, "khong chay")
}

func tailString(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 {
		max = 4000
	}
	if len(s) <= max {
		return s
	}
	return s[len(s)-max:]
}

func isSafeCommand(cmd string) bool {
	cmd = strings.TrimSpace(strings.ToLower(cmd))
	safePrefixes := []string{
		"ls ", "ls",
		"find ", "find",
		"grep ", "grep",
		"cat ", "cat",
		"pwd ", "pwd",
		"echo ", "echo",
		"stat ", "stat",
		"whoami",
		"date",
	}
	for _, p := range safePrefixes {
		if cmd == p || strings.HasPrefix(cmd, p) {
			return true
		}
	}
	return false
}
