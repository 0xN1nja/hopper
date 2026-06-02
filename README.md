# Hopper

A lightweight self-hosted RCON webUI for game servers.

## Quick Start

```yaml
services:
  hopper:
    image: ghcr.io/0xn1nja/hopper:latest
    container_name: hopper
    restart: unless-stopped
    ports:
      - '8080:8080'
    volumes:
      - hopper-data:/data
    environment:
      - USERNAME=admin
      - PASSWORD=changeme

volumes:
  hopper-data:
```

```sh
docker compose up -d
```

Open `http://localhost:8080` and sign in with the credentials above.

> [!NOTE]
> Hopper is designed to run on non-production environments (like homelabs). Do not expose it to the public internet.

## Features

- create multiple RCON sessions
- execute commands
- command history is saved
- RCON connections stay alive when you close the tab

## Development

**Prerequisites:** Go 1.23+, Node.js 20+, Bun

```sh
# Terminal 1 — Go backend (serves API on :8080)
cd backend
DEV=true go run .

# Terminal 2 — SvelteKit dev server (:5173)
bun run dev
```

The frontend dev server proxies API requests to Go on port 8080.

## Building locally

```sh
bun run build                                         # build frontend
cp -r build/. backend/static/files/                  # embed into Go
cd backend && CGO_ENABLED=0 go build -o ../hopper .  # compile binary
./hopper
```
