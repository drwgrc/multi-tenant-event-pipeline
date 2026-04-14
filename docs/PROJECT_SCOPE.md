# Project Scope

## One-sentence summary
A multi-tenant event ingestion and analytics backend in Go using Postgres as the system of record and a Postgres-backed jobs table for async processing.

## Core write path
1. `POST /v1/events` authenticates an ingestion API key.
2. The API validates the request and computes a request hash.
3. The service enforces idempotency using `(tenant_id, idempotency_key)`.
4. The write transaction inserts the ingestion request, valid raw events, rejected events, and a `process_ingestion` job.
5. The API returns `202 Accepted` with a response snapshot that can be replayed for duplicate same-hash requests.

## Core async path
1. A worker claims a queued job with `FOR UPDATE SKIP LOCKED`.
2. The worker loads the ingestion request and associated events.
3. The worker enriches events into an allowlisted set of dimensions.
4. The worker upserts aggregate buckets.
5. The worker marks the ingestion as processed and the job as succeeded, or retries/dead-letters on failure.

## Core read path
- Raw event browsing reads from `events`.
- Analytics reads from `aggregate_buckets`.
- Ops reads from `jobs` and `dead_letters`.

## Important constraints
- Analytics endpoints must not query raw events for charting.
- v1 dimensions are limited to `page_path`, `referrer_domain`, `device_type`, `source_id`, and `source_name`.
- v1 uses Postgres-backed jobs instead of Kafka or Redis Streams.
- v1 includes rate limiting, logs, metrics, traces, health checks, pprof, load generation, chaos tests, and a public demo flow.
- v1 excludes advanced dashboarding, cohorts, funnels, Kubernetes, and public multi-language SDKs.

## Performance targets
- 500 events/sec sustained end-to-end on a single 2 vCPU / 4 GB / 40 GB benchmark box
- ingestion p95 under 200 ms at the benchmark target
- aggregation lag p95 under 5 seconds at the benchmark target
- no duplicate aggregate writes under retry or worker restart

## Key correctness requirements
- same idempotency key + same payload => return original response snapshot
- same idempotency key + different payload => `409 Conflict`
- worker retries must not lose events or double-count aggregates
- tenant A must not read tenant B data
