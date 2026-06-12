package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/DimaZmarko/simple-web-crawler/apps/api/internal/api/types"
)

// readyzTimeout bounds how long the readiness probe waits on the dependency check.
const readyzTimeout = 2 * time.Second

// Pinger is the minimal dependency the readiness probe needs to verify the
// database is reachable. *pgxpool.Pool satisfies this interface.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Healthz is the liveness probe. It always reports the process is up and serving.
//
// @Summary  Liveness probe
// @Tags     health
// @Produce  json
// @Success  200 {object} types.HealthStatus
// @Router   /healthz [get]
func Healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, types.HealthStatus{Status: types.HealthStatusStatusOk})
}

// Readyz returns a readiness handler that pings the supplied dependency.
//
// @Summary  Readiness probe
// @Tags     health
// @Produce  json
// @Success  200 {object} types.ReadinessStatus
// @Failure  503 {object} types.ReadinessStatus
// @Router   /readyz [get]
func Readyz(pinger Pinger, logger zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readyzTimeout)
		defer cancel()

		if err := pinger.Ping(ctx); err != nil {
			logger.Error().Err(err).Msg("readiness check failed: database ping error")

			resp := types.ReadinessStatus{Status: types.Degraded}
			resp.Checks.Db = types.ReadinessStatusChecksDbDown
			writeJSON(w, http.StatusServiceUnavailable, resp)
			return
		}

		resp := types.ReadinessStatus{Status: types.Ok}
		resp.Checks.Db = types.ReadinessStatusChecksDbOk
		writeJSON(w, http.StatusOK, resp)
	}
}

// writeJSON sets the content type, writes the status code, and encodes v as JSON.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
