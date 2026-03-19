package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/phamdaiminhquan/vibe-devops/internal/app/agent"
	"github.com/phamdaiminhquan/vibe-devops/internal/app/bootstrap"
	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/tools/fs"
	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/tools/system"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug [test query]",
	Short: "Run debug diagnostics for troubleshooting",
	Long: `Runs a complete debug session to diagnose issues.
Tests config, API, streaming, and agent mode step by step.

Usage:
  vibe debug                           # Run with default test
  vibe debug "list files"              # Test with custom query
  vibe debug --no-stream "list files"  # Test without streaming`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDebug,
}

var debugNoStream bool

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.Flags().BoolVar(&debugNoStream, "no-stream", false, "Disable streaming (use non-streaming Generate)")
}

func runDebug(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🔧 VIBE DEBUG DIAGNOSTICS")
	fmt.Println(strings.Repeat("=", 60))

	// ── Step 1: Configuration ─────────────────────────
	fmt.Println("\n📋 Step 1: Configuration Check")
	fmt.Println(strings.Repeat("-", 40))

	appCtx, err := bootstrap.Initialize(ctx)
	if err != nil {
		fmt.Printf("❌ Bootstrap failed: %v\n", err)
		return err
	}
	defer func() { _ = appCtx.Provider.Close() }()

	fmt.Printf("✅ AI Provider: %s\n", appCtx.Provider.Name())

	if err := appCtx.Provider.IsConfigured(ctx); err != nil {
		fmt.Printf("❌ Provider not configured: %v\n", err)
		fmt.Println("💡 Fix: Run 'vibe config' to set up your API key")
		return err
	}
	fmt.Println("✅ Provider is configured")

	// ── Step 2: API Connectivity ──────────────────────
	fmt.Println("\n🌐 Step 2: API Connectivity Test (non-streaming)")
	fmt.Println(strings.Repeat("-", 40))

	apiCtx, apiCancel := context.WithTimeout(ctx, 30*time.Second)
	defer apiCancel()

	start := time.Now()
	resp, err := appCtx.Provider.Generate(apiCtx, ports.GenerateRequest{
		Prompt: "Reply with exactly: VIBE_OK",
	})
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("❌ API failed (%s): %v\n", elapsed, err)
		return err
	}
	fmt.Printf("✅ API responded in %s\n", elapsed)
	fmt.Printf("   Response: %q\n", strings.TrimSpace(resp.Text))

	// ── Step 3: Stream Test ───────────────────────────
	fmt.Println("\n📡 Step 3: Stream Test")
	fmt.Println(strings.Repeat("-", 40))

	if debugNoStream {
		fmt.Println("⏭️  Skipped (--no-stream flag set)")
	} else {
		streamCtx, streamCancel := context.WithTimeout(ctx, 30*time.Second)
		defer streamCancel()

		start = time.Now()
		ch, err := appCtx.Provider.StreamGenerate(streamCtx, ports.GenerateRequest{
			Prompt: "Reply with exactly: STREAM_OK",
		})
		if err != nil {
			fmt.Printf("❌ StreamGenerate failed: %v\n", err)
		} else {
			var streamBuf strings.Builder
			chunkCount := 0
			for chunk := range ch {
				if chunk.Error != nil {
					fmt.Printf("❌ Stream chunk error: %v\n", chunk.Error)
					fmt.Printf("   Buffered so far (%d chunks): %q\n", chunkCount, streamBuf.String())
					break
				}
				streamBuf.WriteString(chunk.Content)
				chunkCount++
			}
			elapsed = time.Since(start)
			fmt.Printf("✅ Stream completed in %s (%d chunks)\n", elapsed, chunkCount)
			fmt.Printf("   Response: %q\n", strings.TrimSpace(streamBuf.String()))
		}
	}

	// ── Step 4: Agent Test ────────────────────────────
	fmt.Println("\n🤖 Step 4: Agent Test")
	fmt.Println(strings.Repeat("-", 40))

	testQuery := "list files in current directory"
	if len(args) > 0 {
		testQuery = args[0]
	}
	fmt.Printf("Query: %q\n", testQuery)

	if debugNoStream {
		fmt.Println("Mode: non-streaming (--no-stream)")
	} else {
		fmt.Println("Mode: streaming")
	}

	tools := []ports.Tool{
		fs.NewListDirTool("."),
		fs.NewReadFileTool("."),
		fs.NewGrepTool("."),
		system.NewSafeShellTool(),
	}

	ag := agent.NewService(appCtx.Provider, tools, appCtx.Logger, 5)

	agentCtx, agentCancel := context.WithTimeout(ctx, 60*time.Second)
	defer agentCancel()

	var tokenBuf strings.Builder

	var onToken func(string)
	if !debugNoStream {
		onToken = func(token string) {
			tokenBuf.WriteString(token)
			fmt.Print(token) // Show live to user
		}
	}

	onProgress := func(step agent.StepInfo) {
		fmt.Printf("\n   [Step %d] %s: %s\n", step.Step, step.Type, step.Message)
	}

	start = time.Now()
	agentResp, err := ag.SuggestCommand(agentCtx, agent.SuggestRequest{
		UserRequest: testQuery,
		GOOS:        runtime.GOOS,
		OnProgress:  onProgress,
		OnToken:     onToken,
	})
	elapsed = time.Since(start)

	fmt.Println() // newline after stream

	if err != nil {
		fmt.Printf("❌ Agent failed (%s): %v\n", elapsed, err)
		return err
	}

	fmt.Printf("✅ Agent completed in %s (%d steps)\n", elapsed, agentResp.StepsUsed)
	fmt.Printf("   Command:     %q\n", agentResp.Command)
	fmt.Printf("   Explanation: %q\n", agentResp.Explanation)

	if tokenBuf.Len() > 0 {
		fmt.Printf("   Streamed tokens: %d chars\n", tokenBuf.Len())
	} else if !debugNoStream {
		fmt.Println("   ⚠️  No tokens were streamed (onToken was never called)")
	}

	// ── Summary ───────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ ALL DIAGNOSTICS PASSED")
	fmt.Println(strings.Repeat("=", 60))

	// Write log file
	logFile := fmt.Sprintf("debug_%s.log", time.Now().Format("20060102_150405"))
	summary := fmt.Sprintf(
		"Provider: %s\nAPI: OK (%s)\nStream: %s\nAgent: %d steps in %s\nCommand: %s\nExplanation: %s\nOS: %s\nTime: %s\n",
		appCtx.Provider.Name(), elapsed,
		func() string {
			if debugNoStream {
				return "skipped"
			}
			return "OK"
		}(),
		agentResp.StepsUsed, elapsed,
		agentResp.Command, agentResp.Explanation,
		runtime.GOOS, time.Now().Format(time.RFC3339),
	)
	if err := os.WriteFile(logFile, []byte(summary), 0644); err == nil {
		fmt.Printf("📝 Summary saved to: %s\n", logFile)
	}

	return nil
}
