# 004 — Live monitoring & results

- Status: draft
- Verification tier: Standard (auto — score 3: concurrent counter reads +2, spans BE/FE +1; mostly reads)
- Depends on: 003

## Goal
Let users watch a crawl as it runs and inspect what it found. Adds rich live stats and a paginated, filterable results view, with a dashboard that updates while a crawl is in progress.

## User scenarios
- As a user, I watch a running crawl's progress update without reloading.
- As a user, I browse the discovered links, page through them, and filter by status code or depth.

## API surface (contract delta)
- `GET /crawls/{id}` enriched stats: `{ pagesFetched, errors, maxDepthReached, bytesDownloaded, startedAt, finishedAt }`.
- `GET /crawls/{id}/results` — paginated discovered links; query params `cursor`, `limit`, `statusCode`, `minDepth`, `maxDepth`. `200` `{ items: [Link], nextCursor }`.
- Stretch: `GET /crawls/{id}/events` — Server-Sent Events stream of stat updates for live push instead of polling.

## Behavior & rules
- Stats are a consistent snapshot — reading mid-crawl never shows torn values (read atomics coherently).
- Results pagination is stable under concurrent writes (cursor keyed on `discovered_at, id`).
- Filters compose (status code AND depth range).

## Data
- Read queries over `links` with the filter/pagination predicates and supporting indexes.
- Stats read either from live in-memory counters (running) or persisted counters (finished) — the endpoint hides which.

## Concurrency notes
- Atomic counters read concurrently with worker writes; ensure no data race (covered by `-race`).
- If SSE is implemented: a per-crawl fan-out goroutine pushes updates to subscribers over channels; subscribers cleaned up on disconnect (context cancellation).

## Frontend scope
- Crawl detail becomes a live dashboard: progress (pages/errors/depth), updating via polling (default) or SSE (if implemented).
- Results table with pagination and the status-code / depth filters.

## Out of scope
- Cancellation (005), auth (006), durable batch-flush hardening and metrics (007).

## Acceptance criteria
- While a crawl runs, the dashboard reflects increasing counters; after completion the numbers reconcile exactly with the persisted `links` rows.
- Results paginate past one page and filters narrow correctly.
- `go test -race ./...` clean; frontend has a live-view and a results-table test against the generated client types.

## Open questions
- Polling interval vs SSE as the default? Default: polling for simplicity, SSE as stretch.
- Expose `bytesDownloaded` if the fetcher streams bodies — confirm it's tracked in 003.
