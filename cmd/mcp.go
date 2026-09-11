package cmd

import (
	"github.com/FacileStudio/sonar/internal/mcpserver"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Serve web search to an agent over MCP on stdio",
	Long: "Serve web search to an agent over MCP on stdio.\n\n" +
		"Tool: search. An agent launches this as a subprocess and speaks JSON-RPC " +
		"over its stdin and stdout, so nothing here prints to the terminal.\n\n" +
		"Wire it into your agent as a stdio server: command sonar, args [mcp].",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mcpserver.Serve(cmd.Context(), version)
	},
}
