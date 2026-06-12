# 001 — Foundation & health

- Status: **done** — implemented & verified 2026-06-12 (working tree, not yet committed)
- Verification tier: Standard (auto — score 0; deterministic floor + code-reviewer; criteria are machine-checkable)
- Depends on: —

## Goal
Stand up the walking skeleton so every later spec has a working chain to extend. After this spec, `docker compose watch` boots Postgres → migrations → Go API → Next.js web, and a health check proves the request → DB path end to end.

## User scenarios
- As a developer, I run `docker compose watch` and get all three services healthy with live reload.
- As an operator, I can probe liveness and readiness for orchestration.
- As a user, I open the web app and see whether the API and its database are reachable.

## API surface (contract delta)
- `GET /healthz` — liveness. Always `200` with `{ "status": "ok" }` while the process is up.
- `GET /readyz` — readiness. `200` `{ "status": "ok", "checks": { "db": "ok" } }` when the DB ping succeeds; `503` `{ "status": "degraded", "checks": { "db": "down" } }` otherwise. ⚠️ **As-built uses `"db": "down"`, not `"fail"`** — the locked contract enum is `[ok, down]` and the implementation + tests match it. This spec line is corrected to reflect reality; if you truly want `"fail"`, that's a breaking contract change to drive through api-contract-keeper.

## Behavior & rules
- `/readyz` pings the Postgres pool with a short context timeout; a slow/dead DB yields `503`, never a hang.
- Standard middleware chain present from day one: request id, structured logging (zerolog), panic recovery.
- Graceful shutdown via `signal.NotifyContext` + `http.Server.Shutdown`.

## Data
- One initial migration so the `migrate` service succeeds against a non-empty dir (e.g. enable `pgcrypto`). No domain tables yet.

## Concurrency notes
- Establish the shutdown pattern and the request-scoped context that all later DB calls will derive from.

## Frontend scope
- App shell with `output: 'standalone'`, a TanStack Query provider, and one page that calls `/readyz` through the generated client and renders ok / degraded.

## Out of scope
- Any crawl functionality, auth, metrics.

## Acceptance criteria
- [x] `docker compose watch` brings up db → migrate (completes, `1/u enable_pgcrypto`) → api → web with no errors.
- [x] `GET /healthz` → 200 `{"status":"ok"}`; `GET /readyz` → 200 with db check passing; stopping Postgres flips `/readyz` to **503 `{"checks":{"db":"down"},"status":"degraded"}`** and it recovers on restart (verified live via `docker-compose stop db`).
- [x] Web page reflects the `/readyz` result (renders "API: ok" / "API: degraded"). *Verified via served page + the cross-origin `/readyz` fetch returning ok+CORS + the component test; no headed-browser capture (Node isn't on the host).*
- [x] `go test -race ./...` passes (httptest tests for both endpoints, plus a CORS test and a contract-enum guard); `make swagger` generates cleanly.
- [x] Saving a Go handler (`air` rebuild) or a web component (`next dev` recompile) live-reloads under watch.

## Open questions
- ~~Go module path (org/repo)?~~ → **`github.com/DimaZmarko/simple-web-crawler/apps/api`**.
- Auth provider direction for later specs (generic bearer vs Keycloak) — still open; does not block 001 (revisited in 006).

## Implementation notes & deviations
- **Go 1.25, not 1.23** — `pgx/v5@latest` and `air@latest` now require Go ≥ 1.25; both Dockerfiles use `golang:1.25-alpine`.
- **CORS added (not in the draft)** — the web app calls the API cross-origin (`:3000` → `:8080`), so `go-chi/cors` is the first middleware (origins from `CORS_ALLOWED_ORIGINS`, default `http://localhost:3000`), with a test. Later specs adding non-GET/POST methods must widen the allowed methods/headers in lockstep with the contract.
- **Web toolchain runs in Docker** (no host Node/pnpm); pnpm pinned to `9.15.9`. `make gen-contract` = `oapi-codegen` (Go types) + `openapi-typescript` (TS client, in a container).
- Contract bumped to `0.2.0` (additive: `/healthz`, `/readyz`, `HealthStatus`, `ReadinessStatus`).
