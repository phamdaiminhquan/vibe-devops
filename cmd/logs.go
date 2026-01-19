package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/adapters/context/logs"
	"github.com/phamdaiminhquan/vibe-devops/internal/app/bootstrap"
	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
	"github.com/spf13/cobra"
)

var (
	logsTail    int
	logsErrors  bool
	logsJSON    bool
	logsAnalyze bool
	logsNoColor bool
)

var logsCmd = &cobra.Command{
	Use:   "logs [file]",
	Short: "View and analyze log files",
	Long: `View log files with syntax highlighting, error detection, and AI analysis.

Examples:
  vibe logs /var/log/nginx/error.log          # View with highlighting
  vibe logs --tail 50 app.log                 # Last 50 lines
  vibe logs --errors app.log                  # Show only errors
  vibe logs --analyze app.log                 # AI analysis
  vibe logs --json app.json.log               # JSON pretty-print`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		// Resolve path
		if !filepath.IsAbs(filePath) {
			cwd, _ := os.Getwd()
			filePath = filepath.Join(cwd, filePath)
		}

		// Read file
		lines, err := readLogFile(filePath, logsTail)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		// Detect format
		logFormat := logs.FormatPlain
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				logFormat = logs.DetectFormat(line)
				break
			}
		}

		useColor := !logsNoColor

		// Display based on options
		if logsErrors {
			displayErrorsOnly(lines, useColor)
		} else if logsJSON && logFormat == logs.FormatJSON {
			displayJSONLogs(lines, useColor)
		} else {
			displayLogs(lines, logFormat, useColor)
		}

		// AI analysis if requested
		if logsAnalyze {
			analyzeLogsWithAI(cmd.Context(), lines, filePath)
		}
	},
}

func readLogFile(path string, n int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Return last N lines
	if n > 0 && len(lines) > n {
		return lines[len(lines)-n:], nil
	}
	return lines, nil
}

func displayLogs(lines []string, format logs.LogFormat, useColor bool) {
	// Analyze issues
	issues := logs.AnalyzeLines(lines)
	levelCounts := logs.CountByLevel(lines)

	// Header
	fmt.Println("\n📄 Log Analysis")
	fmt.Println("═══════════════════════════════════════")

	// Summary
	fmt.Println("\n📊 Summary:")
	for _, level := range []logs.LogLevel{logs.LevelFatal, logs.LevelError, logs.LevelWarn, logs.LevelInfo} {
		if count := levelCounts[level]; count > 0 {
			fmt.Printf("  • %s: %d\n", logs.LevelString(level), count)
		}
	}

	// Issues
	if len(issues) > 0 {
		fmt.Printf("\n⚠️ Issues detected: %d\n", len(issues))
		summary := logs.SummarizeIssues(issues)
		for cat, count := range summary {
			fmt.Printf("  • %s: %d\n", cat, count)
		}
	} else {
		fmt.Println("\n✅ No obvious errors detected")
	}

	// Log content
	fmt.Printf("\n📃 Content (%d lines):\n", len(lines))
	fmt.Println("───────────────────────────────────────")

	for i, line := range lines {
		lineNum := fmt.Sprintf("%4d │ ", i+1)

		switch format {
		case logs.FormatJSON:
			if entry, err := logs.ParseJSONLine(line); err == nil {
				fmt.Print(lineNum + logs.FormatParsedEntry(entry, useColor) + "\n")
			} else {
				fmt.Print(lineNum + line + "\n")
			}
		case logs.FormatLogfmt:
			entry := logs.ParseLogfmtLine(line)
			fmt.Print(lineNum + logs.FormatParsedEntry(entry, useColor) + "\n")
		default:
			if useColor {
				fmt.Print(lineNum + logs.HighlightLine(line) + "\n")
			} else {
				fmt.Print(lineNum + line + "\n")
			}
		}
	}
}

func displayErrorsOnly(lines []string, useColor bool) {
	fmt.Println("\n🔴 Errors and Warnings Only")
	fmt.Println("═══════════════════════════════════════")

	count := 0
	for i, line := range lines {
		level := logs.DetectLevel(line)
		if level == logs.LevelError || level == logs.LevelFatal || level == logs.LevelWarn {
			lineNum := fmt.Sprintf("%4d │ ", i+1)
			if useColor {
				fmt.Print(lineNum + logs.HighlightLine(line) + "\n")
			} else {
				fmt.Printf("%s[%s] %s\n", lineNum, logs.LevelString(level), line)
			}
			count++
		}
	}

	if count == 0 {
		fmt.Println("✅ No errors or warnings found")
	} else {
		fmt.Printf("\n📊 Total: %d issues\n", count)
	}
}

func displayJSONLogs(lines []string, useColor bool) {
	fmt.Println("\n📋 JSON Logs (pretty-print)")
	fmt.Println("═══════════════════════════════════════")

	for i, line := range lines {
		fmt.Printf("\n--- Entry %d ---\n", i+1)
		fmt.Println(logs.PrettyPrintJSON(line, useColor))
	}
}

func analyzeLogsWithAI(ctx context.Context, lines []string, filePath string) {
	fmt.Println("\n🤖 Analyzing with AI...")

	// Initialize provider
	appCtx, err := bootstrap.Initialize(ctx)
	if err != nil {
		fmt.Printf("⚠️  Cannot connect to AI: %v\n", err)
		return
	}
	defer appCtx.Provider.Close()

	// Build prompt
	var sb strings.Builder
	sb.WriteString("Analyze the following log file and provide insights:\n\n")
	sb.WriteString(fmt.Sprintf("File: %s\n", filepath.Base(filePath)))
	sb.WriteString(fmt.Sprintf("Lines: %d\n\n", len(lines)))

	// Include issues
	issues := logs.AnalyzeLines(lines)
	if len(issues) > 0 {
		sb.WriteString("Detected issues:\n")
		for _, issue := range issues {
			sb.WriteString(fmt.Sprintf("- Line %d: %s (%s)\n", issue.Line, issue.Description, issue.Category))
		}
		sb.WriteString("\n")
	}

	// Include sample of error lines (first 20)
	errorLines := 0
	sb.WriteString("Sample error lines:\n")
	for _, line := range lines {
		level := logs.DetectLevel(line)
		if level == logs.LevelError || level == logs.LevelFatal {
			sb.WriteString("  " + line + "\n")
			errorLines++
			if errorLines >= 20 {
				break
			}
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

	fmt.Println("\n📋 AI Analysis:")
	fmt.Println("─────────────────")
	fmt.Println(resp.Text)
}

func init() {
	logsCmd.Flags().IntVar(&logsTail, "tail", 100, "Number of lines from end of file")
	logsCmd.Flags().BoolVar(&logsErrors, "errors", false, "Show only errors and warnings")
	logsCmd.Flags().BoolVar(&logsJSON, "json", false, "Pretty-print JSON logs")
	logsCmd.Flags().BoolVar(&logsAnalyze, "analyze", false, "Analyze with AI")
	logsCmd.Flags().BoolVar(&logsNoColor, "no-color", false, "Disable color output")
	rootCmd.AddCommand(logsCmd)
}
