package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/phamdaiminhquan/vibe-devops/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize vibe in a directory",
	Long:  `Creates a default .vibe.yaml configuration file in the specified directory.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absDir)
	}

	configFile := filepath.Join(absDir, config.ConfigFileName)
	if _, err := os.Stat(configFile); err == nil {
		fmt.Printf("✅ Configuration file already exists: %s\n", configFile)
		return nil
	}

	fmt.Printf("⚙️  Creating default configuration file: %s\n", configFile)

	defaultConfig := config.GetDefaultConfig()
	yamlData, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config to YAML: %w", err)
	}

	if err := os.WriteFile(configFile, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	fmt.Println("🎉 Initialization complete!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Run 'vibe config provider gemini' to select the provider.")
	fmt.Println("  2. Run 'vibe config api-key \"<your_api_key>\"' to set your API key and pick a model.")
	fmt.Println("  3. Run 'vibe run \"<your request>\"' to start using the agent.")

	return nil
}
