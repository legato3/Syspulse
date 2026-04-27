# Running Pulse Locally

This guide runs Pulse from the checked-out source tree for local development or
manual testing.

## Prerequisites

- Go 1.25 or newer
- Node.js and npm
- A free local port `7655`

## Run From Source

From the repository root:

```bash
npm --prefix frontend-modern ci
make build
mkdir -p tmp/pulse-data
BIND_ADDRESS=127.0.0.1 \
PULSE_DATA_DIR="$PWD/tmp/pulse-data" \
./pulse
```

Open the app at:

```text
http://127.0.0.1:7655
```

On first launch, Pulse requires a bootstrap token. In another terminal:

```bash
cat tmp/pulse-data/.bootstrap_token
```

Paste that token into the setup screen and create the admin account.

## Why Set `PULSE_DATA_DIR`?

By default, Pulse stores persistent configuration in `/etc/pulse`, which is
appropriate for a service install but inconvenient for local development. Setting
`PULSE_DATA_DIR="$PWD/tmp/pulse-data"` keeps generated config, the encryption key,
the bootstrap token, and SQLite databases inside the workspace.

## Common Environment Variables

```bash
BIND_ADDRESS=127.0.0.1       # Listen only on localhost
FRONTEND_PORT=7655           # UI/API port
PULSE_DATA_DIR=tmp/pulse-data
PULSE_MOCK_MODE=true         # Optional demo/mock data mode
```

If you change `FRONTEND_PORT`, open that port in the browser instead of `7655`.

## Hot Reload Development

The repo also includes a hot-reload script for day-to-day development:

```bash
npm --prefix frontend-modern ci
PULSE_DATA_DIR="$PWD/tmp/pulse-data" ./scripts/hot-dev.sh
```

The hot-reload script starts a Vite frontend dev server and a Go backend. By
default, the Vite UI is served on:

```text
http://127.0.0.1:5173
```

The backend API remains on:

```text
http://127.0.0.1:7655
```

## Docker Alternative

For a quick non-development run:

```bash
docker compose up -d
```

Then open:

```text
http://127.0.0.1:7655
```

Read the Docker bootstrap token with:

```bash
docker exec pulse cat /data/.bootstrap_token
```

## Troubleshooting

If port `7655` is already in use, either stop the other process or use another
port:

```bash
FRONTEND_PORT=8080 \
BIND_ADDRESS=127.0.0.1 \
PULSE_DATA_DIR="$PWD/tmp/pulse-data" \
./pulse
```

If Go tries to use a cache directory you cannot write to, set local cache paths:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod make build
```

If npm fails because it cannot write to the default cache or log directory:

```bash
npm --prefix frontend-modern ci --cache /tmp/npm-cache --logs-dir /tmp/npm-logs
```
