package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/engines"
	"github.com/FacileStudio/sonar/internal/search"
)

var (
	flagJSON          bool
	flagCount         int
	flagEngine        string
	flagSearchTimeout time.Duration
)

var searchCmd = &cobra.Command{
	Use:   "search [query...]",
	Short: "Search the web across all enabled engines",
	Args:  cobra.ArbitraryArgs,
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().BoolVar(&flagJSON, "json", false, "emit results as JSON")
	searchCmd.Flags().IntVarP(&flagCount, "count", "n", 0, "max results to return (default: config count)")
	searchCmd.Flags().StringVar(&flagEngine, "engines", "", "comma-separated engines to run (default: all enabled)")
	searchCmd.Flags().DurationVarP(&flagSearchTimeout, "timeout", "t", 0, "timeout for search (default: config timeout)")
}

func parseEngines(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var out []string
	for name := range strings.SplitSeq(strings.TrimSpace(s), ",") {
		if trimmed := strings.TrimSpace(name); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		return fmt.Errorf("a query is required")
	}
	cfg, err := config.Load(configPath())
	if err != nil {
		return err
	}
	count := cfg.DefaultCount
	if flagCount > 0 {
		count = flagCount
	}
	if flagSearchTimeout > 0 {
		cfg.TimeoutSeconds = int(flagSearchTimeout.Seconds())
	}
	include := parseEngines(flagEngine)
	results, err := search.Query(cmd.Context(), cfg, query, count, include)
	if err != nil {
		return err
	}
	if flagJSON {
		return writeJSON(results)
	}
	writeText(results, query)
	return nil
}

func writeJSON(results []engines.Result) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

// palette holds the styles for one result list, chosen against the terminal
// background so colors read on a light or dark terminal alike.
type palette struct {
	rank    lipgloss.Style
	title   lipgloss.Style
	snippet lipgloss.Style
	url     lipgloss.Style
	engine  lipgloss.Style
}

// makePalette builds styles with adaptive colors: ANSI-safe names that resolve
// differently for a light and a dark background. Styles render desc messages
// plainly and strip ANSI when stdout is not a terminal.
func makePalette() palette {
	lightDark := lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
	return palette{
		rank:    lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("#3fa45c"), lipgloss.Color("#9bcf6a"))).Bold(true),
		title:   lipgloss.NewStyle().Bold(true),
		snippet: lipgloss.NewStyle().Faint(true),
		url:     lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("#5b21b6"), lipgloss.Color("#c39bfa"))),
		engine:  lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("#64748b"), lipgloss.Color("#7190a5"))),
	}
}

func writeText(results []engines.Result, query string) {
	p := makePalette()
	lipgloss.Println(fmt.Sprintf("sonar: %d results for %s", len(results), query))
	for i, r := range results {
		lipgloss.Println(p.rank.Render(fmt.Sprintf("%d.", i+1)) + " " + p.title.Render(r.Title))
		if r.Snippet != "" {
			lipgloss.Println("   " + p.snippet.Render(r.Snippet))
		}
		lipgloss.Println("   " + p.url.Render(r.URL) + "   " + p.engine.Render(r.Engine))
	}
}

func configPath() string { return config.Path() }
