---
name: review-openapi-fetch-nonok
description: openapi-fetch returns non-2xx response bodies in `error`, not `data` — recurring gap in frontend readiness/health handling
metadata:
  type: feedback
---

In `apps/web`, the `openapi-fetch` client returns `{ data, error }`. For non-2xx responses (e.g. /readyz 503), the parsed body is placed in `error`, and `data` is `undefined`.

**Why:** A handler that comments "503 is delivered as parsed body too" but only reads `data?.status` will never see the degraded body — it lands in the `error` branch. Tests that mock only `{ data: ... }` don't exercise the real 503 transport path.

**How to apply:** When reviewing frontend code that calls the generated client and cares about an error-status body, confirm it reads the body out of `error` (not `data`), and confirm a test mocks the `{ error: ... }` shape, not just `{ data: ... }`.
