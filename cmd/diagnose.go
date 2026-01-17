package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/app/bootstrap"
	"github.com/phamdaiminhquan/vibe-devops/internal/app/diagnose"
	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
	"github.com/spf13/cobra"
)

var (
	diagnoseWithAI bool
)

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Diagnose system health",
	Long:  `Run comprehensive system diagnostics: disk, RAM, Docker, network, services.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		fmt.Println("🔍 Collecting system information...")

		svc := diagnose.NewService()
		result, err := svc.Run(ctx)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		fmt.Print(diagnose.FormatReport(result))

		// AI Analysis if requested
		if diagnoseWithAI && (len(result.Warnings) > 0 || len(result.Errors) > 0) {
			analyzeWithAI(ctx, result)
		}
	},
}

func analyzeWithAI(ctx context.Context, result *diagnose.DiagnoseResult) {
	fmt.Println("\n🤖 Analyzing with AI...")

	// Initialize provider
	appCtx, err := bootstrap.Initialize(ctx)
	if err != nil {
		fmt.Printf("⚠️ Cannot connect to AI: %v\n", err)
		return
	}
	defer appCtx.Provider.Close()

	// Build prompt with diagnostics data
	var sb strings.Builder
	sb.WriteString("Analyze the following system diagnostics and suggest fixes:\n\n")

	if len(result.Errors) > 0 {
		sb.WriteString("CRITICAL ERRORS:\n")
		for _, issue := range result.Errors {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", issue.Category, issue.Description))
		}
	}

	if len(result.Warnings) > 0 {
		sb.WriteString("\nWARNINGS:\n")
		for _, issue := range result.Warnings {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", issue.Category, issue.Description))
		}
	}

	sb.WriteString("\nPlease analyze root causes and suggest specific remediation steps.")

	// Generate AI response
	resp, err := appCtx.Provider.Generate(ctx, ports.GenerateRequest{
		Prompt: sb.String(),
	})
	if err != nil {
		fmt.Printf("⚠️  AI Error: %v\n", err)
		return
	}

	fmt.Println("\nAI Analysis:")
	fmt.Println("─────────────────")
	fmt.Println(resp.Text)
}

func init() {
	diagnoseCmd.Flags().BoolVar(&diagnoseWithAI, "ai", false, "Analyze results with AI")
	rootCmd.AddCommand(diagnoseCmd)
}
