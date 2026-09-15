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
)

var fetchCmd = &cobra.Command{
	Use:   "fetch <URL> [--wait-until networkidle] [--selector css] [--timeout 30s]",
	Short: "Fetch a single web page as markdown",
	Long:  "Extract the HTML content from a single URL using a headless browser and print clean markdown to stdout.",
	Args:  cobra.ExactArgs(1),
	RunE:  runFetch,
}

func init() {
	fetchCmd.Flags().StringVar(&flagWaitUntil, "wait-until", "networkidle", "wait condition: load, domcontentloaded, networkidle, or none")
	fetchCmd.Flags().StringVarP(&flagSelector, "selector", "s", "", "CSS selector to target specific container element")
	fetchCmd.Flags().DurationVarP(&flagTimeout, "timeout", "t", 30*time.Second, "navigation and rendering timeout")
}

func runFetch(cmd *cobra.Command, args []string) error {
	url := strings.TrimSpace(args[0])
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	count := rodCount()
	rodEngine := &engines.Rod{
		Count:         count,
		WaitCondition: flagWaitUntil,
		Selector:      flagSelector,
		Timeout:       flagTimeout,
	}
	results, err := rodEngine.Search(cmd.Context(), url, 1)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		return fmt.Errorf("no results returned")
	}
	fmt.Println(results[0].Snippet)
	return nil
}

func rodCount() int {
	cfg, err := config.Load(configPath())
	if err == nil && cfg != nil {
		if def, ok := cfg.Engines["rod"]; ok && def.Count > 0 {
			return def.Count
		}
	}
	return 1
}
