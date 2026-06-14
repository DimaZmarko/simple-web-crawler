# 003 — Crawl engine

- Status: draft
- Verification tier: Rigorous (auto — score 6: concurrency +2, persistence +2, completion-correctness judgment +2; also an always-Rigorous concurrency-core trigger)
- Depends on: 002

## Goal
Execute queued crawls. This is the core: a dispatcher owning the visited set, a worker pool fetching and parsing pages, per-domain rate limiting, deduplication, and context-based cancellation. Crawls transition `queued → running → completed | failed`, and discovered links are persisted.

## User scenarios
- As a user, after I submit a crawl it starts running on its own and reaches `completed`.
- As a site owner of a crawled domain, my server is not hammered — requests to one host are rate limited.

## API surface (contract delta)
- No new endpoints. The `status` enum expands: `running`, `completed`, `failed`. `GET /crawls/{id}` now also returns coarse counters `{ pagesFetched, errors }` (full live stats are 004).

## Behavior & rules
- A scheduler picks up `queued` crawls and runs them.
- Fetch only same-scheme http(s); respect `maxDepth`; stop at `maxPages`.
- Deduplicate URLs within a crawl — a URL is fetched at most once.
- Politeness: at most N requests/second per host.
- On unrecoverable error the crawl ends `failed` with a reason; otherwise `completed`.

## Data
- `links` table: `id`, `crawl_id` (fk), `url`, `depth`, `status_code`, `discovered_at`, `fetched_at`. Unique `(crawl_id, url)` to enforce dedup at the DB layer as a backstop.
- Persist final counters on the crawl row.

## Concurrency notes (the heart of this spec)
- Single-owner dispatcher: only the dispatcher goroutine touches the visited set; workers send discovered links back over a results channel. No mutex on the hot path.
- Worker pool via `errgroup`; bounded concurrency; jobs and results channels (buffered for backpressure).
- Per-domain limiters in a `sync.Map`, created with `LoadOrStore`.
- Counters via `sync/atomic`.
- Context propagated to every fetch and DB call; cancellation stops workers promptly.
- Completion detection: the crawl ends only when no work is in flight and the queue is drained — design this explicitly (it is the classic bug).
- Verified with `go test -race`.

## Frontend scope
- Crawl list/detail reflect `running` / `completed` / `failed` and the coarse counters (poll on an interval; richer live view is 004).

## Out of scope
- Paginated results browsing and rich live stats (004), cancellation API (005), auth/SSRF (006).

## Acceptance criteria
- A submitted crawl runs to `completed`, fetching reachable pages within depth/page limits and deduping URLs.
- Requests to a single host are rate limited (observable in timing / a test double).
- `go test -race ./...` is clean, including a `testcontainers-go` integration test that drives a crawl against a local fixture server and asserts the persisted link set and final status.
- Killing the process mid-crawl does not corrupt state (crawl remains resumable-or-failed, no half-written rows).

## Open questions
- Single-process scheduler now, or queue-backed (RabbitMQ) for horizontal scale later? Default: in-process now, leave a seam.
- robots.txt honoring — in scope here or a later spec? Default: later; note it.
