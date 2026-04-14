# Delivery Plan

This file summarizes the imported GitHub backlog in a form optimized for coding agents.
Use it to map a task to the right milestone and to understand what belongs in scope.

## 0 - Bootstrap
### Epics
- EPIC: Bootstrap and local dev foundation
### Tickets
- [P0] Initialize repo, Go module, Makefile, and directory scaffold
- [P0] Create Docker Compose stack for API, worker, Postgres, and Redis
- [P0] Implement config loader and environment validation
- [P0] Set up structured slog logging and request ID helpers
- [P0] Add migration runner and tenancy/auth core schema
- [P0] Seed demo tenant, admin user, source, and API key
- [P0] Add GET /livez and GET /readyz
- [P0] Write README skeleton with benchmark target and architecture sketch

## 1 - Ingestion & Idempotency
### Epics
- EPIC: Event ingestion and idempotent write path
### Tickets
- [P0] Define ingestion request DTOs and validation rules
- [P0] Enforce request body size and batch count limits
- [P0] Implement API key auth and tenant resolution for ingestion
- [P0] Implement stable request hashing for idempotency
- [P0] Persist ingestion_requests, events, and rejected_events in one transaction
- [P0] Implement partial rejection persistence and response formatting
- [P0] Implement POST /v1/events
- [P0] Implement same-hash idempotency replay behavior
- [P0] Implement different-hash idempotency conflict behavior
- [P0] Implement GET /v1/events for raw event browsing
- [P0] Implement GET /v1/events/{event_id}
- [P0] Add ingest API tests for happy path, partial rejection, and idempotency
- [P0] Build basic cmd/loadgen traffic generator

## 2 - Worker & Retry Safety
### Epics
- EPIC: Worker, retries, and dead-letter handling
### Tickets
- [P0] Add jobs and dead_letters migrations
- [P0] Implement job repository and claim logic with FOR UPDATE SKIP LOCKED
- [P0] Build worker runner with long-lived loop and graceful shutdown
- [P0] Implement process_ingestion job state machine
- [P0] Add retry backoff scheduling and dead-letter handoff
- [P0] Add recovery sweep for jobs stuck in running
- [P0] Implement GET /v1/jobs
- [P0] Implement GET /v1/dead-letters
- [P0] Implement POST /v1/jobs/{job_id}/retry
- [P0] Add integration tests for claim, retry, DLQ, and crash recovery

## 3 - Aggregation & Analytics
### Epics
- EPIC: Enrichment, rollups, and analytics reads
### Tickets
- [P0] Implement enrichment helpers for supported dimensions
- [P0] Add aggregate_buckets migration and indexes
- [P0] Implement rollup upsert path for aggregate_buckets
- [P0] Prevent double-counting with aggregated_at guard and locking
- [P0] Implement GET /v1/analytics/top-events
- [P0] Implement GET /v1/analytics/timeseries
- [P0] Implement GET /v1/analytics/breakdown
- [P0] Add analytics correctness integration tests
- [P0] Write design note: analytics must never query raw events for charting

## 4 - Tenant Isolation & Hardening
### Epics
- EPIC: Tenant isolation, auth, and rate limiting
### Tickets
- [P0] Implement POST /v1/auth/login and GET /v1/me
- [P0] Add admin auth middleware and RBAC enforcement
- [P0] Implement sources endpoints
- [P0] Implement API key management endpoints
- [P0] Audit tenant scoping across every repository query and write path
- [P0] Reject revoked API keys on the ingestion path
- [P1] Add keyset pagination for events, jobs, and dead letters
- [P1] Implement per-tenant and per-API-key rate limiting
- [P1] Add tenant isolation and auth test suite

## 5 - Observability & Performance
### Epics
- EPIC: Observability, testing, and performance proof
### Tickets
- [P1] Add structured request and worker logs with required fields
- [P1] Expose Prometheus metrics for HTTP, ingest, jobs, analytics, and rate limits
- [P1] Add OpenTelemetry tracing across API, DB, worker, and analytics paths
- [P1] Expose and protect pprof endpoints
- [P1] Wire Prometheus, Grafana, and Jaeger into the local compose stack
- [P1] Add Redis analytics cache and singleflight suppression
- [P1] Build load test scenarios and benchmark harness
- [P1] Add chaos drills for worker kill, DB interruption, Redis loss, and API restart
- [P1] Add race detector, fuzz tests, and govulncheck to CI
- [P1] Publish benchmark report and current bottlenecks note

## 6 - Demo & Documentation Polish
### Epics
- EPIC: Demo, deployment, and documentation polish
### Tickets
- [P2] Seed synthetic demo data and refresh scripts
- [P2] Build public /demo page and GET /v1/demo/status
- [P2] Implement POST /v1/demo/replay-sample
- [P2] Deploy to a low-cost VM with TLS reverse proxy
- [P2] Write ADR: why Postgres-backed jobs in v1
- [P2] Write README sections for benchmark target, public demo, and where concurrency lives
- [P2] Publish supporting artifacts: demo walkthrough, screenshots, architecture notes, scale notes, and OpenAPI

## Recommended implementation order
1. Bootstrap and local development foundation
2. Ingestion API, validation, and idempotency
3. Jobs, worker processing, retries, and dead letters
4. Enrichment and aggregate buckets
5. Analytics reads from rollups
6. Tenant admin APIs, RBAC, pagination, and rate limiting
7. Observability, load generation, performance validation, and demo deployment

## Guardrails
- Do not jump ahead to UI polish before the write path, worker path, and rollups exist.
- Do not mark analytics complete until they read from `aggregate_buckets` only.
- Do not mark reliability work complete without retry and restart evidence.
- Prefer narrow vertical slices that can be demoed end-to-end.
