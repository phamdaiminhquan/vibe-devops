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
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Add global flags here if needed
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vibe.yaml)")
}
