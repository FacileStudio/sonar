package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/engines"
	"github.com/FacileStudio/sonar/internal/merge"
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
	if flagCount > 0 {
		cfg.DefaultCount = flagCount
	}
	disp, err := search.NewDispatcher(cfg)
	if err != nil {
		return fmt.Errorf("engine setup: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*2*time.Second)
	defer cancel()
	results, err := disp.Search(ctx, query, cfg.DefaultCount)
	if err != nil {
		return err
	}
	results = merge.Rank(results, cfg.DefaultCount)
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

var (
	rankStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	snippetStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	urlStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	engineStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Faint(true)
)

func writeText(results []engines.Result, query string) {
	fmt.Printf("sonar: %d results for %q\n", len(results), query)
	for i, r := range results {
		fmt.Printf("%s %s\n", rankStyle.Render(fmt.Sprintf("%d.", i+1)), titleStyle.Render(r.Title))
		if r.Snippet != "" {
			fmt.Printf("   %s\n", snippetStyle.Render(r.Snippet))
		}
		fmt.Printf("   %s   %s\n", urlStyle.Render(r.URL), engineStyle.Render(r.Engine))
	}
}

func configPath() string {
	if p := os.Getenv("SONAR_CONFIG"); p != "" {
		return p
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "sonar", "config.yml")
}
