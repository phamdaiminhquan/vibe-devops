package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage vibe configuration",
	Long:  `Set or get configuration values for vibe.`, 
}

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value (e.g., gemini.apikey)",
	Long:  `Set a configuration value in the .vibe.yaml file. Currently supported keys: "gemini.apikey".`, 
	Args:  cobra.ExactArgs(2),
	RunE:  runSet,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(setCmd)
}

func runSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	dir, err := findVibeConfigDir()
	if err != nil {
		return err
	}
	
	configFile := filepath.Join(dir, config.ConfigFileName)

	cfg, err := config.Load(dir)
	if err != nil {
		return fmt.Errorf("failed to load config file at %s: %w. Please run 'vibe init .' first", dir, err)
	}

	switch strings.ToLower(key) {
	case "gemini.apikey":
		cfg.AI.Gemini.APIKey = value
	default:
		return fmt.Errorf("unsupported configuration key: %s. Supported keys: 'gemini.apikey'", key)
	}

	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	if err := os.WriteFile(configFile, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	fmt.Printf("✅ Successfully updated '%s' in %s\n", key, configFile)
	return nil
}

// findVibeConfigDir searches for the .vibe.yaml file in the current directory and parent directories.
func findVibeConfigDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get current directory: %w", err)
	}

	for {
		configFile := filepath.Join(dir, config.ConfigFileName)
		if _, err := os.Stat(configFile); err == nil {
			return dir, nil
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// Reached the root directory
			return "", fmt.Errorf("could not find a .vibe.yaml file in the current directory or any parent directory. Please run 'vibe init .'")
		}
		dir = parentDir
	}
}
