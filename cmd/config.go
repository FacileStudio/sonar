package cmd

import (
	"fmt"
	"sort"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/search"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show the effective config: engines, keys, and tuning knobs",
	Args:  cobra.NoArgs,
	RunE:  runConfig,
}

func init() {}

// erow is one engine's table row plus the priority it sorts by, so rows are
// ordered by dispatch priority rather than map-hash iteration order.
type erow struct {
	prio  int
	cells []string
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath())
	if err != nil {
		return err
	}
	renderEngineTable(engineRows(cfg))
	printTuning(cfg)
	return nil
}

// engineRows builds the engine rows in priority order, appending each SearXNG
// instance by label. Keys are shown only as set/not, never their value.
func engineRows(cfg *config.Config) []erow {
	var rows []erow
	for name, def := range cfg.Engines {
		if def == nil {
			continue
		}
		rows = append(rows, erow{
			prio:  def.Priority,
			cells: []string{name, fmt.Sprintf("%d", def.Priority), engineState(name, def), keyState(def.APIKey)},
		})
	}
	for i := 0; i < len(cfg.Searxng); i++ {
		rows = append(rows, erow{
			prio:  cfg.SearxngPriority,
			cells: []string{search.SearxngName(cfg.Searxng, i), fmt.Sprintf("%d", cfg.SearxngPriority), "on", "-"},
		})
	}
	sort.SliceStable(rows, func(a, b int) bool { return rows[a].prio < rows[b].prio })
	return rows
}

// engineState classifies an engine for the config table: on, off, or on-but-
// missing its required key.
func engineState(name string, def *config.EngineDef) string {
	if !def.Enabled {
		return "off"
	}
	if search.NeedsKey(name) && def.APIKey == "" {
		return "no key"
	}
	return "on"
}

func keyState(key string) string {
	if key == "" {
		return "-"
	}
	return "set"
}

func renderEngineTable(rows []erow) {
	header := lipgloss.NewStyle().Foreground(lipgloss.Blue)
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(header).
		Headers("engine", "prio", "state", "key").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Blue)
			}
			if col == 2 || col == 3 {
				return lipgloss.NewStyle().Foreground(stateColor(rows[row].cells[col]))
			}
			return lipgloss.NewStyle()
		})
	for _, r := range rows {
		t.Row(r.cells...)
	}
	lipgloss.Println(t)
}

// stateColor colours the state and key columns: an active value (on, set) is
// green, anything else dim so it recedes.
func stateColor(s string) ansi.BasicColor {
	if s == "on" || s == "set" {
		return lipgloss.Green
	}
	return lipgloss.BrightBlack
}

// printTuning shows the block-avoidance knobs that are easy to lose in a config
// file, with bold blue keys matching the table header.
func printTuning(cfg *config.Config) {
	label := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Blue)
	dim := lipgloss.NewStyle().Faint(true)
	lipgloss.Println(fmt.Sprintf("%s %-10s %s %-8s %s %-8s %s %-8s\n%s %-8s %s %-8s\n%s %s",
		label.Render("count"), dim.Render(fmt.Sprintf("%d", cfg.DefaultCount)),
		label.Render("timeout"), dim.Render(fmt.Sprintf("%ds", cfg.TimeoutSeconds)),
		label.Render("minInterval"), dim.Render(fmt.Sprintf("%v", cfg.MinInterval)),
		label.Render("jitter"), dim.Render(fmt.Sprintf("%v", cfg.Jitter)),
		label.Render("breakerCooldown"), dim.Render(fmt.Sprintf("%v", cfg.BreakerCooldown)),
		label.Render("cacheTTL"), dim.Render(fmt.Sprintf("%v", cfg.CacheTTL)),
		label.Render("cacheDir"), dim.Render(cfg.CacheDir)))
}
