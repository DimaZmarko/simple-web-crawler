# apps/web — Next.js frontend

App Router, React, TanStack Query. pnpm.

Run: `make -C ../.. dev` (whole stack) or `pnpm dev` when Node is installed locally.
Test: `make -C ../.. test-web`. Lint/typecheck: `make -C ../.. lint-web` / `make -C ../.. typecheck-web`.

All API access goes through the generated client in `src/api/__generated__`, produced from `packages/api-contract/openapi.yaml` via `make -C ../.. gen-contract`. Never hand-write API types.

Follow the `nextjs-conventions` skill.
