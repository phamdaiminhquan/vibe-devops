package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update vibe to the latest version",
	Long:  `Updates vibe to the latest version available on GitHub Releases by running the installation script.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🚀 Starting update process...")

		installScript := "https://raw.githubusercontent.com/phamdaiminhquan/vibe-devops/main/install.sh"
		// Use curl to fetch the script and pipe it to sh
		cmdStr := fmt.Sprintf("curl -sSL %s | sh", installScript)

		// Execute the installation script
		// We connect Stdin, Stdout, and Stderr to allow interaction (e.g., sudo password) and visibility.
		updateProcess := exec.Command("sh", "-c", cmdStr)
		updateProcess.Stdin = os.Stdin
		updateProcess.Stdout = os.Stdout
		updateProcess.Stderr = os.Stderr

		if err := updateProcess.Run(); err != nil {
			return fmt.Errorf("failed to update vibe: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}