# apps/api — Go backend

The concurrent crawler service. Go 1.25+, chi, pgx/v5, sqlc, golang-migrate, swaggo, zerolog.

Run: `make -C ../.. dev` (whole stack) or `go run ./cmd/server` from here.
Test: `make -C ../.. test-api` from repo root (wraps `go test -race ./...`; `-race` is required).

Layout: `cmd/server` (entrypoint), `internal/api` (handlers, middleware), `internal/crawl` (dispatcher, worker pool), `db/queries` + `db/migrations`.

Follow the `go-crawler-conventions` skill. Responses must match `packages/api-contract/openapi.yaml`; keep swaggo annotations in sync with `make swagger`.
