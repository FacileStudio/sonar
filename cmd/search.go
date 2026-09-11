package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/engines"
	"github.com/FacileStudio/sonar/internal/search"
)

var (
	flagJSON  bool
	flagCount int
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
	var results []engines.Result
	if flagJSON {
		results, err = search.Query(context.Background(), cfg, query, count)
		if err != nil {
			return err
		}
		return writeJSON(results)
	}
	results, err = searchWithSpinner(context.Background(), cfg, query, count)
	if err != nil {
		return err
	}
	writeText(results, query)
	return nil
}

// searchWithSpinner runs the search while a spinner animates on stderr, so
// stdout stays clean for JSON and the user sees the run is alive. Frames stop
// when the query answers, success or failure.
func searchWithSpinner(ctx context.Context, cfg *config.Config, query string, count int) ([]engines.Result, error) {
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		for i := 0; ; i = (i + 1) % len(frames) {
			select {
			case <-done:
				fmt.Fprint(os.Stderr, "\r\033[K")
				return
			case <-time.After(100 * time.Millisecond):
				fmt.Fprintf(os.Stderr, "\r\033[K%s searching…", frames[i])
			}
		}
	}()
	results, err := search.Query(ctx, cfg, query, count)
	close(done)
	wg.Wait()
	return results, err
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
