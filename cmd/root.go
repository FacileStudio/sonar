package cmd

import (
	"context"
	"os"

	"charm.land/fang/v2"
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
	DisableAutoGenTag: true,
}

// Execute runs the sonar command tree through fang (styled help and errors,
// --version, man and completion commands) and exits non-zero on error. Fang
// already renders the styled error, so the wrapper only forwards the exit code.
func Execute() {
	if fang.Execute(context.Background(), rootCmd, fang.WithVersion(version)) != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(cacheCmd)
}
