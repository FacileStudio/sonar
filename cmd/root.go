package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "sonar",
	Short: "Multi-engine web search with block-resistant fallback",
	Long: "sonar queries several search backends (keyed APIs, your SearXNG " +
		"instance, and scrape fallbacks) and merges their results into one " +
		"ranked list. Engines are dispatched in priority order with politeness " +
		"throttling and per-engine circuit breakers so a block-prone IP is not " +
		"hammered. Use --json for machine-readable output.",
	SilenceUsage:      true,
	Version:           version,
	Args:              cobra.ArbitraryArgs,
	DisableAutoGenTag: true,
}

// Execute runs the sonar command tree and exits non-zero on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "sonar:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(mcpCmd)
}
