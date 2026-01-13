package cmd

import (
	"bufio"
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

var runCmd = &cobra.Command{
	Use:   "run [natural language request]",
	Short: "Execute a command based on a natural language request",
	Long: `Takes a natural language request, uses an AI provider to translate it into a shell command,
and executes it after user confirmation.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCommand,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&runAgentMode, "agent", false, "Enable agent mode (model can request safe tools like reading files before proposing a command)")
	runCmd.Flags().IntVar(&runAgentMaxSteps, "agent-max-steps", 5, "Max tool steps in agent mode")
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
	res, err := exec.Run(ctx, ports.ExecSpec{Command: aiCommand})
	if err != nil {
		if res.ExitCode > 0 {
			return fmt.Errorf("command failed with exit code %d", res.ExitCode)
		}
		return err
	}

	fmt.Println("\n✅ Command executed successfully.")
	return nil
}
