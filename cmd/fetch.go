package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/engines"
)

var flagWaitUntil string

var fetchCmd = &cobra.Command{
	Use:   "fetch <URL> [--wait-until networkidle]",
	Short: "Fetch a single web page as markdown",
	Long:  "Extract the HTML content from a single URL and print the markdown representation to stdout.",
	Args:  cobra.ExactArgs(1),
	RunE:  runFetch,
}

func init() {
	fetchCmd.Flags().StringVar(&flagWaitUntil, "wait-until", "networkidle", "wait condition: load, domcontentloaded, networkidle, or none")
}

func runFetch(cmd *cobra.Command, args []string) error {
	url := strings.TrimSpace(args[0])
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	cfg, _ := config.Load(configPath())
	count := 1
	if cfg != nil {
		if def, ok := cfg.Engines["rod"]; ok && def.Count > 0 {
			count = def.Count
		}
	}

	rodEngine := &engines.Rod{Count: count, WaitCondition: flagWaitUntil}
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
