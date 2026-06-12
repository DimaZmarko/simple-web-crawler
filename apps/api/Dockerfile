# syntax=docker/dockerfile:1

# Base: shared module cache layer
FROM golang:1.25-alpine AS base
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download

# Dev: hot reload with air. Source is synced in by `docker compose watch`.
FROM base AS dev
RUN go install github.com/air-verse/air@latest
COPY . .
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# Builder: static binary
FROM base AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# Prod: distroless, non-root
FROM gcr.io/distroless/static-debian12:nonroot AS prod
COPY --from=builder /out/server /server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
