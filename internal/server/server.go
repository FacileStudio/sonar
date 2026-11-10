package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/FacileStudio/sonar/internal/server/migrations"
	"github.com/FacileStudio/sonar/internal/server/store"
	"github.com/FacileStudio/tronc/logger"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/FacileStudio/tronc/migrate"
)

// Options carries what the binary knows about itself.
type Options struct {
	Version string
}

// Run serves until ctx ends and returns the process exit code.
func Run(ctx context.Context, opts Options) int {
	redactQueries()
	cfg, err := Load()
	if err != nil {
		logger.New(logger.Config{}).Error("failed to load config", slog.Any("error", err))
		return 1
	}
	log := logger.New(logger.Config{Level: cfg.LogLevel})
	db, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to open the database", slog.Any("error", err))
		return 1
	}
	defer func() { _ = db.Close() }()
	if err := migrate.Run(ctx, migrate.Config{DB: db, FS: migrations.FS, Logger: log}); err != nil {
		log.Error("failed to run migrations", slog.Any("error", err))
		return 1
	}
	app, err := assemble(ctx, db, cfg, log)
	if err != nil {
		log.Error("failed to assemble the server", slog.Any("error", err))
		return 1
	}
	if app.mcp, err = newMCP(ctx, app, cfg, opts.Version); err != nil {
		log.Error("failed to build the mcp server", slog.Any("error", err))
		return 1
	}
	app.background(ctx)
	return serve(ctx, cfg.Port, buildRouter(app, cfg, opts, db), log)
}

// redactQueries adds the parameters that carry a query or an image URL to the
// request logger's redaction list: sonar logs that a search happened, never
// what it was for.
func redactQueries() {
	for _, key := range []string{"q", "u", "s", "url"} {
		if !containsKey(troncmiddleware.SensitiveQueryKeys, key) {
			troncmiddleware.SensitiveQueryKeys = append(troncmiddleware.SensitiveQueryKeys, key)
		}
	}
}

func containsKey(keys []string, key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}

func serve(ctx context.Context, port int, handler http.Handler, log *slog.Logger) int {
	server := &http.Server{
		Addr: ":" + strconv.Itoa(port), Handler: handler,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second,
		WriteTimeout: 60 * time.Second, IdleTimeout: 120 * time.Second,
	}
	failed := make(chan error, 1)
	go func() { failed <- server.ListenAndServe() }()
	log.Info("server starting", slog.String("addr", server.Addr))
	select {
	case err := <-failed:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server stopped", slog.Any("error", err))
			return 1
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			log.Error("server shutdown failed", slog.Any("error", err))
			return 1
		}
	}
	return 0
}
