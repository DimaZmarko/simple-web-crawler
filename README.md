# Simple Web Crawler

A concurrent web-crawler service built as a small monorepo: a Go REST API backend and a
Next.js/React frontend, developed in parallel from a single shared API contract.

The API accepts crawl jobs and runs them through a dispatcher + worker pool with per-domain
rate limiting, deduplication, and cancellation; the web app submits crawls and watches them
live. The build is **spec-driven** — each vertical slice is specified, planned, and
implemented in order (see [`specs/`](specs/)).

> **Status:** bootstrap slice complete — `docker compose watch` boots Postgres → migrations →
> API → web with a working liveness/readiness health check. The crawl engine and UI are built
> on top of this in later specs. See [`specs/README.md`](specs/README.md) for the roadmap.

## Architecture

```
Browser ──► Next.js web (:3000) ──► Go API (:8080) ──► Postgres (:5432)
                                     chi · pgx/v5 · zerolog · dispatcher + worker pool
```

- **`packages/api-contract/openapi.yaml`** is the single source of truth. The TypeScript client
  and Go types are **generated** from it (`make gen-contract`) — never hand-edited. Any
  request/response or status-code change starts in the contract.
- Backend and frontend share only the contract, so the two tracks can be built independently.

## Repo layout

| Path | What lives here |
| --- | --- |
| `apps/api` | Go backend — chi router, dispatcher, worker pool, sqlc queries, migrations |
| `apps/web` | Next.js (App Router) frontend — consumes the generated client |
| `packages/api-contract` | `openapi.yaml`, the shared HTTP contract |
| `specs/` | Spec-driven build: one folder per vertical slice |

## Prerequisites

- **Docker** + a running engine (Colima recommended on macOS — see [`DOCKER.md`](DOCKER.md))
- **Go 1.25+** on the host (backend tests/build run on the host)
- Node/pnpm are **not** required on the host — the web toolchain runs in a Node 22 container

## Quickstart

```bash
make dev      # start db → migrate → api → web with live reload
```

Then:

| Service | URL |
| --- | --- |
| Web | http://localhost:3000 |
| API | http://localhost:8080 |
| Postgres | localhost:5432 (user/pass/db: `crawler`) |

```bash
curl localhost:8080/healthz   # {"status":"ok"}
curl localhost:8080/readyz    # {"status":"ok","checks":{"db":"ok"}}  (503 if the DB is down)
```

Stop with `make down` (keeps data) or `make clean` (removes the db volume).

## Common commands

Run `make help` for the full list. The essentials:

| Command | Does |
| --- | --- |
| `make dev` / `make watch` | Run the whole stack with live reload |
| `make up` / `make down` | Start detached / stop the stack |
| `make check` | All quality gates: lint + type check + tests |
| `make test` | Backend (`go test -race`) and frontend (Vitest) tests |
| `make lint` | `go vet` + gofmt check, and ESLint on the frontend |
| `make fmt` | Format Go and frontend code in place |
| `make gen-contract` | Regenerate the Go types + TS client from `openapi.yaml` |
| `make swagger` | Regenerate swagger docs from swaggo annotations |
| `make migrate` / `make migrate-down` | Apply / roll back DB migrations |

## Conventions

The golden rules (contract-first, never hand-edit generated code, concurrency verified with
`-race`, definition of done) live in [`CLAUDE.md`](CLAUDE.md); language-specific conventions
are in `.claude/skills/`.
