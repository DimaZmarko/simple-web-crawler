# 006 — Access control & abuse protection

- Status: draft
- Verification tier: Rigorous (auto — score 5: security (auth, SSRF) +2, persisted +2, spans BE/FE +1; also an always-Rigorous auth/SSRF trigger)
- Depends on: 002, 003

## Goal
Protect the API and prevent the crawler from being abused. Adds authentication on write routes, rate limiting on crawl submission, and an SSRF guard so the crawler can't be pointed at internal infrastructure.

## User scenarios
- As an operator, only authenticated callers can create or cancel crawls.
- As an operator, a single caller can't flood the system with crawl submissions.
- As a security reviewer, I'm confident a user can't use the crawler to reach internal/metadata endpoints.

## API surface (contract delta)
- Auth required on write routes (`POST /crawls`, `DELETE /crawls/{id}`): bearer token in `Authorization`. `401` when missing/invalid, `403` when not permitted. Read routes' protection is configurable.
- Rate limit on `POST /crawls`: `429` with `Retry-After` when the per-principal/per-IP limit is exceeded.
- `POST /crawls` may now reject a seed URL that fails the SSRF guard with `400`/`422` and a clear reason.

## Behavior & rules
- Token validation via the chosen provider (generic bearer/JWT; Keycloak-compatible if that's the direction). Principal identity is attached to the request context and recorded on created crawls.
- Submission rate limit: token-bucket per principal (fallback per IP from `X-Forwarded-For`), limiters held in a `sync.Map`.
- SSRF guard (applies at submit time and again before each fetch in the engine): resolve the host and reject private, loopback, link-local, and cloud-metadata ranges; reject non-http(s) schemes and redirects that escape into those ranges.

## Data
- Optional `owner`/`principal` column on `crawls` (revisits the 002 ownership open question). Persist nothing sensitive (no raw tokens).

## Concurrency notes
- Rate-limiter map shares the `sync.Map` + `LoadOrStore` pattern from the engine's per-domain limiters.
- SSRF re-check inside the engine must run under the same request/crawl context.

## Frontend scope
- Attach the auth token to API calls; handle `401` (re-auth) and `429` (back off, surface Retry-After) gracefully.

## Out of scope
- Full identity/user management UI; metrics (007).

## Acceptance criteria
- Write routes reject unauthenticated/insufficient calls with `401`/`403`; authorized calls succeed.
- Exceeding the submission limit returns `429` with `Retry-After`.
- Seed URLs resolving to private/loopback/link-local/metadata ranges are rejected, and a redirect into those ranges mid-crawl is blocked (covered by tests).
- `go test -race ./...` clean; frontend handles 401/429 paths in tests.

## Open questions
- Confirm the auth provider (generic JWT vs Keycloak/OAuth2 from the project context) — drives the validation middleware.
- Allowlist override for trusted internal targets in non-prod?
