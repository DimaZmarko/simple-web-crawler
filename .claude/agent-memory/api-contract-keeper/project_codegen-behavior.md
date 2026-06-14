---
name: codegen-behavior
description: How make gen-contract's two generators react to different kinds of openapi.yaml edits (Go types vs TS client)
metadata:
  type: project
---

`make gen-contract` runs two generators with different sensitivities to contract edits.

**Go types** (`apps/api/internal/api/types/types.gen.go`, via oapi-codegen) emits named types ONLY from `components/schemas`. Adding/removing a `$ref` to an *existing* schema inside an operation's responses/parameters produces NO diff in the Go file. A no-diff `types.gen.go` after such a change is correct, not a failed regen — verify by grepping that the referenced schema type still exists rather than assuming codegen broke.

**TS client** (`apps/web/src/api/__generated__/schema.d.ts`, via openapi-typescript) DOES reflect per-operation response/parameter changes, since it types each operation's full response map. Expect a diff here whenever you add a status code or param to an operation.

oapi-codegen emits a benign OpenAPI 3.1-not-fully-supported WARNING on every run — the repo's spec has always been 3.1 and Go generation works regardless. Not an error.

**Why:** Confirmed 2026-06-14 adding a `400` ValidationError ref to `listCrawls` (v0.3.0 -> 0.3.1): TS gained the 400 branch, Go was unchanged because ValidationError already existed.
**How to apply:** When a contract edit only references existing schemas, expect TS-only diffs; don't chase a "missing" Go diff. When introducing a brand-new schema under components/schemas, expect both files to change.
