# MaidenMap

Self-hosted reverse-geocode for HAM Maidenhead grid codes. Go + Gin backend, React 19 + Vite frontend, Docker Compose deploy behind an nginx reverse proxy.

## Quick start

```bash
# First-time: fetch GeoNames + Natural Earth datasets into ./data (takes a few minutes)
docker compose --profile update run --rm update-data

# Run the stack
docker compose up -d

# Visit
open http://127.0.0.1:8081/
# API
curl http://127.0.0.1:8080/api/grid/JO65ab
```

## Host nginx

The web and API containers bind to `127.0.0.1:8081` and `127.0.0.1:8080`. Front them with a host nginx that terminates TLS and routes `/api` → 8080, `/` → 8081.

## Docs

- Design spec: [`docs/superpowers/specs/2026-04-20-maidenmap-design.md`](docs/superpowers/specs/2026-04-20-maidenmap-design.md)
- Frontend spec: [`docs/superpowers/specs/2026-04-20-maidenmap-frontend-design.md`](docs/superpowers/specs/2026-04-20-maidenmap-frontend-design.md)
