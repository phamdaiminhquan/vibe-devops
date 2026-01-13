package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/configstore/vibeyaml"
	appConfig "github.com/phamdaiminhquan/vibe-devops/internal/app/config"
	"github.com/phamdaiminhquan/vibe-devops/pkg/ai"
	"github.com/phamdaiminhquan/vibe-devops/pkg/config"
	"github.com/spf13/cobra"
)

var modelListOnly bool

var modelCmd = &cobra.Command{
	Use:   "model [model]",
	Short: "Select or change the AI model",
	Long:  "Lists available models for the active provider and updates the configured model in .vibe.yaml.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := appConfig.NewService(vibeyaml.New())
		cfg, err := svc.Load(".")
		if err != nil {
			return fmt.Errorf("failed to load config: %w. Please run 'vibe init' first", err)
		}

		switch strings.ToLower(strings.TrimSpace(cfg.AI.Provider)) {
		case "gemini":
			apiKey := strings.TrimSpace(cfg.AI.Gemini.APIKey)
			if apiKey == "" || apiKey == config.DefaultAPIKeyPlaceholder {
				return fmt.Errorf("Gemini API key is not configured. Run 'vibe config api-key " + "<your_api_key>" + "' first")
			}

			if len(args) == 1 {
				model := strings.TrimSpace(args[0])
				model = strings.TrimPrefix(model, "models/")
				if _, err := svc.SetGeminiModel(".", model); err != nil {
					return fmt.Errorf("failed to write updated config: %w", err)
				}
				fmt.Printf("✅ Model set to '%s'.\n", model)
				return nil
			}

			models, err := ai.GetGeminiModels(apiKey)
			if err != nil {
				return fmt.Errorf("failed to fetch models: %w", err)
			}
			if len(models) == 0 {
				return fmt.Errorf("no models found for this API key")
			}

			if modelListOnly {
				for _, m := range models {
					fmt.Println(strings.TrimPrefix(m, "models/"))
				}
				return nil
			}

			fmt.Println("Select a model to use:")
			for i, m := range models {
				displayName := strings.TrimPrefix(m, "models/")
				fmt.Printf("[%d] %s\n", i+1, displayName)
			}

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter the number of the model: ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			idx := -1
			_, scanErr := fmt.Sscanf(input, "%d", &idx)
			if scanErr != nil || idx < 1 || idx > len(models) {
				return fmt.Errorf("invalid selection")
			}

			selectedModel := strings.TrimPrefix(models[idx-1], "models/")
			if _, err := svc.SetGeminiModel(".", selectedModel); err != nil {
				return fmt.Errorf("failed to write updated config: %w", err)
			}
			fmt.Printf("✅ Model set to '%s'.\n", selectedModel)
			return nil

		default:
			return fmt.Errorf("unsupported provider '%s'", cfg.AI.Provider)
		}
	},
}

func init() {
	rootCmd.AddCommand(modelCmd)
	modelCmd.Flags().BoolVar(&modelListOnly, "list", false, "List available models (does not change config)")
}
