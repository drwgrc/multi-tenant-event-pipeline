# AGENTS.md

## Project overview
This repository is a Go backend for a multi-tenant event ingestion and analytics system.
The system accepts event batches over HTTP, enforces tenant isolation and idempotency,
processes events asynchronously, materializes aggregate buckets, and exposes analytics,
ops, and demo endpoints.

The primary source documents for this repository are:
- `docs/PROJECT_SCOPE.md` — condensed system goals, architecture, and constraints.
- `docs/BUILD_STATUS.md` — current implementation status. Treat this as the source of truth for what is already built.
- `docs/DELIVERY_PLAN.md` — milestone-by-milestone backlog summary derived from the imported GitHub issues.
- `PRD.md` — original long-form PRD.
- `import.json` — imported GitHub issue dataset.

## What to assume
- This is a backend-first project. Avoid spending time on polished UI unless a task explicitly requires it.
- Prefer the simplest design that satisfies the PRD and the current milestone.
- The repository should tell a believable production-systems story: correctness under retries, async processing, tenant isolation, observability, and benchmark honesty matter more than feature breadth.
- If `docs/BUILD_STATUS.md` says a subsystem is not built yet, do not act as if it exists.

## How to work in this repo
When asked to implement or change something:
1. Read `docs/BUILD_STATUS.md` first.
2. Read the matching section in `docs/DELIVERY_PLAN.md`.
3. Check the relevant milestone and acceptance criteria in GitHub issues or `import.json` if needed.
4. Make the smallest coherent change that moves the current milestone forward.
5. Update `docs/BUILD_STATUS.md` if you materially change repository status.

## Engineering priorities
In priority order:
1. Ingestion correctness and idempotency
2. Postgres schema and transaction boundaries
3. Worker safety, retries, and dead-letter handling
4. Retry-safe aggregate writes
5. Analytics reads from rollups, never raw events for charts
6. Tenant isolation and auth boundaries
7. Observability and benchmark evidence
8. Demo and documentation polish

## Constraints
- Do not add Kafka, Kubernetes, GraphQL, funnels, or polished dashboard work in v1.
- Do not query raw events for analytics chart endpoints.
- Every tenant-scoped read/write must filter by `tenant_id` derived from auth context.
- Idempotency is scoped by `(tenant_id, idempotency_key)`.
- Duplicate same-hash requests return the stored response snapshot.
- Duplicate different-hash requests return `409 Conflict`.
- Worker retries must not double-count aggregates.

## Repository shape to preserve
Expected layout:
- `cmd/api`
- `cmd/worker`
- `cmd/loadgen`
- `internal/analytics`
- `internal/auth`
- `internal/config`
- `internal/events`
- `internal/httpapi`
- `internal/ingest`
- `internal/jobs`
- `internal/middleware`
- `internal/observability`
- `internal/store`
- `internal/tenant`
- `migrations`
- `deploy/compose`
- `scripts`
- `docs`

## Definition of done for code changes
A change is not complete until:
- code and docs are consistent
- the affected milestone acceptance criteria are addressed
- any new environment variables, commands, or endpoints are documented
- `docs/BUILD_STATUS.md` is updated when the repository state changes

## Status discipline
This project has both a plan and an implementation state.
Always distinguish between:
- planned work
- partially implemented work
- completed work

Never claim a feature exists just because it is in the PRD or backlog.
