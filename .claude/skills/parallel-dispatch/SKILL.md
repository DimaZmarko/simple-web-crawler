---
name: parallel-dispatch
description: Orchestration playbook for running backend and frontend work concurrently. Use when a feature spans both apps/api and apps/web and the two parts can proceed independently once the API contract is fixed. Coordinates contract-first, then parallel engineer agents in isolated worktrees, then reconciliation.
---

# Parallel dispatch

This skill runs in the main session (the orchestrator). Subagents cannot spawn subagents, so all delegation happens from here.

## Preconditions
A `tasks.md` exists (from spec-workflow) with `[contract]`, `[backend]`, and `[frontend]` tasks, and the backend/frontend tasks have no ordering dependency on each other beyond the contract.

## Procedure

1. **Lock the contract first.** Invoke `api-contract-keeper` with the contract delta. Wait for it to finish and regenerate the client and Go types. Nothing parallel starts until the contract is landed on the integration branch (committed or explicitly staged for merge) — it is the only shared surface, so it must be stable before the tracks diverge.

2. **Dispatch both tracks in parallel.** Spawn `backend-engineer` and `frontend-engineer` at the same time, each in the background. Because both have `isolation: worktree`, each gets its own checkout branched from the default branch and they cannot collide on files. Give each agent only its own task list and the relevant contract operations.
   - Phrase it as parallel explicitly, e.g. "Run the backend tasks and the frontend tasks in parallel using the backend-engineer and frontend-engineer subagents in separate worktrees."

3. **Let them run.** Keep the main context free while they work. Do not babble between them; they do not need to coordinate because they share only the contract.

4. **Gather.** When both report back, invoke `test-runner` (`make test` from repo root) against each worktree branch, then `code-reviewer` on each diff. For the Rigorous tier, also invoke `spec-verifier` with the `spec.md` path and branch diff.

5. **Reconcile.** Merge the two worktree branches. Because backend changes are confined to `apps/api` and frontend to `apps/web`, merges are conflict-free except in shared root config — resolve those by hand. Run the full `make test` once on the merged result.

## When NOT to use
- The frontend depends on a backend behavior that isn't expressible in the contract (rare) — sequence them instead.
- The change touches only one app — invoke that single engineer agent directly, no orchestration needed.

## Escalation
For features too large for one session, or where the tracks need sustained back-and-forth, consider agent teams (`CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`) so each track runs in its own persistent session. Heavier and more expensive — use only when worktree-isolated subagents aren't enough.
