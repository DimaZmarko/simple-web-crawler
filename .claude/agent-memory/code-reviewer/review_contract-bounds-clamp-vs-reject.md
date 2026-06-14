---
name: review-contract-bounds-clamp-vs-reject
description: list endpoint clamps limit>max silently while openapi.yaml documents limit maximum:100 + a 400 example saying "between 1 and 100" — behaviour contradicts contract
metadata:
  type: feedback
---

`GET /crawls` handler (`apps/api/internal/api/crawls.go` List) only 400s when `limit` is non-numeric or `< 1`. A `limit` above the contract maximum (100) is passed through and the service `clampLimit` silently caps it to 100 and returns 200. But `openapi.yaml` declares `limit` `maximum: 100` and the 400 response example reads "must be an integer between 1 and 100", implying over-max is a client error.

**Why:** Two reasonable designs (clamp-and-succeed vs reject-with-400) but the contract documents reject and the code does clamp — a consumer that trusts the contract will expect a 400 for limit=500 and instead gets 200 with 100 items. Either is fine; they must agree.

**How to apply:** On any paginated/bounded query param, cross-check the handler's reject path against the contract's documented bounds and error examples. If the code clamps, the contract should describe clamping (drop the over-max 400 example, note "values above N are capped"); if the contract says 400, the handler must reject over-max, not clamp. Flag the mismatch as should-fix, not a blocker, since it's a silent contract-divergence on an edge value.
