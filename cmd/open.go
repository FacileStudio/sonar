package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/engines"
	"github.com/FacileStudio/sonar/internal/search"
)

var openCmd = &cobra.Command{
	Use:   "open [n] [query...]",
	Short: "Search and open result N (default 1) in the browser",
	Args:  cobra.ArbitraryArgs,
	RunE:  runOpen,
}

func init() {}

func runOpen(cmd *cobra.Command, args []string) error {
	n, rest, err := parseIndex(args)
	if err != nil {
		return err
	}
	query := strings.TrimSpace(strings.Join(rest, " "))
	if query == "" {
		return fmt.Errorf("a query is required")
	}
	cfg, err := config.Load(configPath())
	if err != nil {
		return err
	}
	results, err := search.Query(context.Background(), cfg, query, effectiveCount(n, cfg.DefaultCount), nil)
	if err != nil {
		return err
	}
	if n > len(results) {
		return fmt.Errorf("only %d results; nothing at %d for %q", len(results), n, query)
	}
	return printOpened(n, results[n-1])
}

// effectiveCount ensures the search pulls at least n results so the requested
// result can exist.
func effectiveCount(n, floor int) int {
	if n > floor {
		return n
	}
	return floor
}

// printOpened opens the result in the browser, or prints title and URL when no
// browser could be launched.
func printOpened(n int, r engines.Result) error {
	if !openable(r.URL) {
		return fmt.Errorf("refusing to open %q", r.URL)
	}
	if openBrowser(r.URL) {
		fmt.Printf("opened %d. %s\n", n, r.Title)
		return nil
	}
	fmt.Printf("%d. %s\n%s\n", n, r.Title, r.URL)
	return nil
}

// parseIndex splits a leading integer argument (the result to open) from the
// rest of the query. Without an integer, the first result is implied.
func parseIndex(args []string) (int, []string, error) {
	if len(args) == 0 {
		return 1, args, nil
	}
	if idx, err := strconv.Atoi(args[0]); err == nil {
		if idx < 1 {
			return 0, nil, fmt.Errorf("result index must be at least 1")
		}
		return idx, args[1:], nil
	}
	return 1, args, nil
}

// openable reports whether a URL is one a browser should be asked to open. Only
// http and https qualify; xdg-open dispatches on scheme and would launch an
// arbitrary registered handler for anything else.
func openable(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return false
	}
	return parsed.Scheme == "https" || parsed.Scheme == "http"
}

// openBrowser hands a URL to the platform's default browser. Best effort: false
// means nothing was launched and the caller should print the URL instead.
func openBrowser(rawURL string) bool {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", rawURL)
	case "windows":
		command = exec.Command("cmd", "/C", "start", "", rawURL)
	default:
		command = exec.Command("xdg-open", rawURL)
	}
	return command.Start() == nil
}
