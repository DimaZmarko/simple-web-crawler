---
name: api-contract
description: How the shared API contract works in packages/api-contract. Use when adding or changing an endpoint, request/response shape, or status code, or when regenerating clients. Preloaded into the engineer agents so both sides treat the contract as authoritative.
allowed-tools: Read, Edit, Bash, Grep, Glob
---

# API contract

`packages/api-contract/openapi.yaml` is the single source of truth for the HTTP API. Everything else derives from it.

## Rules
- The contract changes before the code. Backend implements to it; frontend consumes a client generated from it.
- OpenAPI 3.1. Shared schemas go in `components/schemas` and are referenced with `$ref`. Every operation has an `operationId` (used as the generated client method name) and at least one example.
- Generated artifacts are never hand-edited:
  - TS client → `apps/web/src/api/__generated__`
  - Go types → `apps/api/internal/api/types` (or as configured)
- Regenerate both with `make gen-contract` after any contract edit.

## Versioning
- Additive change (new optional field, new endpoint): minor version bump.
- Breaking change (removed/renamed field, changed type, removed/renamed operation, narrowed status): major version bump, called out explicitly so both tracks adapt.

## Backend ↔ contract sync
The backend uses swaggo annotations on handlers. `make swagger` regenerates a spec from the code; it must match `openapi.yaml`. If they diverge, the contract wins — fix the annotations, not the contract, unless the contract itself was wrong (in which case go through api-contract-keeper).

## Why this exists
The contract is the only thing backend and frontend share. Keeping it authoritative and current is what lets the two tracks be built in parallel by separate agents without coordinating.
