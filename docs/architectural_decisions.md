# Architecture Decision Records (ADR) - Nimbus

This document tracks all major architectural and design decisions made during the development of the Nimbus Video Transcoding Platform.

---

## ADR 001: Direct-to-Storage Video Uploads via Presigned URLs
**Date:** 2026-07-12\
**Context:** Video files can be massive (gigabytes in size). If uploaded directly through the API Gateway and Video Service, it would tie up Go routines, consume massive amounts of server RAM, and create a severe network bottleneck.\
**Decision:** We implemented Presigned PUT URLs via MinIO.\
**Consequences:** The Go backend only handles lightweight metadata and URL generation. The heavy lifting of the actual file transfer is offloaded directly from the client to the storage layer, allowing the Go services to remain highly concurrent and scale infinitely.

## ADR 002: PostgreSQL as the Source of Truth
**Date:** 2026-07-12\
**Context:** We need a way to track the lifecycle of a video (`PENDING_UPLOAD`, `UPLOADED`, `PROCESSING`, `COMPLETED`) and map the obscure MinIO `ObjectKey` (UUID) back to the user's `OriginalFileName` for the frontend UI.\
**Decision:** We introduced PostgreSQL as the central metadata store, integrated via GORM. The UUID serves as both the database Primary Key and the MinIO ObjectKey.\
**Consequences:** This creates a strict separation of concerns: MinIO holds raw binary data, while Postgres holds relational metadata and state.

## ADR 003: Interface Segregation in the Repository Layer
**Date:** 2026-07-12\
**Context:** The Video Service needs to interact with both MinIO and PostgreSQL. Putting both clients into a single composite repository violates the Single Responsibility Principle.\
**Decision:** We split the `VideoRepository` into two distinct interfaces: `StorageRepository` (implemented by MinIO) and `MetaDataRepository` (implemented by Postgres). The Service layer orchestrates between them by accepting both as dependencies.\
**Consequences:** The architecture remains highly modular and testable (we can mock the database and storage independently), strictly following Clean Architecture principles.

## ADR 004: StatefulSets for Infrastructure Components
**Date:** 2026-07-12\
**Context:** Running databases (Postgres) and object storage (MinIO) in Kubernetes requires stable, persistent storage. \
**Decision:** We explicitly chose `StatefulSet` over `Deployment` or `ReplicaSet` for these components, paired with `volumeClaimTemplates`. We also pinned images to specific, lightweight versions (e.g., `postgres:15-alpine` instead of `latest`).\
**Consequences:** We guarantee data persistence across pod restarts, avoid split-brain scenarios by enforcing `replicas: 1` for local dev, and keep memory footprints extremely low (avoiding OOMKilled errors).

---

## ADR 005: Outbox Pattern & Debezium (CDC) over Dual Writes
**Date:** 2026-07-13\
**Context:** When a job is created or a video state changes, we must notify the worker ecosystem via Kafka. Doing a dual write (saving to Postgres, then publishing to Kafka) is dangerous because if the process crashes between the DB write and Kafka publish, the event is permanently lost, leaving the job orphaned.\
**Decision:** We will use the Transactional Outbox Pattern combined with Debezium (Change Data Capture). The application explicitly inserts business events (`JobCreated`, `VideoUploaded`) into an `Outbox` table as part of the same database transaction. Debezium watches this table and securely transports the events to Kafka.\
**Consequences:** We achieve 100% guaranteed delivery to Kafka without the Dual Write Problem. It requires slightly more infrastructure setup (Debezium).

## ADR 006: PostgreSQL Before Kafka (No Kafka-First Ingestion)
**Date:** 2026-07-13\
**Context:** Should we put Kafka directly behind the API Gateway to absorb massive traffic bursts and write to Postgres asynchronously?\
**Decision:** No. Kafka will sit *after* Postgres in the flow.\
**Consequences:** Nimbus is a transactional platform. When a user calls `POST /jobs`, they expect a `201 Created` response to guarantee the job is securely recorded. Kafka-first ingestion is suited for telemetry/analytics, not transactional job creation. We will handle traffic bursts using connection pooling, horizontal scaling, and rate limiting instead.

## ADR 007: Direct Kafka Consumption by Workers
**Date:** 2026-07-13\
**Context:** Originally, the architecture proposed a `JobConsumer` service that reads from Kafka and forwards jobs to `Executor Services`.\
**Decision:** Workers will consume directly from Kafka. We will not build an intermediary `JobConsumer` service unless it evolves into a true, complex Scheduler (handling priority, fairness, tenant quotas, etc.).\
**Consequences:** We leverage Kafka Consumer Groups' native capabilities for load balancing, failover, and partition assignment, drastically simplifying the architecture.

## ADR 008: Platform vs Workload Separation
**Date:** 2026-07-13\
**Context:** Designing Nimbus purely as a video transcoder couples the execution engine to FFmpeg, preventing reuse.\
**Decision:** Nimbus is officially defined as a generic **Cloud-Native Distributed Job Execution Platform**. Video Transcoding is simply the *first workload*.\
**Consequences:** The codebase will have strict package separation between Platform logic (`scheduler/`, `worker/`, `queue/`) and Workload logic (`ffmpeg/`, `thumbnail/`). Workers execute generic `Job` interfaces, dynamically routing to workload-specific handlers.

---

## ADR 009: Outbox Event Router (SMT) and Header-Based Event Filtering
**Date:** 2026-08-23\
**Context:** Raw Change Data Capture (CDC) events from Debezium contain database-level WAL metadata (`before`, `after`, `source`, `op`, `ts_ms`), heavily coupling downstream workers to database table internals. Furthermore, workers often need to filter or route messages by event type without the CPU overhead of deserializing entire JSON payloads.\
**Decision:** We use Debezium's `EventRouter` Single Message Transformation (SMT). It extracts the JSON `payload` column directly as the Kafka message value, sets the Kafka message key to `aggregate_id` (Job UUID) for strict per-job partition ordering, routes to `<aggregate_type>.events` (`job.events`), and places `event_type` directly into the Kafka message record headers (`eventType`).\
**Consequences:** Workers receive clean, unpolluted domain payloads. Placing `eventType` in Kafka headers enables zero-deserialization event filtering. Partition ordering is guaranteed per job aggregate.

## ADR 010: Idempotent Execution & Atomic Conditional Updates for Worker Deduplication
**Date:** 2026-08-23\
**Context:** At-least-once message delivery in Kafka and Debezium means duplicate events can reach workers. Separate `SELECT` then `UPDATE` operations create a Time-of-Check to Time-of-Use (TOCTOU) race condition where multiple workers might process the exact same job concurrently.\
**Decision:** Workers claim jobs using atomic conditional updates in PostgreSQL (`UPDATE jobs SET status = 'RUNNING', worker_id = :workerId, started_at = NOW() WHERE id = :jobId AND status = 'QUEUED'`). The number of affected rows dictates job ownership: `RowsAffected == 1` grants the execution lease; `RowsAffected == 0` safely discards/acknowledges the duplicate event.\
**Consequences:** Eliminates race conditions and duplicate workload execution natively at the database layer without needing complex external locking mechanisms.

