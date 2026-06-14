# 002 — Crawl intake & storage — Plan

- Status: **draft** — Standard tier skips the `plan-verifier` gate (escalate to Rigorous and run it only if planning surfaces new concurrency, auth, or migration risk)
- Verification tier: Standard (unchanged from spec)
- Depends on: 001 (✅ done)

## Summary

Add the `crawls` resource end-to-end: flesh out the existing `POST /crawls` stub in `openapi.yaml`, persist rows in Postgres, expose list/detail endpoints, and ship submit + list + detail UI. No engine — status stays `queued`.

## Sides touched

| Side | Changes |
| --- | --- |
| Contract | Replace stub `POST /crawls`; add `GET /crawls`, `GET /crawls/{id}`; add schemas + error shapes |
| Backend | Migration, sqlc queries, service layer, chi handlers, httptest + testcontainers integration |
| Frontend | Submit form, paginated list, detail page, component tests against generated client types |

## Contract delta (exact)

Bump `info.version` minor (0.2.0 → 0.3.0).

### Schemas (`components/schemas`)

- **`CreateCrawlRequest`** — `seedUrl` (uri, http/https), `maxDepth` (integer ≥ 0), `maxPages` (integer ≥ 1, max e.g. 10_000).
- **`CrawlStatus`** — enum: `queued` only for now.
- **`Crawl`** — `id` (uuid), `seedUrl`, `maxDepth`, `maxPages`, `status` (`CrawlStatus`), `createdAt`, `updatedAt` (date-time, UTC).
- **`CrawlSummary`** — `id`, `seedUrl`, `status`, `createdAt`.
- **`CrawlList`** — `items` (`CrawlSummary[]`), `nextCursor` (string, nullable).
- **`ValidationError`** / **`FieldError`** — for `400` field-level errors (`field`, `message`).

### Operations

| Method | Path | Success | Errors |
| --- | --- | --- | --- |
| `POST` | `/crawls` | `202` `Crawl` (subset: id, status, createdAt + echo config) | `400` `ValidationError` |
| `GET` | `/crawls` | `200` `CrawlList` | — |
| `GET` | `/crawls/{id}` | `200` `Crawl` | `404` problem response |

Query param on list: `cursor` (opaque string, optional), `limit` (default 20, max 100).

### Cursor encoding (resolves open question)

Opaque base64url token encoding `(created_at DESC, id DESC)` tuple — stable keyset pagination, no offset scans.

## Backend approach

### Persistence

Migration `000002_crawls`:
```sql
CREATE TABLE crawls (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  seed_url    TEXT NOT NULL,
  max_depth   INT  NOT NULL CHECK (max_depth >= 0),
  max_pages   INT  NOT NULL CHECK (max_pages >= 1),
  status      TEXT NOT NULL CHECK (status = 'queued'),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX crawls_created_at_id_idx ON crawls (created_at DESC, id DESC);
```

### Data access

sqlc queries in `apps/api/db/queries/crawls.sql`:
- `CreateCrawl`
- `ListCrawls` (keyset: `WHERE (created_at, id) < ($cursor_created_at, $cursor_id)` ORDER BY created_at DESC, id DESC LIMIT $n`)
- `GetCrawlByID`

Wire generated querier into `cmd/server` via existing pgx pool.

### HTTP layer

- Package `internal/api/crawls.go` (+ `_test.go`): thin handlers delegating to `internal/crawl/store` or `internal/crawl/service`.
- Validation in service: parse `seedUrl` as absolute http(s); enforce bounds on depth/pages.
- Map domain errors → `400` / `404` in one place; responses match generated oapi types.
- Register routes on existing chi router; extend CORS if needed (already allows POST).
- swaggo annotations on new handlers; `make swagger` after implementation.

### Concurrency

Request-scoped `context.Context` on all DB calls (inherits from 001 pattern). No shared mutable crawl state — inserts/queries only.

### Security (deferred)

No auth (006). Basic URL syntax validation only — SSRF guard deferred to 006.

## Frontend approach

### Routes

- `/crawls/new` — submit form (`"use client"` for interactivity).
- `/crawls` — paginated list (TanStack Query + generated client).
- `/crawls/[id]` — detail view.

### Data fetching

Extend `src/api/client.ts` usage only — no hand-written types. TanStack Query hooks:
- `useCreateCrawl` mutation → `POST /crawls`, navigate to detail on `202`.
- `useCrawlList(cursor)` → `GET /crawls`.
- `useCrawl(id)` → `GET /crawls/{id}`.

Handle `400` field errors in the form; `404` on detail with a not-found state.

### Tests

- Form: valid submit calls client mock, shows success; invalid URL shows field error.
- List: renders items from mocked list response; "load more" passes cursor.
- Detail: renders config + `queued` status from mocked `Crawl`.

## Verification map (acceptance criteria → evidence)

| Criterion | Planned verification |
| --- | --- |
| Valid submit persists + `202 queued` | httptest POST + testcontainers insert/select |
| Invalid input → `400` field errors | httptest table cases (bad URL, maxPages=0) |
| List/detail reflect DB; pagination past one page | testcontainers: insert 25 rows, list with cursor |
| Responses match contract exactly | oapi response types in handlers; optional contract enum test |
| Backend `-race` clean | `make test-api` |
| Frontend form + list tests | `make test-web` |

## Out of scope (unchanged)

Engine execution, live stats, cancel, auth, SSRF hardening.

## Risks / notes

- Single-tenant (no `user_id` column) until 006 — document in migration comment.
- Existing `POST /crawls` stub must be fully replaced, not layered on.
