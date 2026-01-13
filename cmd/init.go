package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize vibe-devops in a directory",
	Long: `Scans a directory to initialize vibe-devops configuration.
This command will analyze the project structure and set up necessary configurations.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Default to current directory if no argument provided
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	// Convert to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absDir)
	}

	fmt.Printf("🔍 Scanning directory: %s\n", absDir)
	
	// Mock scanning process
	fmt.Println("📦 Analyzing project structure...")
	
	// Walk through directory (mock scan)
	fileCount := 0
	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileCount++
		}
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	fmt.Printf("✅ Found %d files\n", fileCount)
	fmt.Println("🎉 Initialization complete!")
	fmt.Println("\nNext steps:")
	fmt.Println("  • Configure your AI provider in .vibe-devops.yaml")
	fmt.Println("  • Run 'vibe-devops run <command>' to execute commands with AI assistance")
	
	return nil
}
