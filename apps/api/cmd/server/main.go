// Command server is the entrypoint for the Simple Web Crawler API.
//
// @title       Simple Web Crawler API
// @version     0.2.0
// @description Single source of truth for the crawler HTTP API. Generated clients derive from this file.
// @BasePath    /
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/DimaZmarko/simple-web-crawler/apps/api/internal/api"
	"github.com/DimaZmarko/simple-web-crawler/apps/api/internal/config"
)

const shutdownTimeout = 10 * time.Second

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().Timestamp().Logger()

	if err := run(logger); err != nil {
		logger.Fatal().Err(err).Msg("server exited with error")
	}
}

func run(logger zerolog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// pgxpool.New does not connect eagerly, so an unreachable DB at boot is fine;
	// readiness is reported per-request by /readyz.
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	router := api.NewRouter(pool, logger, cfg.AllowedOrigins...)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info().Str("addr", server.Addr).Msg("server listening")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		logger.Info().Msg("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	logger.Info().Msg("server stopped cleanly")
	return nil
}
