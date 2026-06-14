# apps/web — Next.js frontend

App Router, React, TanStack Query. pnpm.

Run: `make -C ../.. dev` (whole stack) or `pnpm dev` when Node is installed locally.
Test: `make -C ../.. test-web`. Lint/typecheck: `make -C ../.. lint-web` / `make -C ../.. typecheck-web`.

All API access goes through the generated client in `src/api/__generated__`, produced from `packages/api-contract/openapi.yaml` via `make -C ../.. gen-contract`. Never hand-write API types.

**Styling: Material UI (MUI) is the standard** — build from MUI components styled with `sx`, responsive by default, theme in `src/theme.ts`, `AppRouterCacheProvider` + `ThemeProvider` + `CssBaseline` wired in `app/layout.tsx`. Do not introduce a second styling system or ship unstyled markup. See the `nextjs-conventions` skill for the setup details.

Follow the `nextjs-conventions` skill.
