package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/FacileStudio/sonar/internal/mcpserver"
)

var installAll bool

var installCmd = &cobra.Command{
	Use:   "install [agent...]",
	Short: "Register the search MCP server in agent harnesses",
	Long: "Register sonar's MCP server in the agent harnesses on this machine, " +
		"so each one can spawn 'sonar mcp' instead of the model shelling out to " +
		"the binary.\n\nWith no argument, registers for the harnesses that are " +
		"installed. --all registers for every supported harness. Agent names are " +
		"also accepted directly: claude, gemini, opencode, nacelle.",
	Args: cobra.ArbitraryArgs,
	RunE: runInstall,
}

// agents returns the harnesses sonar can declare itself in, in display order.
func agents() []mcpserver.Agent {
	home, _ := os.UserHomeDir()
	return []mcpserver.Agent{
		mcpserver.Agent{Name: "claude", Path: filepath.Join(home, ".claude.json"), Key: "mcpServers",
			Marker: filepath.Join(home, ".claude")},
		mcpserver.Agent{Name: "gemini", Path: filepath.Join(home, ".gemini", "settings.json"), Key: "mcpServers",
			Marker: filepath.Join(home, ".gemini")},
		mcpserver.Agent{Name: "opencode", Path: filepath.Join(home, ".config", "opencode", "opencode.json"),
			Key: "mcp", Marker: filepath.Join(home, ".config", "opencode")},
		mcpserver.Agent{Name: "nacelle", Path: filepath.Join(home, ".nacelle.yml"), Key: "",
			Marker: filepath.Join(home, ".nacelle.yml")},
	}
}

// installed reports whether the harness's own directory is on this machine; a
// marker that exists proves the tool has run here once.
func installed(a mcpserver.Agent) bool {
	_, err := os.Stat(a.Marker)
	return err == nil
}

// targets returns the agents install should touch: named ones (validated),
// every one with --all, or whatever is installed when nothing is said.
func targets(args []string, all bool) ([]mcpserver.Agent, error) {
	known := agents()
	if all {
		return known, nil
	}
	if len(args) > 0 {
		var chosen []mcpserver.Agent
		for _, name := range args {
			agent, ok := findAgent(known, name)
			if !ok {
				return nil, fmt.Errorf("unknown harness %q (available: %s)", name, available(known))
			}
			chosen = append(chosen, agent)
		}
		return chosen, nil
	}
	var found []mcpserver.Agent
	for _, a := range known {
		if installed(a) {
			found = append(found, a)
		}
	}
	return found, nil
}

func findAgent(known []mcpserver.Agent, name string) (mcpserver.Agent, bool) {
	for _, a := range known {
		if a.Name == name {
			return a, true
		}
	}
	return mcpserver.Agent{}, false
}

func available(known []mcpserver.Agent) string {
	var names []string
	for _, a := range known {
		names = append(names, a.Name)
	}
	return strings.Join(names, ", ")
}

func runInstall(cmd *cobra.Command, args []string) error {
	chosen, err := targets(args, installAll)
	if err != nil {
		return err
	}
	if len(chosen) == 0 {
		lipgloss.Println("sonar: no agent harnesses detected; pass names or --all to register anyway")
		return nil
	}
	var refused []string
	for _, a := range chosen {
		changed, err := mcpserver.Declare(a)
		switch {
		case err != nil:
			refused = append(refused, a.Name+" ("+err.Error()+")")
			lipgloss.Println("  " + a.Name + ": " + err.Error())
		case changed:
			lipgloss.Println("  " + a.Name + " → " + a.Path)
		default:
			lipgloss.Println("  " + a.Name + " already declared (" + a.Path + ")")
		}
	}
	if len(refused) > 0 {
		return fmt.Errorf("MCP registration refused for: %s", strings.Join(refused, ", "))
	}
	return nil
}

func init() {
	installCmd.Flags().BoolVar(&installAll, "all", false, "register for every harness, installed or not")
	rootCmd.AddCommand(installCmd)
}
