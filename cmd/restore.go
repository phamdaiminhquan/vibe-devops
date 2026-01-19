package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/app/safety"
	"github.com/spf13/cobra"
)

var (
	restoreList    bool
	restoreCleanup bool
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore files from safety backups",
	Long: `Restore files that were backed up before dangerous commands.

Examples:
  vibe restore              # Interactive: select a backup to restore
  vibe restore --list       # List all available backups
  vibe restore --cleanup    # Remove backups older than 7 days`,
	Run: func(cmd *cobra.Command, args []string) {
		if restoreCleanup {
			fmt.Println("Cleaning up old backups...")
			safety.CleanupOldBackups()
			fmt.Println("✅ Done!")
			return
		}

		if restoreList {
			listBackups()
			return
		}

		interactiveRestore()
	},
}

func listBackups() {
	backups, err := safety.GetRecentBackups(10)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	if len(backups) == 0 {
		fmt.Println("No safety backups found")
		fmt.Println("   Backups are created when running dangerous commands")
		return
	}

	fmt.Println("\nSafety Backups:")
	fmt.Println("═══════════════════════════════════════")
	for i, b := range backups {
		relTime := b.Timestamp.Format("2006-01-02 15:04:05")
		fmt.Printf("  %d. [%s]\n", i+1, relTime)
		fmt.Printf("     Command: %s\n", truncate(b.Command, 60))
		fmt.Printf("     Files: %d paths\n", len(b.BackupPaths))
	}
	fmt.Println()
}

func interactiveRestore() {
	backups, err := safety.GetRecentBackups(10)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	if len(backups) == 0 {
		fmt.Println("No safety backups found")
		return
	}

	fmt.Println("\nSafety Backups:")
	fmt.Println("═══════════════════════════════════════")
	for i, b := range backups {
		relTime := b.Timestamp.Format("2006-01-02 15:04:05")
		fmt.Printf("  %d. [%s]\n", i+1, relTime)
		fmt.Printf("     Command: %s\n", truncate(b.Command, 60))
		for orig := range b.BackupPaths {
			fmt.Printf("     • %s\n", orig)
		}
	}
	fmt.Print("\nSelect to restore (1-", len(backups), ", or 'q' to quit): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "q" || input == "" {
		fmt.Println("Cancelled.")
		return
	}

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(backups) {
		fmt.Println("❌ Invalid selection")
		return
	}

	selected := backups[idx-1]

	// Confirm restoration
	fmt.Println("\n⚠️  This will overwrite current files with backed up versions:")
	for orig := range selected.BackupPaths {
		fmt.Printf("   • %s\n", orig)
	}
	fmt.Print("\nContinue? (y/N) ")

	confirm, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		fmt.Println("Cancelled.")
		return
	}

	// Get backup path from timestamp
	homeDir, _ := os.UserHomeDir()
	backupPath := filepath.Join(homeDir, safety.BackupDir, selected.Timestamp.Format("2006-01-02_15-04-05"))

	fmt.Println("\nRestoring files...")
	if err := safety.RestoreBackup(backupPath); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	fmt.Println("✅ Done! Files restored.")
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

func init() {
	restoreCmd.Flags().BoolVar(&restoreList, "list", false, "List all available backups")
	restoreCmd.Flags().BoolVar(&restoreCleanup, "cleanup", false, "Remove backups older than 7 days")
	rootCmd.AddCommand(restoreCmd)
}
