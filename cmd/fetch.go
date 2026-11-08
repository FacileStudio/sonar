package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/engines"
	"github.com/FacileStudio/sonar/internal/netguard"
)

var (
	flagWaitUntil  string
	flagSelector   string
	flagTimeout    time.Duration
	flagBrowser    bool
	flagPublicOnly bool
	flagFetchJSON  bool
)

var fetchCmd = &cobra.Command{
	Use:   "fetch <URL>... [--wait-until networkidle] [--selector css] [--timeout 30s] [--browser] [--public-only]",
	Short: "Fetch web pages as markdown",
	Long: "Extract HTML content from one or more URLs and print clean markdown to stdout. " +
		"--public-only (or fetch.publicOnly in the config, or SONAR_FETCH_PUBLIC_ONLY=1) refuses " +
		"every request, redirect hop and subresource that would reach an internal address.",
	Args: cobra.MinimumNArgs(1),
	RunE: runFetch,
}

func init() {
	fetchCmd.Flags().StringVar(&flagWaitUntil, "wait-until", "networkidle", "wait condition: load, domcontentloaded, networkidle, or none")
	fetchCmd.Flags().StringVarP(&flagSelector, "selector", "s", "", "CSS selector to target specific container element")
	fetchCmd.Flags().DurationVarP(&flagTimeout, "timeout", "t", 30*time.Second, "navigation and rendering timeout")
	fetchCmd.Flags().BoolVarP(&flagBrowser, "browser", "b", false, "force headless browser rendering")
	fetchCmd.Flags().BoolVar(&flagPublicOnly, "public-only", false, "refuse any request to loopback, private, link-local or internal hosts")
	fetchCmd.Flags().BoolVar(&flagFetchJSON, "json", false, "emit one JSON object per URL, with an error code on failure")
}

func runFetch(cmd *cobra.Command, args []string) error {
	urls := make([]string, 0, len(args))
	for _, arg := range args {
		u := strings.TrimSpace(arg)
		if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
			return fmt.Errorf("URL must start with http:// or https://: %s", u)
		}
		urls = append(urls, u)
	}
	cfg, _ := config.Load(configPath())
	opts := engines.ExtractOptions{
		ParallelKey:   parallelKey(cfg),
		WaitCondition: flagWaitUntil,
		Selector:      flagSelector,
		Timeout:       flagTimeout,
		ForceBrowser:  flagBrowser,
	}
	if flagPublicOnly || cfg.FetchPublicOnly() {
		opts.Guard = &netguard.Guard{}
	}
	outcomes := engines.ExtractEach(cmd.Context(), urls, opts)
	if flagFetchJSON {
		return printFetchJSON(outcomes)
	}
	return printFetchMarkdown(outcomes)
}

func parallelKey(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	if def, ok := cfg.Engines["parallel"]; ok && def.APIKey != "" {
		return def.APIKey
	}
	return ""
}
