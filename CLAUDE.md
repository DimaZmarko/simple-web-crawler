# Simple Web Crawler — monorepo

A concurrent web-crawler service. Go REST API backend + Next.js/React frontend, developed in parallel from a shared API contract.

## Repo map

| Path | What lives here |
| --- | --- |
| `apps/api` | Go backend — chi router, dispatcher, worker pool, sqlc queries, migrations. Owns `apps/api/CLAUDE.md`. |
| `apps/web` | Next.js (App Router) frontend. Consumes a generated TypeScript client. Owns `apps/web/CLAUDE.md`. |
| `packages/api-contract` | `openapi.yaml` — the single source of truth for the HTTP API. Generated clients/types derive from it. |
| `specs/` | Feature specs from the spec-workflow skill (Specify → Plan → Tasks). |
| `.claude/agents` | Subagents (api-contract-keeper, backend-engineer, frontend-engineer, test-runner, code-reviewer, plan-verifier, spec-verifier). |
| `.claude/skills` | Workflows (spec-workflow, parallel-dispatch), conventions (go, nextjs, api-contract), and `prepare-pr` (draft a PR summary from the branch diff). |
| `.claude/README.md` | Routing decision tree — fast path vs Standard/Rigorous tiers. Start here when deciding how much process a change needs. |
| `.claude/agent-memory` | Per-agent durable notes that persist across sessions. Agents append recurring lessons here (e.g. the code-reviewer's checklist gotchas); read at the start of an agent's run so reviews and implementations get sharper over time. |

## Golden rules

1. **Contract-first, always.** Any change to a request/response shape, status code, or endpoint starts in `packages/api-contract/openapi.yaml` via the api-contract-keeper agent. Never let backend and frontend drift; the contract is the only coordination channel between the two tracks.
2. **Never hand-edit generated code.** The TS client (`apps/web/src/api/__generated__`) and any generated Go types come from codegen. Edit the contract and regenerate.
3. **Concurrency is verified, not assumed.** Backend tests run with `-race`. A change touching goroutines, channels, the dispatcher, or shared state is not done until `make test-api` is green.
4. **Tasks are scoped to one app where possible.** A task that only touches `apps/api` or only `apps/web` should be dispatched to the matching engineer agent so it runs in an isolated worktree and can proceed in parallel.
5. **Definition of done** = code + tests passing + contract in sync + swagger regenerated (backend) / client regenerated (frontend).

## Toolchain

Common commands (see `Makefile` — **prefer these over raw `go`/`pnpm` so agents work without local Node**):
- `make dev` — run api + web together
- `make test` / `make test-api` / `make test-web` — backend (`-race`) and frontend (Vitest in Docker)
- `make test-e2e` — Playwright e2e smoke tests against a running stack (`make up` first); not part of `make check`
- `make check` — lint + typecheck + tests
- `make gen-contract` — regenerate TS client + Go types from `openapi.yaml`
- `make swagger` — regenerate swagger from swaggo annotations into the contract
- `make migrate` — apply DB migrations

## Agent roster

- **api-contract-keeper** — owns `openapi.yaml`. Run it first when an interface changes.
- **backend-engineer** — Go work in `apps/api`, isolated worktree.
- **frontend-engineer** — Next.js work in `apps/web`, isolated worktree.
- **test-runner** — runs `make test*` and returns only failures (keeps verbose output out of the main context).
- **code-reviewer** — read-only review against the checklist before merge.
- **plan-verifier** — read-only pre-code gate (Rigorous tier): checks `plan.md` covers every acceptance criterion and contract delta.
- **spec-verifier** — read-only post-implementation gate (Rigorous tier): checks the diff against `spec.md` acceptance criteria.

Agents with `memory: project` keep durable notes under `.claude/agent-memory/<agent>/` — recurring review findings, codepath locations, architectural decisions. These carry across sessions, so record a lesson there once instead of relearning it; the chat is not the durable record.

## How to build a feature

**Routing first:** see `.claude/README.md` for the decision tree. Small single-app changes (bug fixes, refactors, ≤ 3 files, no contract/migration/concurrency/auth) take the **fast path** — no spec artifacts. Everything else runs the full workflow at the **Standard** or **Rigorous** tier.

For the full workflow, invoke the `spec-workflow` skill to go Specify → Plan → Tasks → Implement. The Implement phase hands cross-cutting features to the `parallel-dispatch` skill, which updates the contract, then runs backend-engineer and frontend-engineer concurrently in separate worktrees, then reconciles.

## MCP servers

`.mcp.json` wires up one project-scoped MCP server. It depends on the local stack — it is inert against a stopped one:
- **postgres** — inspect the schema and query data directly. Runs the maintained `crystaldba/postgres-mcp` (Postgres MCP Pro) Docker image in `--access-mode=restricted` (read-only). It joins the compose network (`simple-web-crawler_default`) and reads credentials from the root `.env` via `--env-file`, so it reaches Postgres at `db:5432` exactly like the app does — bring the DB up first (`make up` or `make watch`). Useful for confirming migrations applied or eyeballing crawl rows during debugging. **Single source of truth:** DB credentials live only in `.env` (gitignored; template in `.env.example`); docker-compose (`db`/`migrate`/`api`) and this MCP server all derive from it. Copy `.env.example` to `.env` on first checkout.

## Local environment

Docker via Colima + docker-compose. Start the stack with live reload using `docker compose watch` (recommended on Colima) or `docker compose up --build`. Services: web on :3000, api on :8080, postgres on :5432. The dev override (`docker-compose.override.yml`) runs the `dev` build targets; the base file (`docker-compose.yml`) runs production-shaped targets. Backend integration tests (testcontainers-go) run on the host, not in compose. See `DOCKER.md` for Colima-specific guidance.
