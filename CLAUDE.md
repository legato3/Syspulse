# Syspulse — Project Context

Fork of [rcourtman/Pulse](https://github.com/rcourtman/Pulse) v5.1.28.
Real-time monitoring dashboard for Proxmox, Docker, and Kubernetes.
Repo name: `legato3/Syspulse` on GitHub.

## Dev Environment
- Host: **PROX-WEB** (LXC 129) @ `192.168.0.87`
- Repo: `/opt/syspulse`
- SSH: `root@192.168.0.87` (password: c_2580_C) — or via Proxmox: `pct exec 129 -- <cmd>`
- Go: 1.25 (`/usr/local/go/bin/go`)
- Node.js: 20 LTS
- Air: v1.65 (hot-reload, `/usr/local/bin/air`)

## Stack

| Layer | Tech | Port |
|-------|------|------|
| Backend | Go (`cmd/pulse`) | 7655 |
| Frontend | TypeScript / React + Vite | 5173 (dev) / embedded in prod |
| Go module | `github.com/rcourtman/pulse-go-rewrite` | — |

## Key Directories

```
cmd/
  pulse/           — main server binary
  pulse-agent/     — Proxmox/Docker agent
  pulse-host-agent/— host metrics agent
internal/
  api/             — HTTP + WebSocket handlers; embeds frontend-modern/dist
  monitoring/      — core polling and alert engine
  config/          — config load/save
  ai/              — BYOK AI chat + Patrol
  alerts/          — threshold alerting
  websocket/       — real-time push to browser
pkg/
  proxmox/         — Proxmox VE/PBS/PMG API client
  pbs/             — PBS-specific queries
  auth/            — password hashing, OIDC, SSO
  server/          — HTTP server setup
frontend-modern/
  src/             — React + TypeScript components
  vite.config.ts   — Vite build config (proxies /api → :7655 in dev)
  dist/            — built assets (embedded into Go binary for prod)
```

## Dev Workflow

### Start both servers (recommended)
```bash
bash /opt/syspulse/dev-start.sh
```
- Backend (Air hot-reload): `http://192.168.0.87:7655`
- Frontend (Vite dev): `http://192.168.0.87:5173`

Vite proxies `/api/*` and `/ws` to the Go backend automatically — use port 5173 in the browser.

### Or separately
```bash
# Backend with hot-reload
export PATH=$PATH:/usr/local/go/bin
cd /opt/syspulse && air

# Frontend
cd /opt/syspulse/frontend-modern && npm run dev -- --host 0.0.0.0
```

### Stop
```bash
pkill -f air; pkill -f vite
```

## Build Commands

```bash
export PATH=$PATH:/usr/local/go/bin
cd /opt/syspulse

# Frontend (outputs to frontend-modern/dist — embedded by Go)
cd frontend-modern && npm run build && cd ..

# Backend binary
go build -o /usr/local/bin/syspulse ./cmd/pulse/...

# Both (Makefile)
make build
```

## Tests

```bash
# Go tests
export PATH=$PATH:/usr/local/go/bin
cd /opt/syspulse && go test ./...

# Frontend tests
cd /opt/syspulse/frontend-modern && npm test

# Type-check
npm run type-check
```

## Config & Data
- Config file: `/data/config.json` (created on first run)
- Default data dir: `/data/` (can override with `DATA_DIR` env var)
- Default port: `7655` (override with `PORT` env var)
- Example env: `.env.example`

## Upstream Sync
Upstream: `rcourtman/Pulse` (main branch).
```bash
git remote add upstream https://github.com/rcourtman/Pulse.git
git fetch upstream && git merge upstream/main
```

## Notes
- The Go embed directive in `internal/api/frontend_embed.go` requires `frontend-modern/dist` to exist — always build frontend before building the prod binary.
- Air config is in `.air.toml` at repo root — watches `cmd/`, `internal/`, `pkg/`.
- Husky pre-commit hooks run gitleaks secret scanning — do not commit `.env` or credential files.
