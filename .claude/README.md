# `.claude` — how work is routed

This directory holds the agents, skills, and conventions that drive development in this repo. The first decision on any change is **how much ceremony it needs**: the fast path (no spec artifacts), the Standard tier, or the Rigorous tier. The deterministic floor — `make check` (or the affected `make test-*`), contract sync, and `-race` — runs at **every** level; the routing below only governs how much spec ceremony and LLM verification sits on top.

## Decision tree

```mermaid
flowchart TD
    A[Incoming change] --> B{Fast-path eligible?<br/>single app OR contract-only typo/docs ·<br/>no contract shape/status/endpoint change ·<br/>≤3 source files ·<br/>no migration / new goroutine / shared state / auth ·<br/>bug fix / refactor / test / lint — not a new feature}
    B -- "No (any criterion fails)" --> C[Run spec-workflow: Specify → Plan → Tasks → Implement]
    B -- "Yes (all hold)" --> F[FAST PATH<br/>engineer agent or inline fix ·<br/>test-runner on affected suite ·<br/>code-reviewer optional, by judgment ·<br/>NO spec.md/plan.md/tasks.md ·<br/>NO plan-verifier / spec-verifier]
    F -. "discover contract delta,<br/>migration, or both apps" .-> C

    C --> D{Risk score ≥ 4,<br/>OR concurrency core / auth-SSRF /<br/>irreversible migration at scale /<br/>completion-correctness needs judgment?}
    D -- No --> S[STANDARD<br/>spec.md → plan.md → tasks.md ·<br/>after implement: test-runner + code-reviewer]
    D -- Yes --> R[RIGOROUS<br/>spec.md → plan.md → tasks.md ·<br/>before code: plan-verifier gate ·<br/>after implement: test-runner + code-reviewer + spec-verifier]
```

In prose:

1. **Is the change fast-path eligible?** All of: single app (or contract-only typo/docs); no contract shape/status/endpoint change; ≤ 3 source files; no migration, new shared state, or auth surface; and the intent is a fix/refactor/cleanup — not a new feature slice. (The `spec-workflow` skill holds the exact thresholds — this is the summary.)
   - **Yes** → fast path: engineer agent (or inline fix) → `test-runner` on the affected suite → `code-reviewer` only if it touches handlers, concurrency, or API-client usage. No spec artifacts, no verifier subagents. If mid-work you hit a contract delta, migration, or both apps, **stop and escalate** to the full workflow.
   - **No** → run the `spec-workflow` skill (Specify → Plan → Tasks → Implement) and pick a tier.

2. **Standard vs Rigorous** (computed at Specify from the risk rubric, re-checked at Plan):
   - Score **0–3** → **Standard**: the three spec artifacts; after implement, `test-runner` + `code-reviewer`.
   - Score **≥ 4**, or any of {concurrency core, auth/SSRF, irreversible migration at scale, completion-correctness needing judgment} → **Rigorous**: same artifacts; `plan-verifier` gate before code; `test-runner` + `code-reviewer` + `spec-verifier` after implement.

A manual tier set without the `(auto)` marker always wins, and tiers only ever ratchet **up** at Plan, never down. See `.claude/skills/spec-workflow/SKILL.md` for the full rubric and procedures.
