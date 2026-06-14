# 002 — Crawl intake & storage

- Status: draft — not started. Note: `openapi.yaml` carries a pre-existing **unimplemented** stub `POST /crawls` (`operationId: createCrawl`, bare `202`, no request/response schema) from the initial scaffold — no backend handler, route, table, or `GET` endpoints exist. Flesh that stub out here (add `CreateCrawlRequest`/`Crawl`/`CrawlSummary` and the two `GET`s); it is not work already done.
- Verification tier: Standard (auto — score 3: persisted data +2, spans BE/FE +1)
- Depends on: 001 (✅ done)

## Goal
Accept, persist, list, and fetch crawl requests. No crawling happens yet — submitted crawls sit in `queued`. This establishes the crawl resource, its persistence, and the submit/list UI that later specs build on.

## User scenarios
- As a user, I submit a crawl (seed URL, max depth, max pages) and get back an id and a `queued` status.
- As a user, I see a list of my crawls with their status and creation time.
- As a user, I open a single crawl and see its configuration and current status.

## API surface (contract delta)
- `POST /crawls` — body `{ seedUrl, maxDepth, maxPages }`. Validates input. `202` `{ id, status: "queued", createdAt }`. `400` on invalid input (bad URL, non-positive limits).
- `GET /crawls` — paginated (cursor). `200` `{ items: [CrawlSummary], nextCursor }`.
- `GET /crawls/{id}` — `200` `Crawl` (id, seedUrl, maxDepth, maxPages, status, createdAt, updatedAt). `404` if unknown.
- Schemas: `Crawl`, `CrawlSummary`, `CreateCrawlRequest`.

## Behavior & rules
- Validation: `seedUrl` must be a syntactically valid absolute http(s) URL; `maxDepth` ≥ 0; `maxPages` ≥ 1 with a sane upper bound.
- Status enum introduced: `queued` (others added in 003/005). Persist all timestamps in UTC.

## Data
- `crawls` table: `id` (uuid pk), `seed_url`, `max_depth`, `max_pages`, `status`, `created_at`, `updated_at`. Index on `status`, `created_at` for listing.
- sqlc queries for insert / list (cursor) / get.

## Concurrency notes
- None yet beyond request-scoped context on the inserts/queries.

## Frontend scope
- Submit form with validation and success/error feedback.
- Crawl list page (paginated) linking to a crawl detail page showing config + status.

## Out of scope
- Executing the crawl, results, live stats, cancellation, auth.

## Acceptance criteria
- Submitting a valid request persists a row and returns `202` with a `queued` status; invalid input returns `400` with field errors.
- List and detail reflect persisted data; pagination works past one page.
- Responses match the contract exactly (no undocumented fields).
- Backend tests (httptest + a testcontainers integration test for persistence) pass with `-race`; frontend has a form + list component test.

## Open questions
- Cursor encoding scheme (created_at + id vs opaque token)?
- Per-user ownership now, or defer until auth (006)? Default: defer, single tenant for now.
