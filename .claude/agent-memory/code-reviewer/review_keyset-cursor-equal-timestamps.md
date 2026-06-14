---
name: review-keyset-cursor-equal-timestamps
description: keyset pagination over (created_at, id) is only collision-safe if created_at round-trips exactly through the cursor; integration tests that sleep between inserts hide ties
metadata:
  type: feedback
---

When reviewing keyset/cursor pagination on `crawls` (and later resources), the cursor encodes `(created_at, id)` and the query seeks with `(created_at, id) < (cursor)`. Two correctness preconditions:

1. **Timestamp must round-trip exactly.** The Go cursor JSON-encodes `time.Time` (RFC3339 nanos) while Postgres `timestamptz` is microsecond precision. It works today only because the value is read back *from* Postgres (already microsecond-truncated) before being put in the cursor — so encode/decode is lossless. If a future change ever builds a cursor from an in-memory `time.Time` that hasn't been through Postgres, sub-microsecond digits would make the seek skip or duplicate the boundary row. Flag any cursor built from a non-DB timestamp.

2. **Equal `created_at` is the real risk, and tests don't cover it.** `internal/crawl/integration_test.go` sleeps 2ms between inserts so every `created_at` is distinct — the `id` tie-breaker in the keyset is never exercised. Bulk inserts (or `now()` firing twice in the same microsecond) can produce equal `created_at`; only then does correct `id` DESC tie-breaking matter. A row-comparison `(created_at, id) < (c, d)` handles it correctly, but the test gives no evidence. Recommend an integration case that inserts a batch sharing one `created_at`.

**Why:** keyset bugs are silent — they skip or duplicate exactly one row at a page boundary and never error. The httptest layer uses canned rows so it can't catch this; only a testcontainers test with colliding timestamps can.

**How to apply:** on any cursor-pagination review, check (a) the cursor timestamp originates from a DB read, and (b) a test exercises equal sort-key values, not just monotonic ones.
