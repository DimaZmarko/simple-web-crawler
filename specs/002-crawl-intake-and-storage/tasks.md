# 002 — Crawl intake & storage — Tasks

Run in order within each group. `[contract]` must finish before `[backend]` and `[frontend]` parallel work begins.

## [contract]

- [ ] Replace stub `POST /crawls` in `packages/api-contract/openapi.yaml` with full request/response schemas (`CreateCrawlRequest`, `Crawl`, `ValidationError`, etc.)
- [ ] Add `GET /crawls` (cursor pagination) and `GET /crawls/{id}` with `200` / `404` responses
- [ ] Add `CrawlSummary`, `CrawlList`, `CrawlStatus` enum (`queued` only)
- [ ] Bump contract minor version; add examples on every operation
- [ ] Run `make gen-contract`; confirm TS schema + Go types regenerate cleanly

## [backend] — parallel with [frontend] after contract lands

- [ ] Migration `000002_crawls` (table + index); `make migrate`
- [ ] sqlc queries: `CreateCrawl`, `ListCrawls` (keyset cursor), `GetCrawlByID`; `sqlc generate`
- [ ] Service layer: validation (http(s) URL, depth ≥ 0, pages ≥ 1), cursor encode/decode
- [ ] Handlers: `POST /crawls`, `GET /crawls`, `GET /crawls/{id}`; wire into router
- [ ] Map errors to `400` / `404`; responses use generated oapi types
- [ ] swaggo annotations; `make swagger`
- [ ] Tests: httptest table (happy path, validation errors, 404)
- [ ] Tests: testcontainers integration (persist + list pagination > 1 page)
- [ ] `make test-api` green

## [frontend] — parallel with [backend] after contract lands

- [ ] Confirm generated client includes new operations (`make gen-contract` if needed)
- [ ] TanStack Query hooks wrapping generated client for create / list / get
- [ ] `/crawls/new` submit form with client-side + server error display
- [ ] `/crawls` paginated list with cursor "load more" / next page
- [ ] `/crawls/[id]` detail page (config + status)
- [ ] Component tests: form validation + success, list render, detail render (mocked client)
- [ ] `make test-web` green

## [integrate] — after both tracks reconcile

- [ ] Merge worktree branches; resolve any root-level conflicts (`openapi.yaml` should be contract-only)
- [ ] `make test` on merged result
- [ ] `code-reviewer` on full diff
- [ ] Standard tier stops here — no `spec-verifier` gate (it runs only at Rigorous)
