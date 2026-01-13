package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/pkg/ai"
	"github.com/phamdaiminhquan/vibe-devops/pkg/config"
	"github.com/spf13/cobra"
)

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
}

func runCommand(cmd *cobra.Command, args []string) error {
	// 1. Load config
	cfg, err := config.Load(".")
	if err != nil {
		return fmt.Errorf("could not load configuration from .vibe.yaml. Please run 'vibe init' first. Error: %w", err)
	}

	// 2. Instantiate AI provider
	var provider ai.Provider
	switch cfg.AI.Provider {
	case "gemini":
		p, err := ai.NewGeminiProvider(cfg.AI.Gemini)
		if err != nil {
			return err
		}
		provider = p
	default:
		return fmt.Errorf("unsupported AI provider: %s", cfg.AI.Provider)
	}

	if !provider.IsConfigured() {
		return fmt.Errorf("AI provider '%s' is not configured. Please add your credentials to .vibe.yaml", provider.GetName())
	}

	// 3. Get user request and create prompt
	userRequest := strings.Join(args, " ")
	prompt := buildPrompt(userRequest)

	fmt.Println("🤖 Calling AI to generate command...")
	aiCommand, err := provider.GetCompletion(prompt)
	if err != nil {
		return fmt.Errorf("AI completion failed: %w", err)
	}
	
	sanitizedCommand := sanitizeAIResponse(aiCommand)
	if strings.HasPrefix(sanitizedCommand, "Error:") {
		return fmt.Errorf("AI returned an error: %s", sanitizedCommand)
	}

	// 4. Ask for user confirmation
	fmt.Printf("\n✨ Vibe suggests the following command:\n\n")
	fmt.Printf("  \033[1;36m%s\033[0m\n\n", sanitizedCommand) // Bold cyan
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
	shell, flag := getShell()
	execCmd := exec.Command(shell, flag, sanitizedCommand)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d", exitErr.ExitCode())
		}
		return fmt.Errorf("failed to execute command: %w", err)
	}

	fmt.Println("\n✅ Command executed successfully.")
	return nil
}

func buildPrompt(userRequest string) string {
	return fmt.Sprintf(
		`You are an expert AI assistant specializing in shell commands. Your task is to convert a user's request into a single, executable shell command for a %s environment.
- Only output the raw command.
- Do not include any explanation, markdown, backticks, or any text other than the command itself.
- If the request is ambiguous or unsafe, reply with "Error: Ambiguous or unsafe request."

User's request: "%s"
Shell command:`,
		runtime.GOOS,
		userRequest,
	)
}

func sanitizeAIResponse(response string) string {
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "`")
	response = strings.TrimSuffix(response, "`")
	response = strings.TrimPrefix(response, "shell")
	return strings.TrimSpace(response)
}

func getShell() (string, string) {
	if runtime.GOOS == "windows" {
		return "powershell", "-Command"
	}
	return "sh", "-c"
}
