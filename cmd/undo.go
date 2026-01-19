package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	vibegit "github.com/phamdaiminhquan/vibe-devops/internal/app/git"
	"github.com/spf13/cobra"
)

var (
	undoLast bool
	undoList bool
)

var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Undo recent AI changes",
	Long: `Restore your workspace to a previous checkpoint created before AI sessions.

Examples:
  vibe undo              # Interactive: select a checkpoint to restore
  vibe undo --last       # Restore to the last checkpoint
  vibe undo --list       # List all available checkpoints`,
	Run: func(cmd *cobra.Command, args []string) {
		workDir, _ := os.Getwd()

		// Check if git repo
		if !vibegit.IsGitRepo(workDir) {
			fmt.Println("Not a git repository")
			return
		}

		if undoList {
			listCheckpoints(workDir)
			return
		}

		if undoLast {
			undoLastCheckpoint(workDir)
			return
		}

		// Interactive mode
		interactiveUndo(workDir)
	},
}

func listCheckpoints(workDir string) {
	checkpoints, err := vibegit.GetRecentCheckpoints(workDir, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(checkpoints) == 0 {
		fmt.Println("No vibe checkpoints found")
		fmt.Println("   Checkpoints are created automatically before AI sessions")
		return
	}

	fmt.Println("\nRecent Vibe Checkpoints:")
	fmt.Println("═══════════════════════════════════════")
	for i, cp := range checkpoints {
		fmt.Printf("  %d. [%s] %s - %s\n", i+1, cp.Hash, cp.Relative, cp.Message)
	}
	fmt.Println()
}

func undoLastCheckpoint(workDir string) {
	// Check for uncommitted changes
	if vibegit.HasUncommittedChanges(workDir) {
		fmt.Print("⚠️  You have uncommitted changes. Continue anyway? (y/N) ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(input)) != "y" {
			fmt.Println("Cancelled.")
			return
		}
	}

	fmt.Println("Restoring to last checkpoint...")
	if err := vibegit.UndoLastCheckpoint(workDir); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Done! Workspace restored to before the last AI session.")
}

func interactiveUndo(workDir string) {
	checkpoints, err := vibegit.GetRecentCheckpoints(workDir, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(checkpoints) == 0 {
		fmt.Println("No vibe checkpoints found")
		fmt.Println("   Checkpoints are created automatically before AI sessions")
		return
	}

	fmt.Println("\nRecent Vibe Checkpoints:")
	fmt.Println("═══════════════════════════════════════")
	for i, cp := range checkpoints {
		fmt.Printf("  %d. [%s] %s - %s\n", i+1, cp.Hash, cp.Relative, cp.Message)
	}
	fmt.Print("\nSelect to restore (1-", len(checkpoints), ", or 'q' to quit): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "q" || input == "" {
		fmt.Println("Cancelled.")
		return
	}

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(checkpoints) {
		fmt.Println("Invalid selection")
		return
	}

	selected := checkpoints[idx-1]

	// Check for uncommitted changes
	if vibegit.HasUncommittedChanges(workDir) {
		fmt.Print("You have uncommitted changes. Continue anyway? (y/N) ")
		confirm, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
			fmt.Println("Cancelled.")
			return
		}
	}

	fmt.Printf("Restoring to checkpoint %s...\n", selected.Hash)
	if err := vibegit.RestoreCheckpoint(workDir, selected.Hash); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Done! Workspace restored.")
}

func init() {
	undoCmd.Flags().BoolVar(&undoLast, "last", false, "Undo the last AI session")
	undoCmd.Flags().BoolVar(&undoList, "list", false, "List all available checkpoints")
	rootCmd.AddCommand(undoCmd)
}
