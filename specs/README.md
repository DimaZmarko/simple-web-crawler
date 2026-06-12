# Specs

Spec-driven build of the crawler platform. Each spec is a vertical slice: it delivers something observable end-to-end, depends only on earlier specs, and states its API-contract delta so the backend and frontend can be planned and built in parallel within it.

These are **Specify-phase** artifacts. For each one you then run Plan → Tasks → Implement (see the `spec-workflow` skill).

## Build order

| # | Spec | Delivers | Depends on |
| --- | --- | --- | --- |
| 001 | Foundation & health | Walking skeleton: compose boots, `/healthz` + `/readyz`, web shows API status | — |
| 002 | Crawl intake & storage | Submit / list / fetch crawls; they persist as `queued` (no engine yet) — `plan.md` + `tasks.md` ready | 001 |
| 003 | Crawl engine | Dispatcher, worker pool, per-domain rate limiting, dedup, cancellation; `queued → running → completed` | 002 |
| 004 | Live monitoring & results | Live stats on a running crawl; paginated discovered links; dashboard live view | 003 |
| 005 | Lifecycle control | Cancel a crawl; enforce max pages / depth / timeout | 003 |
| 006 | Access control & abuse protection | Auth on writes, rate limit on intake, SSRF guard on seed URLs | 002, 003 |
| 007 | Observability & ops | Metrics endpoint, structured request logs, durable stats flush, graceful-shutdown hardening | 003, 004 |

Phases 1–2 are a strict chain. Within Phase 3, 006 and 007 are independent of each other and can run in parallel once the engine (003) exists.

## How to drive each spec

For a given spec, ask the agent:

1. **Plan** — "Read `specs/003-crawl-engine/spec.md` and produce `plan.md`: the technical approach, the exact `openapi.yaml` delta, concurrency design, and persistence changes." (Plan mode, approve before code.)
2. **Tasks** — "Produce `tasks.md` from the plan, tagging each task `[contract]`, `[backend]`, or `[frontend]`, grouped for parallel execution."
3. **Implement** — "Implement `tasks.md` using parallel-dispatch: contract first, then backend-engineer and frontend-engineer in parallel, then run the gates for this spec's tier." (The `spec-workflow` skill's *Implement* phase has the gates-by-tier table — Standard stops at `test-runner` + `code-reviewer`; Rigorous adds the `plan-verifier` gate before code and `spec-verifier` after. Don't hardcode the gate set here.)

Keep the spec/plan/tasks files updated if scope shifts mid-build — they are the durable record, not the chat.

## Spec template

Each `spec.md` header carries: Status · Verification tier · Depends on. The body uses: Goal · User scenarios · API surface (contract delta, prose) · Behavior & rules · Data · Concurrency notes · Frontend scope · Out of scope · Acceptance criteria · Open questions.

## Verification tiers
Each `spec.md` header sets a tier. The routing is documented once, canonically, in **`.claude/README.md`** (fast path vs Standard/Rigorous decision tree) and the **`spec-workflow`** skill (full risk rubric, auto-selection, and override rules). The deterministic floor — `go test -race`, contract match, lint — runs regardless of tier. In brief:

- **Standard** — floor + `code-reviewer` after implementation.
- **Rigorous** — adds the `plan-verifier` gate before code and `spec-verifier` after.

Current tiers: 001 → Standard, 002 → Standard, 003 → Rigorous, 004 → Standard, 005 → Rigorous, 006 → Rigorous, 007 → Rigorous.
