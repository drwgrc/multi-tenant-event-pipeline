# Build Status

## Read this first
This file is the machine-readable human summary of what is already implemented in the repository.
Update it whenever a milestone meaningfully changes status.

## Current overall state
- Repository planning exists.
- GitHub issues and milestones exist.
- Bootstrap repository foundation is implemented.
- Remaining major subsystems are still mostly `not_started` or placeholder-only until later tickets land.

## Status values
- `not_started`
- `in_progress`
- `partial`
- `complete`
- `blocked`

## Subsystem status
| Subsystem | Status | Notes |
|---|---|---|
| Repo scaffold and Go module | complete | Go module, Makefile bootstrap commands, and repository scaffold are committed. |
| Docker Compose local stack | partial | Compose now includes API, worker, Postgres, and Redis with named volumes, exposed dev ports, health-gated startup, and explicit env wiring; application internals still do not consume the full config contract. |
| Config loader | complete | Shared startup config parsing and validation now power `cmd/api` and `cmd/worker`; `REDIS_ADDR` remains a temporary fallback alias behind canonical `REDIS_URL`. |
| Structured logging | complete | API and worker now emit JSON `slog` logs; API propagates `X-Request-ID` and worker logs include `worker_id` correlation. |
| Database migrations | not_started | Base schema and seed path. |
| Seed data | not_started | Demo tenant, source, API key, admin user. |
| Health endpoints | partial | Bootstrap `/livez` exists in `cmd/api`; `/readyz` and dependency checks are pending. |
| Ingestion API | not_started | `POST /v1/events`. |
| Event validation | not_started | Batch and event-level validation, partial rejection. |
| Idempotency | not_started | Request hashing, replay, `409 Conflict`. |
| Ingestion persistence | not_started | `ingestion_requests`, `events`, `rejected_events`. |
| Raw event reads | not_started | `GET /v1/events`, `GET /v1/events/{id}`. |
| Jobs table and repositories | not_started | Claiming and status transitions. |
| Worker loop | not_started | Long-lived poll/claim/process/shutdown path. |
| Retry and dead-letter handling | not_started | Backoff, DLQ persistence, retry endpoint. |
| Enrichment pipeline | not_started | Allowlisted dimensions only. |
| Aggregate buckets | not_started | Retry-safe upsert logic. |
| Analytics endpoints | not_started | Top events, timeseries, breakdown. |
| Tenant admin APIs | not_started | API keys, sources, usage. |
| Tenant isolation and RBAC | not_started | Query scoping and permissions. |
| Rate limiting | not_started | Tenant/API key token bucket. |
| Redis caching | not_started | Analytics only, short TTLs. |
| Metrics | not_started | Prometheus counters/histograms. |
| Tracing | not_started | OTEL spans for ingest, jobs, analytics. |
| pprof | not_started | Protected or non-prod only. |
| Load generator | partial | `cmd/loadgen` scaffold exists; no traffic generation scenarios yet. |
| Benchmarks and perf reports | not_started | pprof, EXPLAIN, bottlenecks. |
| Demo flow | not_started | `/demo`, status, replay sample. |
| Deployment | not_started | Public demo deployment and docs. |
| README and docs polish | partial | Bootstrap README exists; architecture, benchmark, and runbook content are still pending. |

## Update rules
When updating this file:
- prefer `partial` over `complete` unless the acceptance criteria are fully met
- add a short note for gaps, shortcuts, or known missing pieces
- do not mark performance or reliability features complete without evidence
