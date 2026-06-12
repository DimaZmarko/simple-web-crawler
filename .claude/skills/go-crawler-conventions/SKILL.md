---
name: go-crawler-conventions
description: Go backend conventions for the crawler service in apps/api. Use when writing or modifying Go code — handlers, the dispatcher, the worker pool, queries, or tests. Covers concurrency patterns, data access, error handling, and test structure specific to this project.
allowed-tools: Read, Edit, Bash, Grep, Glob
---

# Go conventions — crawler service

## Concurrency
- The dispatcher is the single owner of the visited set. Workers never touch it directly; they send discovered links back on the results channel. No mutex on the hot path.
- Counters (pages fetched, errors, bytes) use `sync/atomic`. Never an unguarded int.
- Per-domain rate limiters live in a `sync.Map` keyed by host; create with `LoadOrStore` to win the create race.
- Every blocking call (DB, HTTP fetch, limiter `Wait`) takes a `context.Context` derived from the request or crawl context. Cancellation must propagate; a worker must exit promptly when the context is done.
- Goroutines have an explicit exit path. Use `errgroup` for the worker pool and `signal.NotifyContext` for graceful shutdown.

## Data access
- Postgres via `pgx/v5`. Queries via `sqlc` — write SQL in `apps/api/db/queries`, run `sqlc generate`, call the generated methods. Do not hand-write query structs.
- Migrations via `golang-migrate` in `apps/api/db/migrations`, sequential and reversible.

## HTTP layer
- `chi` router. Middleware chain: request ID, structured logging (`zerolog`), recoverer, auth where required, rate limit on `POST /crawls`.
- Handlers are thin: decode, validate, call a service, encode. Business logic lives in services, not handlers.
- Responses match `openapi.yaml` exactly. Annotate handlers with swaggo comments so `make swagger` regenerates docs in sync.

## Errors
- Wrap with `fmt.Errorf("...: %w", err)`. Never discard an error. Map typed domain errors to HTTP status in one place.

## Tests
- Table-driven with `t.Run` subtests.
- API tests use `net/http/httptest` in-process.
- Integration tests use `testcontainers-go` against a real Postgres — do not mock the DB layer.
- Run `make test-api` from the repo root (includes `-race`). A change is not done until the race detector is clean.
