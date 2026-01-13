package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [command]",
	Short: "Execute a shell command",
	Long: `Execute a shell command with AI-powered assistance.
This command runs the specified shell command and can be enhanced with AI capabilities.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCommand,
}

func init() {
	rootCmd.AddCommand(runCmd)
	// Add flags for run command
	runCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
}

func runCommand(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	
	// Join all arguments to form the complete command
	commandStr := strings.Join(args, " ")
	
	if verbose {
		fmt.Printf("🚀 Executing command: %s\n", commandStr)
	}

	// Create the command using sh -c to properly handle shell commands
	// Note: This intentionally executes user-provided commands as-is for DevOps automation.
	// Users are responsible for the commands they execute. This is the expected behavior
	// for a DevOps CLI tool that needs to run arbitrary shell commands.
	shellCmd := exec.Command("sh", "-c", commandStr)
	
	// Set up command to use standard input/output/error
	shellCmd.Stdin = os.Stdin
	shellCmd.Stdout = os.Stdout
	shellCmd.Stderr = os.Stderr

	// Execute the command
	err := shellCmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Command executed but returned non-zero exit code
			return fmt.Errorf("command failed with exit code %d", exitErr.ExitCode())
		}
		return fmt.Errorf("failed to execute command: %w", err)
	}

	if verbose {
		fmt.Println("✅ Command executed successfully")
	}

	return nil
}
