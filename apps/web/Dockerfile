# syntax=docker/dockerfile:1

# Base: node + pnpm via corepack. Build context is ./apps/web.
FROM node:22-alpine AS base
RUN corepack enable
WORKDIR /app

# Deps: cached install
FROM base AS deps
COPY package.json pnpm-lock.yaml* ./
RUN --mount=type=cache,id=pnpm,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile

# Dev: next dev. Source is synced in by `docker compose watch`.
FROM base AS dev
ENV NODE_ENV=development NEXT_TELEMETRY_DISABLED=1
COPY --from=deps /app/node_modules ./node_modules
COPY . .
EXPOSE 3000
CMD ["pnpm", "dev"]

# Builder: produces .next/standalone (requires output: 'standalone' in next.config)
FROM base AS builder
ENV NEXT_TELEMETRY_DISABLED=1
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN pnpm build

# Prod: minimal standalone runtime, non-root
FROM base AS prod
ENV NODE_ENV=production NEXT_TELEMETRY_DISABLED=1
WORKDIR /app
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
EXPOSE 3000
USER node
CMD ["node", "server.js"]
