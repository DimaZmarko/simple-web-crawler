# 005 — Lifecycle control

- Status: draft
- Verification tier: Rigorous (auto — score 4: cancellation/idempotency concurrency +2, persisted terminal state +2)
- Depends on: 003

## Goal
Give users control over a running crawl: stop it on demand, and enforce the limits (max pages, max depth, overall timeout) server-side so a crawl always terminates.

## User scenarios
- As a user, I cancel a running crawl and it stops quickly, keeping whatever it already found.
- As a user, I trust that a crawl won't run forever — it stops at its page/depth/time limits on its own.

## API surface (contract delta)
- `DELETE /crawls/{id}` — request cancellation. `202` `{ id, status: "cancelling" }` then converges to `cancelled`. Idempotent: cancelling an already-finished or already-cancelled crawl returns `200`/`202` with the terminal status, never an error.
- Status enum expands: `cancelling`, `cancelled`.

## Behavior & rules
- Cancellation calls the crawl's context `cancel()`; in-flight workers observe `ctx.Done()` and exit promptly (no killing goroutines).
- Partial results are retained and remain browsable via 004.
- Limit enforcement: a crawl auto-finishes (`completed`) when `maxPages` is reached or all in-depth URLs are exhausted; an overall `timeout` (config or default) cancels it as `cancelled`/`failed` with a reason.
- Race between natural completion and a cancel request resolves deterministically to a single terminal state.

## Data
- Persist terminal status and a `reason`/`finishedAt`. No new tables.

## Concurrency notes
- Cancellation propagation through the context tree built in 003.
- Idempotency guard so concurrent cancels / a cancel landing as the crawl completes don't double-transition.

## Frontend scope
- Cancel button on the crawl detail with optimistic `cancelling` state, settling to `cancelled`.
- Disable cancel for terminal crawls.

## Out of scope
- Auth on the cancel route (006 adds protection), metrics (007).

## Acceptance criteria
- Cancelling a running crawl moves it to `cancelled` within a small bound and stops further fetches; partial links remain.
- A crawl hitting `maxPages` stops on its own as `completed`; a crawl exceeding `timeout` ends with a clear reason.
- Cancelling a finished crawl is a no-op (idempotent), not an error.
- `go test -race ./...` clean, including a test that cancels mid-crawl and asserts prompt, clean shutdown.

## Open questions
- Default overall timeout value, and is it per-crawl configurable in `POST /crawls` (would amend 002's request schema)?
