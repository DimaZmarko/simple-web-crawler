// Package config loads runtime configuration from the environment.
package config

import (
	"errors"
	"os"
	"strings"
)

// Config holds the runtime configuration for the API server.
type Config struct {
	DatabaseURL string
	Port        string
	// AllowedOrigins are the browser origins permitted by CORS. The web app
	// calls the API cross-origin (page on :3000, API on :8080), so the browser
	// requires Access-Control-Allow-Origin on the response.
	AllowedOrigins []string
}

// Load reads configuration from the environment. DATABASE_URL is required;
// PORT defaults to "8080" and CORS_ALLOWED_ORIGINS to the local web origin.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		Port:           os.Getenv("PORT"),
		AllowedOrigins: parseOrigins(os.Getenv("CORS_ALLOWED_ORIGINS")),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}

// parseOrigins splits a comma-separated origin list, trimming blanks. When
// unset it defaults to the local Next.js dev origin.
func parseOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{"http://localhost:3000"}
	}

	var origins []string
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}
