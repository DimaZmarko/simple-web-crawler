# Local environment (Colima + docker-compose)

## Prerequisites
```bash
brew install colima docker docker-compose
colima start --cpu 4 --memory 8 --vm-type vz --mount-type virtiofs
```
`vz` + `virtiofs` (Apple Silicon, recent macOS) give the best filesystem and networking behavior. Verify Docker can reach the VM: `docker info`.

## Day-to-day

Start everything with live reload:
```bash
docker compose watch
```
This builds the `dev` targets, starts Postgres → runs migrations → starts api and web, and then syncs source changes into the running containers. The in-container watchers (`air` for Go, `next dev` for the frontend) recompile on the synced files.

Without live sync:
```bash
docker compose up --build
```

Production-shaped build (distroless api, standalone web):
```bash
docker compose -f docker-compose.yml up --build
```

| Service | URL |
| --- | --- |
| web | http://localhost:3000 |
| api | http://localhost:8080 |
| postgres | localhost:5432 (user/pass/db: crawler) |

The published `5432` matches `.mcp.json`, so the Postgres MCP server connects to the same database from the host.

## Colima notes

**File watching.** Host filesystem events don't always propagate into the Lima VM. `docker compose watch` sidesteps this — the Docker CLI on the host watches files and pushes changes in, so the container's own watcher sees normal local writes. Prefer it over bind-mount + in-container watch.

**Two API URLs.** Next.js server code calls the API over the compose network (`API_URL=http://api:8080`); browser code calls it through the host (`NEXT_PUBLIC_API_URL=http://localhost:8080`). Both are set on the `web` service.

**testcontainers-go.** Backend integration tests run on the host, not in compose, and need to find Colima's Docker socket. If they can't:
```bash
export DOCKER_HOST="unix://${HOME}/.colima/default/docker.sock"
export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE="/var/run/docker.sock"
```
The first points the Docker client at Colima; the second tells testcontainers which socket path to bind inside helper containers (Ryuk). Confirm the socket path with `colima status`.

**Architecture.** On Apple Silicon, images build `arm64` by default — fine locally. If you deploy to `amd64`, build with `--platform linux/amd64` (or set it in the build config).
