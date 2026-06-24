---
name: review-swagger-version-drift
description: openapi.yaml info.version drifts ahead of the swaggo-generated docs (docs.go/swagger.yaml) and main.go @version — they are not auto-synced
metadata:
  type: feedback
---

The contract `packages/api-contract/openapi.yaml` carries `info.version` (e.g. bumped to 0.3.1 for crawl intake), but the backend swagger source of truth is the swaggo annotation block in `apps/api/cmd/server/main.go` (`@version`), which feeds `apps/api/docs/docs.go` + `docs/swagger.json` + `docs/swagger.yaml` via `make swagger`. These are two independent version strings.

**Why:** On the crawl-intake branch the contract was 0.3.1 while main.go `@version` and all three generated swagger files were still 0.2.0. `make swagger` regenerates the docs from the annotation, so it never picks up the contract's version — they only match if someone hand-bumps `@version` in main.go to match openapi.yaml. CLAUDE.md golden rule 5 says swagger must be regenerated as part of DoD, but regeneration alone does not reconcile the version number.

**How to apply:** On any review where `openapi.yaml info.version` changed, grep `@version` in `apps/api/cmd/server/main.go` and `version:` in `apps/api/docs/swagger.yaml` / `docs.go`; flag a mismatch as should-fix (cosmetic but a contract-sync smell). Note the path/status content of the swaggo annotations IS the thing to check for correctness — the version string is the recurring drift.
