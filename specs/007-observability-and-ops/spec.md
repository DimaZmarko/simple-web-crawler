# 007 — Observability & operational hardening

- Status: draft
- Verification tier: Rigorous (auto — score 4: shutdown-drain concurrency +2, flusher correctness judgment +2)
- Depends on: 003, 004

## Goal
Make the service operable in production: metrics, correlated structured logs, durable stat persistence under load, and verified graceful shutdown. Closes the loop on the batch-flush pattern the engine relies on.

## User scenarios
- As an operator, I scrape metrics and see crawl throughput, fetch latency, queue depth, active workers, and error rate.
- As an operator, I trace a single request through structured, correlated logs.
- As an operator, a deploy/SIGTERM drains in-flight work cleanly with no lost stats.

## API surface (contract delta)
- `GET /metrics` — Prometheus exposition format. (Unauthenticated or internal-only per deployment.)
- `GET /version` — build/version info. `/healthz` and `/readyz` already exist (001).

## Behavior & rules
- Metrics: counters (pages fetched, errors, crawls by terminal status), histograms (fetch latency, request latency), gauges (active workers, queue depth, in-flight crawls).
- Logging: every request logged with request id, method, path, status, latency; logs from a crawl carry the crawl id; consistent fields across the service.
- Durable stats: a `time.Ticker` flusher drains the in-memory atomic counters to Postgres periodically and on shutdown, so a restart doesn't lose recent progress (formalizes what 003/004 read).
- Graceful shutdown: stop accepting requests, cancel running crawls' contexts, wait (bounded) for workers to drain, do a final stats flush, then exit.

## Data
- Reuses `crawls` counters; the flusher writes periodic snapshots. Possibly a small `crawl_stats` time-series table if historical charts are wanted (optional).

## Concurrency notes
- The flusher is a dedicated goroutine: `select` over the ticker and a done channel; one final flush on shutdown.
- Shutdown ordering matters: server stop → context cancel → worker drain (bounded by a deadline) → final flush.
- All verified under `-race`, ideally with a small load test.

## Frontend scope
- Minimal/none (optional ops panel). This spec is backend-weighted.

## Out of scope
- Distributed tracing backends, alerting rules, dashboards-as-code (note as follow-ups).

## Acceptance criteria
- `GET /metrics` exposes the listed series and they move under a real crawl.
- Logs are structured and correlated by request id and crawl id.
- A SIGTERM during an active crawl drains within the deadline, flushes final stats, and exits 0; persisted stats match what ran.
- `go test -race ./...` clean, including a shutdown-drain test.

## Open questions
- Keep the periodic `crawl_stats` snapshots for historical charts, or only flush latest counters? Default: latest counters; add snapshots only if 004 wants history.
