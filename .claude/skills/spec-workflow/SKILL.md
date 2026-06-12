---
name: spec-workflow
description: Structured feature development using the Specify, Plan, Tasks, Implement loop. Use when the user asks to build or add a feature, a new endpoint plus its UI, or any non-trivial change spanning more than one file. Produces durable spec artifacts under specs/ before code is written.
---

# Spec-driven feature workflow

## Fast path — skip the spec artifacts

Before reaching for the four phases, check whether the change qualifies for the **fast path**. The fast path produces **no** `spec.md`, `plan.md`, or `tasks.md` and runs **no** `plan-verifier` / `spec-verifier`. The deterministic floor still applies — `make check` (or the affected `make test-*`), contract sync, and `-race` are never skipped.

Use the fast path only when **all** of these hold:

| Criterion | Threshold |
|-----------|-----------|
| Scope | Single app only (`apps/api` **or** `apps/web`), OR a contract-only typo/docs fix with no behavior change |
| Contract | No change to `openapi.yaml` request/response shapes, status codes, or new endpoints |
| Files | ≤ 3 source files changed (exclude generated artifacts, lockfiles, formatting-only) |
| Risk | No migrations, no new goroutines / shared mutable state, no auth or security surface |
| Intent | Bug fix, refactor, test fix, lint/format, or small internal cleanup — **not** a new feature slice |

**Procedure:**
1. Invoke the relevant engineer agent directly (`backend-engineer` or `frontend-engineer`), or fix inline if trivial.
2. Run `test-runner` for the affected suite (`make test-api`, `make test-web`, or `make test`).
3. Optionally run `code-reviewer` if the change touches handlers, concurrency, or API-client usage — use judgment, not a mandate.
4. Do **not** create or update `specs/*/spec.md`, `plan.md`, or `tasks.md`.
5. Do **not** invoke `plan-verifier` or `spec-verifier`.

**Escalation:** if mid-work you discover the change needs a contract delta, a migration, or spans both apps, **stop** and switch to the Standard or Rigorous track — say so explicitly and start at Specify.

If any fast-path criterion fails, run the full workflow below.

## The four phases

Drive a feature through four phases. Each phase produces an artifact; do not skip ahead.

## 1. Specify
Write `specs/<feature>/spec.md`: the user-facing behavior, the API surface it needs (in prose, not yet OpenAPI), acceptance criteria, and out-of-scope notes. No implementation detail. Set the `Verification tier` (see below) in the spec header — record it, don't improvise it later. Confirm the spec with the user before planning.

## 2. Plan
Write `specs/<feature>/plan.md`: the technical approach. Identify which side(s) change — backend, frontend, or both — and the exact contract delta. Call out concurrency implications on the backend (new goroutines, shared state, cancellation) and data-fetching implications on the frontend. Re-confirm the verification tier here; the plan now exists, so if it revealed more risk than the spec assumed, raise the tier and say why.

For the **Rigorous** tier, run the **plan-verifier** gate before writing any code: a fresh-context subagent checks that the plan covers every acceptance criterion and that the contract delta is complete. On fail, revise the plan and re-run (cap at 2 rounds, then escalate).

## 3. Tasks
Write `specs/<feature>/tasks.md`: an ordered checklist, each task tagged `[backend]`, `[frontend]`, or `[contract]`. Group tasks so that backend-only and frontend-only work can run in parallel. The contract task always comes first.

## 4. Implement
If the feature is contract + backend + frontend, hand off to the `parallel-dispatch` skill. If it touches only one side, invoke that engineer agent directly. Then run the gates **for the spec's tier** — do not improvise the set; both tiers are listed here so the right gates always fire:

| Gate | When | Standard | Rigorous |
| --- | --- | --- | --- |
| `plan-verifier` | before code (phase 2) | — | ✅ |
| deterministic floor — `make check` / affected `make test-*`, contract match, `-race` | after implement | ✅ | ✅ |
| `test-runner` (affected suite) | after implement | ✅ | ✅ |
| `code-reviewer` (on the diff) | after implement | ✅ | ✅ |
| `spec-verifier` (diff vs acceptance criteria) | after implement | — | ✅ |

On any unmet acceptance criterion from `spec-verifier`, the relevant engineer fixes and it re-verifies (cap 2–3 rounds, then escalate). The tier comes from the spec header — if you can't find it, you skipped Specify; go back. See **Verification tiers** below for what sets the tier.

## Verification tiers
Once a change is on the full workflow (not the fast path), it runs at one of **two** tiers. Rigor scales to risk. The deterministic floor — `make test-api` / `make test-web` (or `make check`), the contract match, lint, `-race` — runs at every tier; the tier only governs how much LLM verification sits on top.

### Selecting the tier (automatic)
At Specify, compute the tier from the spec's risk properties rather than asking. Score the spec:

- +2 — touches persisted data or migrations (hard to reverse)
- +2 — concurrency or shared mutable state
- +2 — security surface: auth, or untrusted external input
- +2 — acceptance criteria need judgment (not fully test-checkable)
- +1 — spans backend and frontend
- +1 — more than 5 acceptance criteria

Map the total to a tier: **0–3 → Standard · 4+ → Rigorous.** Regardless of score, choose **Rigorous** if the feature touches any of: the concurrency core, auth/SSRF, irreversible migrations at scale, or completion-correctness that needs human judgment.

Then record the result in the spec header as `Verification tier: Standard (auto — …)` or `Verification tier: Rigorous (auto — …)` with the one-line reason (which factors fired). Two overrides:
- An author-set tier with no `(auto)` marker always wins — never lower it automatically.
- Re-run the computation at Plan once the technical shape is known; if planning surfaced risk the spec missed (new shared state, a new auth path), the score rises and the tier with it. Only ratchet up, never down, without the author saying so.

### What each tier runs
- **Standard** (default for feature work; score 0–3). Artifacts: `spec.md` → `plan.md` → `tasks.md`. After implement: `test-runner` + `code-reviewer`. No dedicated verifier subagents.
- **Rigorous** (score ≥ 4, or any of the always-Rigorous triggers above). Same three artifacts. **Before code:** `plan-verifier` gate in phase 2. **After implement:** `test-runner` + `code-reviewer` + `spec-verifier` (fresh context, read-only — checks the diff against the spec's acceptance criteria; on unmet criteria the relevant engineer fixes and it re-verifies, cap 2–3 rounds).

There is no parallel verifier committee — Rigorous is exactly plan-verifier + code-reviewer + spec-verifier.

Keep the spec artifacts updated if scope changes mid-build — they are the durable record, the chat is not.
