package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/engines"
)

var (
	flagWaitUntil string
	flagSelector  string
	flagTimeout   time.Duration
	flagBrowser   bool
)

var fetchCmd = &cobra.Command{
	Use:   "fetch <URL>... [--wait-until networkidle] [--selector css] [--timeout 30s] [--browser]",
	Short: "Fetch web pages as markdown",
	Long:  "Extract HTML content from one or more URLs and print clean markdown to stdout.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runFetch,
}

func init() {
	fetchCmd.Flags().StringVar(&flagWaitUntil, "wait-until", "networkidle", "wait condition: load, domcontentloaded, networkidle, or none")
	fetchCmd.Flags().StringVarP(&flagSelector, "selector", "s", "", "CSS selector to target specific container element")
	fetchCmd.Flags().DurationVarP(&flagTimeout, "timeout", "t", 30*time.Second, "navigation and rendering timeout")
	fetchCmd.Flags().BoolVarP(&flagBrowser, "browser", "b", false, "force headless browser rendering")
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
	opts := engines.ExtractOptions{
		ParallelKey:   parallelKey(),
		WaitCondition: flagWaitUntil,
		Selector:      flagSelector,
		Timeout:       flagTimeout,
		ForceBrowser:  flagBrowser,
	}
	results, err := engines.ExtractPages(cmd.Context(), urls, opts)
	if err != nil {
		return err
	}
	for _, r := range results {
		if len(results) > 1 {
			fmt.Printf("# %s\n\n%s\n\n", r.URL, r.Snippet)
		} else {
			fmt.Println(r.Snippet)
		}
	}
	return nil
}

func parallelKey() string {
	cfg, err := config.Load(configPath())
	if err == nil && cfg != nil {
		if def, ok := cfg.Engines["parallel"]; ok && def.APIKey != "" {
			return def.APIKey
		}
	}
	return ""
}
