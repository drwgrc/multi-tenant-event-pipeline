# Multi-tenant Event Pipeline

Go backend for a multi-tenant event ingestion and analytics system. The intended system accepts tenant-scoped event batches over HTTP, enforces idempotency, processes events asynchronously, and serves analytics from rollups instead of raw-event scans.

## Current status

The repository is still in the bootstrap phase. These pieces are implemented today:

- Go module and repository scaffold
- command entrypoints for `api`, `worker`, `loadgen`, `migrate`, and `seed`
- local Docker Compose stack for API, worker, Postgres, and Redis
- shared config loading and validation
- structured JSON logging and request ID propagation
- initial tenancy and auth schema for `tenants`, `users`, `memberships`, `api_keys`, and `sources`
- deterministic local seed flow for one demo tenant, admin user, source, and API key
- `GET /livez` and `GET /readyz`

These are still planned and not implemented yet:

- `POST /v1/events`
- ingestion validation and idempotency behavior
- jobs, worker processing, retries, and dead letters
- aggregate rollups and analytics endpoints
- login, tenant admin APIs, and demo flow

## Benchmark target

The PRD target for this system is:

- `500 events/sec` sustained end-to-end
- single `2 vCPU / 4 GB / 40 GB` benchmark box
- `25-event` batches during benchmarking
- ingestion `p95 < 200 ms`
- aggregation lag `p95 < 5 seconds`

This repository does not meet or prove that target yet. Benchmark evidence and performance reports will land in later milestones.

## Architecture sketch

### Write path

1. `POST /v1/events` authenticates an ingestion API key.
2. The API validates the request and computes a stable request hash.
3. The service enforces idempotency on `(tenant_id, idempotency_key)`.
4. One transaction persists the ingestion request, accepted raw events, rejected events, and a `process_ingestion` job.
5. The API returns `202 Accepted` and stores a response snapshot for same-hash replay behavior.

### Async path

1. A worker claims a queued `process_ingestion` job.
2. The worker loads the ingestion request and associated events.
3. The worker enriches supported dimensions.
4. The worker upserts aggregate buckets.
5. The worker marks the ingestion and job as complete, or retries/dead-letters on failure.

### Read path

- Raw event browsing reads from `events`.
- Analytics reads from `aggregate_buckets`.
- Ops and retry inspection read from `jobs` and `dead_letters`.

## Repository layout

```text
cmd/                runnable binaries
internal/           application packages by subsystem
migrations/         database migration files
deploy/compose/     local Docker Compose stack
scripts/            developer helper scripts
docs/               scope, delivery plan, and build status
```

## Local development

Common commands:

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

Start the local stack:

```bash
make up
```

The Compose stack starts:

- `api` on `http://localhost:8080`
- `postgres` on `localhost:5432`
- `redis` on `localhost:6379`

Compose injects local-development settings into the API and worker containers:

```text
APP_ENV=development
HTTP_ADDR=:8080
DATABASE_URL=postgres://postgres:postgres@postgres:5432/event_pipeline?sslmode=disable
REDIS_URL=redis://redis:6379
JWT_SIGNING_KEY=development-signing-key
INGEST_MAX_BODY_BYTES=1048576
INGEST_MAX_BATCH_EVENTS=1000
```

The Compose file also waits for Postgres and Redis health before starting the application containers.

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

## Bootstrap flow

Apply the current schema:

```bash
make migrate
```

Seed the local demo records:

```bash
make seed
```

The seed flow is deterministic and rerunnable. It upserts the same local-only records each time:

- tenant slug: `demo`
- tenant name: `Demo Tenant`
- admin email: `admin@demo.local`
- admin password: `demo-admin-password`
- membership role: `admin`
- source name: `demo-web`
- API key name: `demo-ingest`

`make seed` prints the raw seeded API key once during the run. The database stores only the key hash plus the prefix (`evt_demo_loc`), not the raw secret.

This bootstrap data exists so later tickets can add login and ingestion against a real tenant without extra manual setup. Those APIs are not available yet.

## Current API surface

The bootstrap API currently exposes:

- `GET /livez` for process liveness
- `GET /readyz` for readiness against Postgres and Redis

Both endpoints return JSON. `/readyz` returns `200 OK` only when both dependencies are reachable and `503 Service Unavailable` when either check fails.

## Configuration

The binaries read configuration from environment variables at startup. There is no automatic `.env` loading.

Required:

- `DATABASE_URL`
- `REDIS_URL` or temporary fallback `REDIS_ADDR`
- `JWT_SIGNING_KEY` with at least 16 characters

Development defaults:

- `APP_ENV=development` when unset
- `HTTP_ADDR=:8080` for the API when `APP_ENV=development`
- `INGEST_MAX_BODY_BYTES=1048576`
- `INGEST_MAX_BATCH_EVENTS=1000`

Invalid or missing configuration causes startup to fail before serving traffic.

## Roadmap references

- [Project scope](docs/PROJECT_SCOPE.md)
- [Delivery plan](docs/DELIVERY_PLAN.md)
- [Build status](docs/BUILD_STATUS.md)
