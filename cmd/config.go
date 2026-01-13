package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/configstore/vibeyaml"
	appConfig "github.com/phamdaiminhquan/vibe-devops/internal/app/config"
	"github.com/phamdaiminhquan/vibe-devops/pkg/ai"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage vibe configuration",
	Long:  `Manage vibe configuration settings such as the AI provider and API keys.`,
}

var setProviderCmd = &cobra.Command{
	Use:   "provider [name]",
	Short: "Set the active AI provider",
	Long:  `Sets the active AI provider (e.g., "gemini").`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName := strings.ToLower(args[0])

		svc := appConfig.NewService(vibeyaml.New())
		_, err := svc.Load(".")
		if err != nil {
			return fmt.Errorf("failed to load config: %w. Please run 'vibe init' first", err)
		}

		// For now, we only support gemini, but this is extensible
		if providerName != "gemini" {
			return fmt.Errorf("unsupported provider: '%s'. Only 'gemini' is currently supported", providerName)
		}

		if _, err := svc.SetProvider(".", providerName); err != nil {
			return fmt.Errorf("failed to write updated config: %w", err)
		}

		fmt.Printf("✅ Active provider set to '%s'.\n", providerName)
		return nil
	},
}

var setApiKeyCmd = &cobra.Command{
	Use:   "api-key [key]",
	Short: "Set the API key for the active provider",
	Long:  `Sets the API key for the currently configured active AI provider.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey := args[0]

		svc := appConfig.NewService(vibeyaml.New())
		cfg, err := svc.Load(".")
		if err != nil {
			return fmt.Errorf("failed to load config: %w. Please run 'vibe init' first", err)
		}

		switch cfg.AI.Provider {
		case "gemini":
			if _, err := svc.SetGeminiAPIKey(".", apiKey); err != nil {
				return fmt.Errorf("failed to write updated config: %w", err)
			}

			fmt.Println("🔄 Validating API key and fetching available models...")
			models, err := ai.GetGeminiModels(apiKey)
			if err != nil {
				return fmt.Errorf("failed to validate API key: %w", err)
			}

			if len(models) == 0 {
				return fmt.Errorf("no models found for this API key")
			}

			fmt.Println("✅ API Key is valid.")
			fmt.Println("Select a model to use:")
			for i, m := range models {
				displayName := strings.TrimPrefix(m, "models/")
				fmt.Printf("[%d] %s\n", i+1, displayName)
			}

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter the number of the model: ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			index, err := strconv.Atoi(input)
			if err != nil || index < 1 || index > len(models) {
				return fmt.Errorf("invalid selection")
			}

			selectedModel := models[index-1]
			model := strings.TrimPrefix(selectedModel, "models/")
			if _, err := svc.SetGeminiModel(".", model); err != nil {
				return fmt.Errorf("failed to write updated config: %w", err)
			}
			fmt.Printf("Selected model: %s\n", model)

		default:
			return fmt.Errorf("no active provider set or provider '%s' is not supported for API key configuration", cfg.AI.Provider)
		}

		fmt.Printf("✅ API key for provider '%s' has been set.\n", cfg.AI.Provider)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(setProviderCmd)
	configCmd.AddCommand(setApiKeyCmd)
}
