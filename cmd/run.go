package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/executor/local"
	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/provider/gemini"
	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/tools/fs"
	appAgent "github.com/phamdaiminhquan/vibe-devops/internal/app/agent"
	appRun "github.com/phamdaiminhquan/vibe-devops/internal/app/run"
	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
	"github.com/phamdaiminhquan/vibe-devops/pkg/config"
	"github.com/spf13/cobra"
)

var runAgentMode bool
var runAgentMaxSteps int
var runSelfHeal bool
var runSelfHealMaxAttempts int

var runCmd = &cobra.Command{
	Use:   "run [natural language request]",
	Short: "Execute a command based on a natural language request",
	Long: `Takes a natural language request, uses an AI provider to translate it into a shell command,
and executes it after user confirmation.`,
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE:         runCommand,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&runAgentMode, "agent", false, "Enable agent mode (model can request safe tools like reading files before proposing a command)")
	runCmd.Flags().IntVar(&runAgentMaxSteps, "agent-max-steps", 5, "Max tool steps in agent mode")
	runCmd.Flags().BoolVar(&runSelfHeal, "self-heal", true, "In agent mode, keep iterating after execution by reading command output and proposing next steps until an answer is reached")
	runCmd.Flags().IntVar(&runSelfHealMaxAttempts, "self-heal-max-attempts", 3, "Max execution/repair iterations in self-heal loop (agent mode only)")
}

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

func runCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// 1. Load config
	cfg, err := config.Load(".")
	if err != nil {
		return fmt.Errorf("could not load configuration from .vibe.yaml. Please run 'vibe init' first. Error: %w", err)
	}

	// 2. Instantiate AI provider
	var provider ports.Provider
	switch cfg.AI.Provider {
	case "gemini":
		p, err := gemini.New(cfg.AI.Gemini.APIKey, cfg.AI.Gemini.Model)
		if err != nil {
			return err
		}
		provider = p
	default:
		return fmt.Errorf("unsupported AI provider: %s", cfg.AI.Provider)
	}
	defer func() { _ = provider.Close() }()

	if err := provider.IsConfigured(ctx); err != nil {
		return fmt.Errorf("AI provider '%s' is not configured. Please add your credentials to .vibe.yaml", provider.Name())
	}

	// 3. Get user request and ask AI for a command suggestion
	userRequest := strings.Join(args, " ")

	var aiCommand string
	var agentTranscript []string
	if runAgentMode {
		fmt.Println("🤖 Calling AI (agent mode)...")
		tools := []ports.Tool{
			fs.NewListDirTool("."),
			fs.NewReadFileTool("."),
			fs.NewGrepTool("."),
		}
		agent := appAgent.NewService(provider, tools, slog.Default(), runAgentMaxSteps)
		resp, err := agent.SuggestCommand(ctx, appAgent.SuggestRequest{UserRequest: userRequest, GOOS: runtime.GOOS})
		if err != nil {
			return fmt.Errorf("AI completion failed: %w", err)
		}
		aiCommand = resp.Command
		agentTranscript = resp.Transcript
		if strings.TrimSpace(resp.Explanation) != "" {
			fmt.Println("\n🧾 Explanation:")
			fmt.Println(resp.Explanation)
		}
	} else {
		runner := appRun.NewService(provider, slog.Default())
		fmt.Println("🤖 Calling AI to generate command...")
		fmt.Println("ℹ️  Note: Currently, Vibe only interprets the command you send without any additional context.")
		cmd, err := runner.SuggestCommand(ctx, appRun.SuggestRequest{UserRequest: userRequest, GOOS: runtime.GOOS})
		if err != nil {
			return fmt.Errorf("AI completion failed: %w", err)
		}
		aiCommand = cmd
	}

	// 4. Ask for user confirmation
	fmt.Printf("\n✨ Vibe suggests the following command:\n\n")
	fmt.Printf("  \033[1;36m%s\033[0m\n\n", aiCommand) // Bold cyan
	fmt.Print("Do you want to execute it? (y/N) ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if strings.ToLower(input) != "y" {
		fmt.Println("Execution cancelled.")
		return nil
	}

	// 5. Execute the command
	fmt.Println("🚀 Executing command...")
	exec := local.NewForOS(runtime.GOOS)
	shouldSelfHeal := runAgentMode && runSelfHeal && looksLikeDiagnosticQuestion(userRequest)

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	spec := ports.ExecSpec{Command: aiCommand}
	if shouldSelfHeal {
		spec.Stdout = &stdoutBuf
		spec.Stderr = &stderrBuf
	}
	res, err := exec.Run(ctx, spec)
	if err != nil {
		if shouldSelfHeal {
			fmt.Printf("\n❌ Command failed (exit code %d).\n", res.ExitCode)
			if t := tailString(stderrBuf.String(), 4000); t != "" {
				fmt.Println("\n--- stderr (tail) ---")
				fmt.Println(t)
			}
		} else {
			if res.ExitCode > 0 {
				return fmt.Errorf("command failed with exit code %d", res.ExitCode)
			}
			return err
		}
	} else {
		fmt.Println("\n✅ Command executed successfully.")
	}

	// Agent self-heal loop: read execution output and keep iterating until the agent can answer.
	if shouldSelfHeal {
		attempts := runSelfHealMaxAttempts
		if attempts <= 0 {
			attempts = 3
		}

		// Seed transcript if missing (shouldn't happen in agent mode).
		if len(agentTranscript) == 0 {
			agentTranscript = []string{"USER_REQUEST: " + userRequest, "GOOS: " + strings.TrimSpace(runtime.GOOS)}
		}
		agentTranscript = append(agentTranscript, "EXEC_COMMAND: "+aiCommand)
		stdoutTail := tailString(stdoutBuf.String(), 4000)
		stderrTail := tailString(stderrBuf.String(), 4000)
		agentTranscript = append(agentTranscript,
			fmt.Sprintf("EXEC_RESULT: exit_code=%d", res.ExitCode),
			"EXEC_STDOUT_TAIL: "+stdoutTail,
			"EXEC_STDERR_TAIL: "+stderrTail,
			"INSTRUCTION: Based on the execution result above, either answer the user's question (type=answer) or propose the next best command (type=done).",
		)

		tools := []ports.Tool{
			fs.NewListDirTool("."),
			fs.NewReadFileTool("."),
			fs.NewGrepTool("."),
		}
		agent := appAgent.NewService(provider, tools, slog.Default(), runAgentMaxSteps)

		for i := 0; i < attempts; i++ {
			resp, err := agent.SuggestCommand(ctx, appAgent.SuggestRequest{UserRequest: userRequest, GOOS: runtime.GOOS, Transcript: agentTranscript})
			if err != nil {
				return fmt.Errorf("AI completion failed (self-heal): %w", err)
			}
			agentTranscript = resp.Transcript

			if strings.TrimSpace(resp.Explanation) != "" {
				fmt.Println("\n🧠 Agent analysis:")
				fmt.Println(resp.Explanation)
			}

			if strings.TrimSpace(resp.Command) == "" {
				break
			}

			fmt.Printf("\n✨ Next suggested command:\n\n")
			fmt.Printf("  \033[1;36m%s\033[0m\n\n", resp.Command)
			fmt.Print("Do you want to execute it? (y/N) ")

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if strings.ToLower(input) != "y" {
				break
			}

			fmt.Println("🚀 Executing command...")
			stdoutBuf.Reset()
			stderrBuf.Reset()
			spec := ports.ExecSpec{Command: resp.Command, Stdout: &stdoutBuf, Stderr: &stderrBuf}
			res, err := exec.Run(ctx, spec)
			if err != nil {
				fmt.Printf("\n❌ Command failed (exit code %d).\n", res.ExitCode)
				if t := tailString(stderrBuf.String(), 4000); t != "" {
					fmt.Println("\n--- stderr (tail) ---")
					fmt.Println(t)
				}
				// Feed failure back into transcript and continue.
				agentTranscript = append(agentTranscript,
					"EXEC_COMMAND: "+resp.Command,
					fmt.Sprintf("EXEC_RESULT: exit_code=%d", res.ExitCode),
					"EXEC_STDOUT_TAIL: "+tailString(stdoutBuf.String(), 4000),
					"EXEC_STDERR_TAIL: "+tailString(stderrBuf.String(), 4000),
				)
				continue
			}

			fmt.Println("\n✅ Command executed successfully.")
			stdoutTail := tailString(stdoutBuf.String(), 4000)
			stderrTail := tailString(stderrBuf.String(), 4000)
			agentTranscript = append(agentTranscript,
				"EXEC_COMMAND: "+resp.Command,
				"EXEC_RESULT: exit_code=0",
				"EXEC_STDOUT_TAIL: "+stdoutTail,
				"EXEC_STDERR_TAIL: "+stderrTail,
				"INSTRUCTION: Continue until you can answer (type=answer) or stop if no more steps.",
			)
		}
	}

	return nil
}
