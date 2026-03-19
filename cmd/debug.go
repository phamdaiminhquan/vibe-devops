package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/phamdaiminhquan/vibe-devops/internal/app/bootstrap"
	"github.com/phamdaiminhquan/vibe-devops/internal/app/command"
	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug [test query]",
	Short: "Run debug diagnostics for troubleshooting",
	Long: `Runs a complete debug session with logging enabled.
Helps diagnose issues with API configuration, network, and response formatting.

Usage:
  vibe debug                           # Run basic connectivity test
  vibe debug "list files"              # Test with custom query`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDebug,
}

var debugFlags struct {
	OutputFile string
	Verbose    bool
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.Flags().StringVar(&debugFlags.OutputFile, "output", "", "Save debug log to file (default: debug_<timestamp>.log)")
	debugCmd.Flags().BoolVar(&debugFlags.Verbose, "verbose", true, "Show all logs (default: true)")
}

func runDebug(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🔧 VIBE DEBUG DIAGNOSTICS")
	fmt.Println(strings.Repeat("=", 60))

	// Step 1: Check configuration
	fmt.Println("\n📋 Step 1: Configuration Check")
	fmt.Println(strings.Repeat("-", 40))

	appCtx, err := bootstrap.Initialize(ctx)
	if err != nil {
		fmt.Printf("❌ Bootstrap failed: %v\n", err)
		return err
	}

	fmt.Printf("✅ AI Provider: %s\n", appCtx.Provider.Name())
	fmt.Printf("✅ Logger initialized\n")

	// Step 2: Test provider configuration
	fmt.Println("\n🔐 Step 2: Provider Configuration")
	fmt.Println(strings.Repeat("-", 40))

	if err := appCtx.Provider.IsConfigured(ctx); err != nil {
		fmt.Printf("❌ Provider not configured: %v\n", err)
		fmt.Println("\n💡 Fix: Run 'vibe config' to set up your API key")
		return err
	}
	fmt.Printf("✅ Provider is configured\n")

	// Step 3: Test API connectivity
	fmt.Println("\n🌐 Step 3: API Connectivity Test")
	fmt.Println(strings.Repeat("-", 40))

	testPrompt := "Say 'Hello from vibe debug' in one sentence."
	fmt.Printf("Testing API with prompt: %q\n", testPrompt)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := appCtx.Provider.Generate(ctx, ports.GenerateRequest{Prompt: testPrompt})

	if err != nil {
		fmt.Printf("❌ API test failed: %v\n", err)
		fmt.Println("\n💡 Possible causes:")
		fmt.Println("  - Invalid API key")
		fmt.Println("  - Network/firewall blocking API")
		fmt.Println("  - API rate limit exceeded")
		return err
	}

	fmt.Printf("✅ API responded successfully\n")
	if len(resp.Text) > 100 {
		fmt.Printf("   Response preview: %s...\n", resp.Text[:100])
	} else {
		fmt.Printf("   Response: %s\n", resp.Text)
	}

	// Step 4: Test agent mode with custom query
	fmt.Println("\n🤖 Step 4: Agent Test")
	fmt.Println(strings.Repeat("-", 40))

	testQuery := "list files"
	if len(args) > 0 {
		testQuery = args[0]
	}

	fmt.Printf("Running agent with query: %q\n", testQuery)
	fmt.Printf("(Streaming enabled with debug logging)\n\n")

	// Capture output to file
	logFile := debugFlags.OutputFile
	if logFile == "" {
		timestamp := time.Now().Format("20060102_150405")
		logFile = filepath.Join(".", fmt.Sprintf("debug_%s.log", timestamp))
	}

	logWriter, err := os.Create(logFile)
	if err != nil {
		fmt.Printf("⚠️  Could not create log file: %v\n", err)
		logWriter = os.Stdout
	} else {
		defer logWriter.Close()
		fmt.Printf("📝 Logging to: %s\n\n", logFile)
	}

	// Redirect output
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	if logWriter != os.Stdout {
		os.Stdout = logWriter
		os.Stderr = logWriter
	}

	// Run agent with flags
	runHandler := command.NewRunHandler(appCtx, nil, command.RunFlags{
		AgentMode:           true,
		AgentMaxSteps:       5,
		SelfHeal:            false,
		SelfHealMaxAttempts: 0,
	})

	err = runHandler.Handle(ctx, testQuery)

	// Restore output
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	if err != nil {
		fmt.Printf("❌ Agent test failed: %v\n", err)
		fmt.Printf("📝 Full logs saved to: %s\n", logFile)
		return err
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ DEBUG DIAGNOSTICS COMPLETE")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📝 Full logs saved to: %s\n", logFile)
	fmt.Println("\n💡 Next steps:")
	fmt.Println("  1. Check the log file for any errors")
	fmt.Println("  2. Look for 'explanation field not found' if response is missing")
	fmt.Println("  3. Check API response format in the logs")
	fmt.Println("  4. Share the log file if you need help debugging\n")

	return nil
}
