# MaxwellGoSpine

[![CI](https://github.com/Hex-Zero/MaxwellGoSpine/actions/workflows/ci.yml/badge.svg)](https://github.com/Hex-Zero/MaxwellGoSpine/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hex-zero/MaxwellGoSpine)](https://goreportcard.com/report/github.com/hex-zero/MaxwellGoSpine)
<!-- TODO: Add coverage badge (Codecov or Coveralls) after uploading reports there -->

Production-ready Go 1.22 REST API spine (net/http + chi) with layered architecture, PostgreSQL, structured logging, metrics, health/readiness, graceful shutdown.

## Quick Start

```bash
cp .env.example .env # edit values
docker compose up -d db
go mod tidy
go run ./cmd/server

# or with docker-compose (app + db)
docker compose up --build
```

## Make Targets

```bash
make run    # run dev server
make test   # run tests
make lint   # golangci-lint
make build  # build binary
make migrate-up   # apply migrations (requires migrate CLI)
make auto-commit  # run lint/test/build then auto commit & push if changes
```

## Environment Variables

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| APP_NAME | no | maxwell-api | App name |
| ENV | no | dev | dev or prod |
| HTTP_PORT | no | 8080 | HTTP listen port |
| DB_DSN | yes | - | Postgres DSN |
| READ_TIMEOUT | no | 10s | Read timeout |
| WRITE_TIMEOUT | no | 15s | Write timeout |
| CORS_ORIGINS | no | (empty) | Comma list of allowed origins |
| LOG_LEVEL | no | info | zap log level |
| PPROF_ENABLED | no | 0 | Enable /debug/pprof when 1 |
| CACHE_MAX_COST | no | 10000 | Ristretto max cost (approx entries) |
| CACHE_NUM_COUNTERS | no | 100000 | Ristretto counters (10x max items) |
| CACHE_BUFFER_ITEMS | no | 64 | Ristretto buffer items |
| REDIS_ADDR | no | (empty) | Redis host:port to enable shared cache |
| REDIS_PASSWORD | no | (empty) | Redis password |
| REDIS_DB | no | 0 | Redis DB number |

## API Examples

```bash
curl -X POST http://localhost:8080/v1/users -d '{"name":"Alice","email":"alice@example.com"}' -H 'Content-Type: application/json'
curl http://localhost:8080/v1/users?page=1&page_size=10
curl http://localhost:8080/v1/users/{uuid}
curl -X PATCH http://localhost:8080/v1/users/{uuid} -d '{"name":"New"}' -H 'Content-Type: application/json'
curl -X DELETE http://localhost:8080/v1/users/{uuid}
curl http://localhost:8080/metrics
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

## Tests & Lint

```bash
make test
make lint
```

## Migrations

Uses plain SQL compatible with golang-migrate.

```bash
migrate -path migrations -database $DB_DSN up
```

## OpenAPI

Minimal `openapi.yaml` included; expand as needed or generate from annotations.

## Seed Data

Example seed script at `scripts/seed.sql`:

```bash
psql "$DB_DSN" -f scripts/seed.sql
```

## Notes

* Layers: cmd/server, internal/* (config, log, middleware, http handlers/render, core domain/services, storage).
* Replace module path in `go.mod` with your repository path if different.
* No global mutable singletons; dependencies passed via constructors.
* Enhancements: soft deletes, email normalization, merge-patch updates.
* New make target `auto-commit` to run checks then commit & push changes.
* Caching: layered Ristretto (in-process) + optional Redis; ETag middleware for GET responses.
