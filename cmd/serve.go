package cmd

import (
	"fmt"
	"os/signal"
	"syscall"

	"github.com/FacileStudio/sonar/internal/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the hosted sonar: web UI, JSON API and image proxy",
	Long: "Run the hosted sonar server, the one deployed at sonar.facile.studio.\n\n" +
		"It signs people in through Registre (OIDC_ISSUER, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET), " +
		"keeps per-user daily quotas in PostgreSQL (DATABASE_URL), runs searches on the engine keys in " +
		"SONAR_<ENGINE>_KEY and serves the web client from CLIENT_DIR. Queries are never logged or stored. " +
		"See docs/hosted.md for every setting.",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		if code := server.Run(ctx, server.Options{Version: version}); code != 0 {
			return fmt.Errorf("sonar serve exited with status %d", code)
		}
		return nil
	},
}
