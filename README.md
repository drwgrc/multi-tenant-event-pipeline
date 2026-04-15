# Multi-tenant Event Pipeline

Bootstrap repository for a Go backend that ingests tenant-scoped event batches, processes them asynchronously, and serves analytics from rollups.

## Current state

This repository currently contains the bootstrap scaffold from Issue `#2`:

- Go module initialization
- command entrypoints for `api`, `worker`, and `loadgen`
- internal package layout for planned subsystems
- local Docker Compose wiring for API, worker, Postgres, and Redis
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

Run the API directly:

```bash
make run-api
```

Run the worker directly:

```bash
make run-worker
```

The bootstrap API currently exposes a minimal liveness endpoint at `http://localhost:8080/livez`.

`make migrate` and `make seed` are placeholder commands that fail intentionally until the migration and seed tickets are implemented.

## Roadmap references

- [Project scope](docs/PROJECT_SCOPE.md)
- [Delivery plan](docs/DELIVERY_PLAN.md)
- [Build status](docs/BUILD_STATUS.md)
