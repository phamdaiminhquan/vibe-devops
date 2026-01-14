package cmd

import (
	"context"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/app/bootstrap"
	"github.com/phamdaiminhquan/vibe-devops/internal/app/command"
	appSession "github.com/phamdaiminhquan/vibe-devops/internal/app/session"
	"github.com/spf13/cobra"
)

var runAgentMode bool
var runAgentMaxSteps int
var runSelfHeal bool
var runSelfHealMaxAttempts int

var runSessionName string
var runSessionScope string
var runResumeSession bool
var runNoSession bool
var runContextBudget int
var runContextRecentLines int

var runCmd = &cobra.Command{
	Use:   "run [natural language request]",
	Short: "Execute a command based on a natural language request",
	Long: `Takes a natural language request, uses an AI provider to translate it into a shell command,
and executes it after user confirmation.`,
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE:         runCommand,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&runAgentMode, "agent", true, "Enable agent mode (default: true). Use --agent=false for simple single-shot mode")
	runCmd.Flags().IntVar(&runAgentMaxSteps, "agent-max-steps", 10, "Max tool steps in agent mode")
	runCmd.Flags().BoolVar(&runSelfHeal, "self-heal", true, "In agent mode, keep iterating after execution by reading command output and proposing next steps until an answer is reached (default: true)")
	runCmd.Flags().IntVar(&runSelfHealMaxAttempts, "self-heal-max-attempts", 3, "Max execution/repair iterations in self-heal loop (agent mode only)")

	runCmd.Flags().StringVar(&runSessionName, "session", "default", "Session name for agent memory persistence")
	runCmd.Flags().StringVar(&runSessionScope, "session-scope", "both", "Session scope: none|project|global|both")
	runCmd.Flags().BoolVar(&runResumeSession, "resume", true, "Resume session memory (agent mode). When false, starts fresh but still writes updates.")
	runCmd.Flags().BoolVar(&runNoSession, "no-session", false, "Disable session persistence (agent mode)")
	runCmd.Flags().IntVar(&runContextBudget, "context-budget", 8000, "Approx char budget for session context tail")
	runCmd.Flags().IntVar(&runContextRecentLines, "context-recent-lines", 40, "Max recent transcript lines to keep in session memory")
}

func runCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// 1. Bootstrap Application
	appCtx, err := bootstrap.Initialize(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = appCtx.Provider.Close() }()

	// 2. Setup Session Service
	sessionCfg := bootstrap.SessionConfig{
		Name:      runSessionName,
		Scope:     runSessionScope,
		Resume:    runResumeSession,
		NoSession: runNoSession,
		Budget: appSession.Budget{
			MaxRecentLines: runContextRecentLines,
			MaxRecentChars: runContextBudget,
		},
	}
	sessionSvc := bootstrap.InitializeSessionService(appCtx.Provider, sessionCfg)

	// 3. Setup Command Handler
	flags := command.RunFlags{
		AgentMode:           runAgentMode,
		AgentMaxSteps:       runAgentMaxSteps,
		SelfHeal:            runSelfHeal,
		SelfHealMaxAttempts: runSelfHealMaxAttempts,
	}
	handler := command.NewRunHandler(appCtx, sessionSvc, flags)

	// 4. Delegate to Handler
	userRequest := strings.Join(args, " ")
	return handler.Handle(ctx, userRequest)
}
