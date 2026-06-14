---
name: nextjs-conventions
description: Next.js and React conventions for the frontend in apps/web. Use when writing or modifying frontend code — routes, components, or data fetching. Covers App Router structure, the generated API client, state management, and component tests for this project.
allowed-tools: Read, Edit, Bash, Grep, Glob
---

# Next.js conventions — crawler UI

## Structure
- App Router. Server components by default; add `"use client"` only when a component needs interactivity, state, or browser APIs.
- Routes under `apps/web/src/app`. Shared UI in `apps/web/src/components`. Feature logic colocated with its route.

## API access
- All HTTP goes through the generated client in `apps/web/src/api/__generated__`, produced from `packages/api-contract/openapi.yaml`.
- Never hand-write request or response types. If a type is missing, the contract is incomplete — flag it, do not patch around it.
- Client-side reads use TanStack Query wrapping the generated client (caching, loading, error states). Server components can call the client directly in async functions.
- While the backend is unbuilt, develop against a mock layer that satisfies the generated client interface, so UI work proceeds without a live API.

## State
- Server state via TanStack Query; local UI state via React hooks. Avoid global stores unless genuinely shared.

## Styling — Material UI (MUI) is the project standard
- **MUI is the one component + styling system.** Build UI from MUI components (`Container`, `Stack`, `Grid`, `Card`, `AppBar`, `Table`, `TextField`, `Button`, `Alert`, etc.) and style with the `sx` prop or `styled()`. Do **not** introduce a second system (Tailwind, Bootstrap, CSS Modules, plain global CSS beyond resets) — and do **not** ship bare unstyled `<div>/<form>/<input>` markup.
- **Responsive by default.** Use `Container` for page width, `Grid`/`Stack` for layout, and responsive `sx` values (e.g. `sx={{ p: { xs: 2, md: 4 } }}`) rather than fixed pixel layouts. Use theme spacing/breakpoints, not magic numbers.
- **Theme is centralized.** A single theme lives in `src/theme.ts` (palette, typography, shape). Components consume it via the `sx` prop / `useTheme` — never hardcode colors or fonts.
- **App Router integration (required, easy to get wrong):** the root `app/layout.tsx` must wrap children in `AppRouterCacheProvider` (from `@mui/material-nextjs/v15-appRouter`) → `ThemeProvider` (the shared theme) → `<CssBaseline />`, *outside* the existing `Providers` (TanStack Query) or composed with it. This is what makes SSR styling flush correctly and avoids the flash of unstyled content. Most interactive MUI components need `"use client"`.
- **One-time setup (do this on the first task that needs MUI, since it isn't installed yet):** add deps `@mui/material @emotion/react @emotion/styled @mui/material-nextjs @emotion/cache` (and `@mui/icons-material` if icons are used) to `apps/web/package.json`, create `src/theme.ts`, and wire `app/layout.tsx` as above. Keep the `@mui/*` packages on one consistent major.

## Tests
- Component tests via `make test-web` from the repo root (Vitest; runs in Docker when Node is not on the host).
- Test against the generated client's types so a contract change that breaks the UI fails at test time.
- Keep component tests colocated with components under `src/` (`*.test.tsx`). Vitest only collects `src/`.
- End-to-end tests live in `apps/web/e2e/` (`*.spec.ts`, Playwright) and run against a live stack via `make test-e2e` — they exercise the real browser→Next→API transport, not mocks. Keep them thin smoke tests asserting observable user-facing behavior. The `@playwright/test` version in `package.json` must match the image tag in the `test-e2e` Make target.
