# Simple Web Crawler — developer Makefile
#
# Backend (Go) runs on the host; the web toolchain (pnpm/Vitest/ESLint/codegen)
# runs in a throwaway node:22-alpine container, so Node/pnpm are not required
# on the host. Postgres + migrations + both apps run via docker compose.

# Auto-detect the compose flavor: `docker compose` plugin or standalone `docker-compose`.
COMPOSE := $(shell docker compose version >/dev/null 2>&1 && echo "docker compose" || echo "docker-compose")

# Run a command inside a disposable Node 22 container, bind-mounting apps/web.
define web_exec
docker run --rm -v "$(CURDIR)/apps/web:/app" -w /app \
	-e NODE_OPTIONS=--max-old-space-size=2048 node:22-alpine \
	sh -c 'corepack enable && $(1)'
endef

.DEFAULT_GOAL := help
.PHONY: help \
        dev watch up down logs ps build clean \
        check test test-api test-web test-e2e \
        lint lint-api lint-web vet typecheck-web fmt fmt-api fmt-web \
        gen-contract swagger migrate migrate-down

# Playwright runs in the official image, which bakes in browsers for exactly
# this version. It MUST match `@playwright/test` in apps/web/package.json.
PLAYWRIGHT_VERSION := v1.51.1
# Where the e2e suite points the browser. The app is published on the host by
# `make up`, so from inside the container we reach it via host.docker.internal
# (works on Docker Desktop and Colima). Override for CI or a remote target.
E2E_BASE_URL ?= http://host.docker.internal:3000

help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

# ----------------------------------------------------------------------------
# Local environment
# ----------------------------------------------------------------------------
dev: watch ## Run the whole stack with live reload (alias for `watch`)

watch: ## Start db + api + web with live reload (recommended on Colima)
	$(COMPOSE) watch

up: ## Build and start the full stack, detached
	$(COMPOSE) up -d --build

down: ## Stop the stack and remove containers (keeps the db volume)
	$(COMPOSE) down

logs: ## Follow logs from all services
	$(COMPOSE) logs -f

ps: ## Show service status
	$(COMPOSE) ps

build: ## Build the production-shaped images (distroless api, standalone web)
	$(COMPOSE) -f docker-compose.yml build

clean: ## Stop the stack and remove volumes — DESTROYS the local database
	$(COMPOSE) down -v

# ----------------------------------------------------------------------------
# Quality gates
# ----------------------------------------------------------------------------
check: lint typecheck-web test ## Run all linters, type checks, and tests

test: test-api test-web ## Run backend and frontend test suites

test-api: ## Backend tests with the race detector (required)
	cd apps/api && go test -race ./...

test-web: ## Frontend component tests (Vitest, in Docker)
	$(call web_exec,pnpm install --frozen-lockfile --child-concurrency=1 && pnpm test)

test-e2e: ## Playwright e2e smoke tests (requires the stack running: make up)
	docker run --rm --add-host=host.docker.internal:host-gateway \
		-v "$(CURDIR)/apps/web:/app" -w /app \
		-e CI=1 -e E2E_BASE_URL=$(E2E_BASE_URL) \
		mcr.microsoft.com/playwright:$(PLAYWRIGHT_VERSION)-noble \
		sh -c 'corepack enable && pnpm install --frozen-lockfile --child-concurrency=1 && pnpm exec playwright test'

lint: lint-api lint-web ## Lint backend and frontend

lint-api: vet ## go vet + gofmt check (fails if any file needs formatting)
	@unformatted=$$(cd apps/api && gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "These Go files need gofmt:"; echo "$$unformatted"; exit 1; \
	fi

vet: ## go vet the backend
	cd apps/api && go vet ./...

lint-web: ## ESLint the frontend (next lint, in Docker)
	$(call web_exec,pnpm install --frozen-lockfile --child-concurrency=1 && pnpm lint)

typecheck-web: ## TypeScript type check the frontend (tsc --noEmit, in Docker)
	$(call web_exec,pnpm install --frozen-lockfile --child-concurrency=1 && pnpm exec tsc --noEmit)

fmt: fmt-api fmt-web ## Format backend and frontend in place

fmt-api: ## gofmt -w the Go backend
	cd apps/api && gofmt -w .

fmt-web: ## Prettier --write the frontend (in Docker)
	$(call web_exec,pnpm install --frozen-lockfile --child-concurrency=1 && pnpm exec prettier --write .)

# ----------------------------------------------------------------------------
# Codegen & database
# ----------------------------------------------------------------------------
# Codegen tool versions are pinned so regeneration is reproducible and the CI
# drift gate (git diff --exit-code on generated files) can't fail on a tool bump.
# Keep these in sync with .github/workflows/ci.yml.
OAPI_CODEGEN_VERSION := v2.7.1
OPENAPI_TS_VERSION   := 7.13.0
SWAG_VERSION         := v1.16.6

gen-contract: ## Regenerate Go types + TS client from packages/api-contract/openapi.yaml
	cd apps/api/internal/api/types && go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION) \
		-config oapi-codegen.yaml ../../../../../packages/api-contract/openapi.yaml
	docker run --rm -v "$(CURDIR):/work" -w /work node:22-alpine \
		npx -y openapi-typescript@$(OPENAPI_TS_VERSION) packages/api-contract/openapi.yaml -o apps/web/src/api/__generated__/schema.d.ts

swagger: ## Regenerate swagger docs from swaggo annotations (apps/api/docs)
	cd apps/api && go run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init -g cmd/server/main.go -o ./docs

migrate: ## Apply DB migrations (reuses the compose `migrate` service)
	$(COMPOSE) run --rm migrate

migrate-down: ## Roll back the last DB migration
	$(COMPOSE) run --rm migrate \
		-path=/migrations -database=postgres://crawler:crawler@db:5432/crawler?sslmode=disable down 1
