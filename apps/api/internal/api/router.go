// Package api wires the HTTP router, middleware, and handlers for the crawler service.
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
)

// NewRouter builds the chi router with the standard middleware chain and mounts
// the health endpoints. pinger backs the readiness probe. allowedOrigins are the
// CORS origins permitted for browser calls; when empty, all origins are allowed
// (suitable for tests).
func NewRouter(pinger Pinger, logger zerolog.Logger, allowedOrigins ...string) *chi.Mux {
	r := chi.NewRouter()

	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.RequestID)
	r.Use(requestLogger(logger))
	r.Use(middleware.Recoverer)

	r.Get("/healthz", Healthz)
	r.Get("/readyz", Readyz(pinger, logger))

	return r
}
