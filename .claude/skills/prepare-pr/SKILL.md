---
name: prepare-pr
description: Draft a pull-request summary for the current branch — what the change does and which areas of the monorepo it touches. Use when the user is ready to open a PR, asks for a PR description/summary, or wants to review what a branch changed before pushing. Produces a paste-ready PR body mapped to this repo's structure (api / web / contract / migrations / specs / tooling) and a CI-aligned checklist.
---

# Prepare PR

Generates a structured pull-request summary from the current branch's diff against its base. It **drafts and reports** — it does not push or open the PR unless the user explicitly asks (pushing is gated to `ask` in settings, and opening a PR is outward-facing).

## When to use

The user is wrapping up a branch and wants a PR description, a summary of what was done, or an overview of which areas were touched. Not for fast-path commits that aren't becoming a PR.

## Inputs to gather first

Run these read-only commands (all pre-approved) to ground the summary in the real diff — never invent changes:

```bash
BASE=$(git merge-base HEAD origin/main 2>/dev/null || git merge-base HEAD main)
git log --oneline "$BASE"..HEAD          # commits on this branch
git diff --stat "$BASE"..HEAD            # files + churn
git diff --name-status "$BASE"..HEAD     # add/modify/delete per file
```

If a `specs/<feature>/spec.md` exists for this branch, read it: the spec's task statement and acceptance criteria are the most authoritative source for the "what" and "why" — summarize from it rather than re-deriving intent from the diff. Mention the spec path in the PR body.

## Map changed files to areas

Bucket every changed path so the "Areas touched" section reflects this monorepo's structure. Use this mapping:

| Path glob | Area label |
| --- | --- |
| `packages/api-contract/openapi.yaml` | **API contract** (coordination surface — call out first) |
| `apps/api/internal/api/types/types.gen.go`, `apps/web/src/api/__generated__/**` | **Generated client/types** (note: from codegen, not hand-edited) |
| `apps/api/db/migrations/**` | **DB migration** (flag reversibility + that it's schema-changing) |
| `apps/api/docs/**` | **Swagger docs** (regenerated from swaggo annotations) |
| `apps/api/**` (other) | **Backend** (Go crawler — handlers/dispatcher/worker pool/queries) |
| `apps/web/**` (other) | **Frontend** (Next.js App Router / components / data fetching) |
| `specs/**` | **Spec artifacts** |
| `.claude/**` | **Tooling** (agents/skills/hooks) |
| `Makefile`, `docker-compose*.yml`, `.github/**`, `.env.example`, root configs | **Infra/CI** |

Lead with the contract and migrations if present — they are the highest-signal, cross-cutting changes a reviewer needs to see first.

## Output: the PR body

Emit this as a fenced markdown block the user can paste, **and** write it to `/tmp/pr-body-<branch>.md` so it can be fed to a PR tool later. Keep it tight — a reviewer should grasp the change in 30 seconds.

```markdown
## Summary
<1–3 sentences: what this PR does and why. If a spec exists, state the task it implements.>

## Areas touched
- **<Area>** — <what changed there, one line>
- ...

## Contract / data changes
<Only if the contract, a migration, or generated artifacts changed. State the delta, whether it's additive or breaking, and that clients/types/swagger were regenerated. Omit this section entirely if none apply.>

## Verification
- [ ] `make check` (lint + typecheck + tests)
- [ ] `make test-api` green with `-race`            <!-- keep if backend touched -->
- [ ] Contract regenerated, artifacts committed (`make gen-contract`) <!-- keep if contract/generated touched -->
- [ ] Swagger regenerated, committed (`make swagger`)  <!-- keep if backend/docs touched -->
- [ ] `make test-web` green                          <!-- keep if frontend touched -->
<Report which of these you actually ran and their result. Don't tick a box you didn't verify — say "not run" plainly.>

## Out of scope / follow-ups
<Anything deliberately deferred. Omit if none.>
```

Tailor the Verification checklist to the areas actually touched (drop irrelevant lines) — it mirrors the gates in `.github/workflows/ci.yml`, so a green local run predicts a green CI.

## Honesty rules

- Summarize only what the diff and commits show. If intent is unclear and there's no spec, ask the user rather than guessing the "why".
- For the Verification section, report the true state. If tests weren't run in this session, say so and offer to run them (delegate to the `test-runner` agent) before the user opens the PR.
- Don't claim the contract/swagger are in sync unless a regen + `git diff --exit-code` confirms it — that exact drift check is what CI runs and what will fail the PR otherwise.

## Opening the PR (only if asked)

This host has no `gh` and `git push` is gated to `ask`. So by default, stop after producing the body. If the user wants to go further:
1. Confirm the target base (default `main`) and that tests are green.
2. Push the branch (`git push -u origin HEAD`) — this will prompt for approval.
3. If `gh` is available: `gh pr create --base main --title "<title>" --body-file /tmp/pr-body-<branch>.md`. Otherwise give the user the body and the compare URL to open it in the browser.
