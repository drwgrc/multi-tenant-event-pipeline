# Multi-tenant Event Pipeline

Bootstrap repository for a Go backend that ingests tenant-scoped event batches, processes them asynchronously, and serves analytics from rollups.

## Current state

This repository currently contains the bootstrap scaffold from Issue `#2` plus the local infrastructure stack from Issue `#3`:

- Go module initialization
- command entrypoints for `api`, `worker`, and `loadgen`
- internal package layout for planned subsystems
- local Docker Compose stack for API, worker, Postgres, and Redis
- bootstrap developer commands in `Makefile`

The business logic, config loader, migrations, seed data, readiness checks, and ingestion pipeline are tracked in later tickets and are not implemented yet.

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
REDIS_ADDR=redis:6379
```

The compose file also configures health checks for Postgres and Redis so the `api` and `worker` containers wait for those dependencies before starting.

Run the API directly:

```bash
make run-api
```

Run the worker directly:

```bash
make run-worker
```

The bootstrap API currently exposes a minimal liveness endpoint at `http://localhost:8080/livez`.

The application binaries do not consume the full compose env contract yet. The config loader, migrations, seed flow, and readiness checks are tracked in later milestone-0 tickets.

`make migrate` and `make seed` are placeholder commands that fail intentionally until the migration and seed tickets are implemented.

## Roadmap references

- [Project scope](docs/PROJECT_SCOPE.md)
- [Delivery plan](docs/DELIVERY_PLAN.md)
- [Build status](docs/BUILD_STATUS.md)
