package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vibe",
	Short: "An Open Source AI Agent for DevOps",
	Long: `vibe is an open-source AI terminal agent for automated VPS management 
and self-healing Docker deployments. It helps with VPS and Docker automation.`,
	Version: "dev",
}

func rewriteArgsForDefaultRun(argv []string) []string {
	if len(argv) <= 1 {
		return argv
	}

	first := argv[1]
	if first == "" {
		return argv
	}
	// If the first token is a flag, default to `run` unless it's help/version.
	if strings.HasPrefix(first, "-") {
		switch first {
		case "-h", "--help", "--version", "-v":
			return argv
		default:
			out := make([]string, 0, len(argv)+1)
			out = append(out, argv[0], "run")
			out = append(out, argv[1:]...)
			return out
		}
	}

	// If the first token matches a known subcommand (or alias), don't rewrite.
	for _, c := range rootCmd.Commands() {
		if c.Name() == first {
			return argv
		}
		for _, a := range c.Aliases {
			if a == first {
				return argv
			}
		}
	}

	// Otherwise treat it as a natural language request and default to `run`.
	out := make([]string, 0, len(argv)+1)
	out = append(out, argv[0], "run")
	out = append(out, argv[1:]...)
	return out
}

// SetVersionInfo sets version metadata (usually injected at build time).
func SetVersionInfo(version, commit, date string) {
	if version == "" {
		version = rootCmd.Version
	}

	var meta []string
	if commit != "" {
		meta = append(meta, "commit "+commit)
	}
	if date != "" {
		meta = append(meta, "built "+date)
	}

	if len(meta) > 0 {
		rootCmd.Version = fmt.Sprintf("%s (%s)", version, strings.Join(meta, ", "))
		return
	}

	rootCmd.Version = version
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	os.Args = rewriteArgsForDefaultRun(os.Args)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Add global flags here if needed
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vibe.yaml)")
}
