# Multi-tenant Event Pipeline

Bootstrap repository for a Go backend that ingests tenant-scoped event batches, processes them asynchronously, and serves analytics from rollups.

## Current state

This repository currently contains the bootstrap scaffold from Issue `#2`, the local infrastructure stack from Issue `#3`, and the initial tenancy/auth schema work from Issue `#6`:

- Go module initialization
- command entrypoints for `api`, `worker`, and `loadgen`
- internal package layout for planned subsystems
- local Docker Compose stack for API, worker, Postgres, and Redis
- migration runner plus initial `tenants`, `users`, `memberships`, `api_keys`, and `sources` schema
- bootstrap developer commands in `Makefile`

The ingestion pipeline, seed data, readiness checks, and business logic are tracked in later tickets and are not implemented yet.

## Repository layout

```text
cmd/                runnable binaries
internal/           application packages by subsystem
migrations/         database migration files
deploy/compose/     local Docker Compose stack
scripts/            developer helper scripts
docs/               scope, delivery plan, and build status
```

## Developer commands

```bash
make up
make down
make run
make run-api
make run-worker
make run-loadgen
make migrate
make seed
make test
```

## Bootstrap usage

Start the local stack:

```bash
make up
```

The compose stack starts:

- `api` on `http://localhost:8080`
- `postgres` on `localhost:5432`
- `redis` on `localhost:6379`

Compose uses named volumes for Postgres and Redis state and injects explicit local-development environment variables into the `api` and `worker` containers:

```text
APP_ENV=development
HTTP_ADDR=:8080                # api only
DATABASE_URL=postgres://postgres:postgres@postgres:5432/event_pipeline?sslmode=disable
REDIS_URL=redis://redis:6379
JWT_SIGNING_KEY=development-signing-key
INGEST_MAX_BODY_BYTES=1048576
INGEST_MAX_BATCH_EVENTS=1000
```

The compose file also configures health checks for Postgres and Redis so the `api` and `worker` containers wait for those dependencies before starting.

Run the API directly:

```bash
APP_ENV=development \
DATABASE_URL=postgres://postgres:postgres@localhost:5432/event_pipeline?sslmode=disable \
REDIS_URL=redis://localhost:6379 \
JWT_SIGNING_KEY=development-signing-key \
make run-api
```

Run the worker directly:

```bash
APP_ENV=development \
DATABASE_URL=postgres://postgres:postgres@localhost:5432/event_pipeline?sslmode=disable \
REDIS_URL=redis://localhost:6379 \
JWT_SIGNING_KEY=development-signing-key \
make run-worker
```

The bootstrap API currently exposes a minimal liveness endpoint at `http://localhost:8080/livez`.

## Configuration

The bootstrap binaries now load and validate configuration from process environment variables at startup. There is no automatic `.env` loading; use shell exports, inline environment variables, or Compose to provide settings.

Required environment variables:

- `DATABASE_URL`
- `REDIS_URL` (preferred) or `REDIS_ADDR` as a temporary backward-compatible fallback
- `JWT_SIGNING_KEY` with at least 16 characters

Development defaults:

- `APP_ENV=development` when unset
- `HTTP_ADDR=:8080` for the API only when `APP_ENV=development`
- `INGEST_MAX_BODY_BYTES=1048576`
- `INGEST_MAX_BATCH_EVENTS=1000`

Invalid or missing configuration causes the binary to exit before serving traffic.

## Logging and request correlation

The bootstrap API and worker now emit structured JSON logs via Go's standard-library `slog` package. Each log line includes a `service` field so API and worker output can be separated easily in Compose or aggregated logs.

The API wraps every request with request correlation middleware:

- incoming `X-Request-ID` is preserved when provided
- otherwise the API generates a new opaque request ID
- the resolved request ID is echoed back in the response header
- request completion logs include `request_id`, `method`, `path`, `status`, `remote_addr`, and `duration_ms`

The worker boot path now includes a process-level `worker_id` field on startup and heartbeat logs so later job-processing code can reuse the same correlation field.

## Database migrations

Run the current schema migrations from the repository root:

```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5432/event_pipeline?sslmode=disable \
make migrate
```

`make migrate` defaults to `up`. To inspect or roll back the current migration state, pass a command through to the runner:

```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5432/event_pipeline?sslmode=disable \
./scripts/migrate.sh version

DATABASE_URL=postgres://postgres:postgres@localhost:5432/event_pipeline?sslmode=disable \
./scripts/migrate.sh down
```

The migration runner reads:

- `DATABASE_URL` (required)
- `MIGRATIONS_DIR` (optional, defaults to `migrations`)

For local development with Docker Compose, start Postgres first and then run migrations against `localhost:5432` from the repo checkout:

```bash
docker compose -f deploy/compose/docker-compose.yml up -d postgres

DATABASE_URL=postgres://postgres:postgres@localhost:5432/event_pipeline?sslmode=disable \
make migrate
```

For deploy environments, run the same `go run ./cmd/migrate up` command as a release step before starting or updating the API and worker processes. The container build now includes the migration files so the same command can be executed from the built artifact as long as `DATABASE_URL` is set.

`make seed` is still a placeholder command until the seed-data ticket is implemented.

## Roadmap references

- [Project scope](docs/PROJECT_SCOPE.md)
- [Delivery plan](docs/DELIVERY_PLAN.md)
- [Build status](docs/BUILD_STATUS.md)
