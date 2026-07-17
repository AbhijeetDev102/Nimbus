# Nimbus Project Context

## Project Vision

Project Name: **Nimbus**

Nimbus is **NOT** a video transcoding application.

Nimbus is a **Cloud-Native Distributed Job Execution Platform** whose first implemented workload is **video transcoding**.

The platform is designed so that future workloads (image processing, AI inference, document conversion, CSV ETL, etc.) can be added without changing the core execution engine.

The platform layer should remain completely independent from any specific workload.

---

# Core Philosophy

Separate the project into two logical layers.

## Platform Layer

Responsible for:

* Job creation
* Scheduling
* Queueing
* Execution
* Retry
* Cancellation
* Progress tracking
* Heartbeats
* Worker management
* Scaling
* Event publishing
* Observability

This layer knows nothing about videos.

---

## Workload Layer

Current workload:

Video Transcoding

Contains:

* FFmpeg
* FFprobe
* Thumbnail generation
* Preview generation
* Metadata extraction

If tomorrow video processing is removed, the Platform Layer should continue functioning.

This separation is the primary architectural goal.

---

# The 4-Core Service Architecture

Nimbus is composed of four primary microservices (with an optional fifth for search), designed to strictly separate responsibilities:

| Service | Responsibility |
|---|---|
| **API Gateway** | Routes external requests via REST and WebSocket. |
| **Resource Service** | Uploads, metadata, storage abstraction. Knows about resources, not jobs. |
| **Job Service** | Creates and manages jobs, publishes events via Outbox. Knows about jobs, not execution. |
| **Worker Service** | Event-driven execution engine. Consumes from Kafka, runs tasks (video, image, etc.), updates progress. No HTTP API. |
| **Search Service (Optional)** | Read-only queries and filtering. |

*Note: The "Scheduler" is handled natively by Kafka Consumer Groups. The "Dispatcher" is simply a Go package inside the Worker Service.*

---

# Current Project Progress

Completed:

* Kubernetes local development environment using Tilt
* PostgreSQL StatefulSet
* MinIO StatefulSet
* API Gateway
* gRPC communication
* Resource Service
* GORM integration
* UUID based Resource model
* Presigned URL generation
* Direct upload architecture
* PostgreSQL metadata persistence
* Production-style abstractions
* Repository pattern
* Environment configuration package

Current estimated completion:

Approximately **35-40%**

Architecture foundation is considered complete.

---

# Upload Flow (Final Decision)

Upload uses Presigned URLs.

Flow:

```
Client
↓
POST /videos/upload-url
↓
API Gateway
↓
Resource Service
↓
Generate UUID
↓
Insert Video Metadata (UPLOAD_REQUESTED)
↓
Generate Presigned PUT URL
↓
Return
↓
Client uploads directly to MinIO
↓
HTTP 200 from MinIO
↓
POST /jobs
```

Backend never streams video bytes.

Large files never pass through Go services.

---

# Why Upload and Job Creation are Separate

Upload success does not automatically imply processing.

Client explicitly decides when processing begins.

Flow:

```
Upload
↓
Success
↓
POST /jobs
```

This allows:

* validation
* authorization
* duplicate detection
* object existence verification
* future quota enforcement

---

# Job Creation

POST /jobs is considered the point where the platform begins processing.

Responsibilities:

* Verify object exists in MinIO
* Update video status
* Create Job
* Create Outbox event
* Commit Transaction

After commit:

Debezium publishes event.

Workers begin processing.

---

# Current API Structure

Upload URL

```
POST /videos/upload-url
```

Create Job

```
POST /jobs
```

Search

```
GET /jobs/{id}
```

Download

Future:

```
GET /videos/{id}/download
```

---

# PostgreSQL is the Source of Truth

Very important design decision.

Every important state change is written to PostgreSQL first.

Examples:

* Upload requested
* Upload completed
* Job created
* Processing started
* Processing completed
* Failed
* Cancelled

Kafka is never considered the source of truth.

---

# Redis Usage

Redis stores runtime information only.

Examples:

```
progress:<job>
heartbeat:<worker>
cancel:<job>
lease:<job>
```

Redis never stores durable business data.

---

# Kafka Usage

Kafka is only responsible for asynchronous work distribution.

Kafka should never become the source of truth.

---

# CDC Decision

Current decision:

Use Debezium CDC with the Transactional Outbox Pattern.

Flow:

```
Resource Service
↓
BEGIN
↓
Update Video
↓
Insert Job
↓
Insert Outbox Event
↓
COMMIT
↓
Debezium
↓
Kafka
↓
Workers
```

Application never publishes directly to Kafka (planned architecture).

Reason:

Avoid Dual Write Problem.

---

# Important Discussion Today

Question:

"If we are not using database events, why use CDC?"

Answer:

CDC is NOT watching every database update.

CDC should watch an Outbox table.

Application explicitly inserts business events into the Outbox.

Example:

```
VideoUploaded
JobCreated
JobCancelled
JobCompleted
```

Debezium transports these events to Kafka.

It is not publishing arbitrary UPDATE statements.

---

# Discussion About Kafka Before PostgreSQL

Question:

Should Kafka sit before PostgreSQL to absorb traffic bursts?

Example:

```
API
↓
Kafka
↓
Consumer
↓
Database
```

Decision:

NO.

Reason:

Nimbus is a transactional system.

Client expects:

```
POST /jobs
↓
201 Created
```

This should guarantee that the job exists.

Database must remain authoritative.

---

Kafka-first architecture is appropriate for:

* Analytics
* Logs
* Telemetry
* IoT
* Clickstream

Nimbus is not one of these systems.

---

# Handling Database Bursts

Instead of putting Kafka before PostgreSQL, use:

* Connection Pooling
* Kubernetes Horizontal Scaling
* Rate Limiting
* Backpressure
* Read Replicas (future)
* PostgreSQL tuning

Only redesign ingestion if platform reaches extremely large scale.

---

# Job Processing Architecture Discussion

Current architecture:

```
PostgreSQL
↓
Debezium
↓
Kafka
↓
Workers
```

Original design had:

```
Kafka
↓
JobConsumer
↓
Executor Services
```

Discussion outcome:

If JobConsumer only forwards Kafka messages, it is unnecessary because Kafka Consumer Groups already provide:

* load balancing
* failover
* partition assignment

Workers can consume Kafka directly.

---

Future possibility:

Introduce JobConsumer only if it becomes a true Scheduler.

Responsibilities could include:

* Priority scheduling
* Fair scheduling
* Tenant quotas
* Worker capability matching
* GPU routing
* Delayed jobs
* Cron scheduling

Until then,

Workers should consume Kafka directly.

---

# Worker Philosophy

Workers are NOT "Video Workers."

Workers are generic Job Executors.

Current mental model:

Bad:

```
ProcessVideo(videoID)
```

Preferred:

```
Execute(Job)
```

Dispatcher:

```
switch(job.Type)
VIDEO
IMAGE
CSV
AI
PDF
```

Workers should execute handlers rather than being hardcoded to video logic.

---

# Event Payload Philosophy

Workers should receive complete Job information.

Example:

```
JobID
VideoID
ObjectKey
Pipeline
RetryCount
Priority
TenantID
```

Workers should not immediately query PostgreSQL just to know what to execute.

---

# Worker Lifecycle

```
Receive Kafka Message
↓
Update DB Status = RUNNING
↓
Download From MinIO
↓
FFprobe
↓
FFmpeg
↓
Thumbnail
↓
Upload Result
↓
Update PostgreSQL
↓
Insert Outbox Event
↓
Commit
↓
Kafka Offset Commit
```

Offset should only be committed after successful database transaction.

---

# Failure Philosophy

Every arrow in the architecture must answer:

"What happens if the process crashes here?"

Examples:

* Crash after upload
* Crash before Kafka publish
* Crash after worker download
* Crash before DB update
* Crash after upload but before offset commit

System should be designed around failure handling.

---

# Current Architectural Insight

Most important realization from today's discussion:

Nimbus should not be presented as:

"A Video Transcoding Platform."

Instead:

"A Cloud-Native Distributed Job Execution Platform with Video Transcoding as its first workload."

This distinction changes the architecture.

Platform:

* Scheduler
* Worker
* Queue
* Retry
* Scaling
* Monitoring

Workload:

* FFmpeg
* Video Processing

Platform should remain reusable.

---

# Clean Architecture Guidelines

Nimbus adheres strictly to Hexagonal (Ports & Adapters) Architecture.

## Domain Naming & Interfaces
The domain defines **what it needs (Ports)**, not **how it's implemented (Adapters)**. 
- `ResourceRepository`: Manages CRUD operations for the Resource domain entity (implemented by PostgreSQL).
- `StorageProvider`: Manages blob storage operations like generating presigned URLs or checking object existence (implemented by MinIO/S3).

## Infrastructure Separation
Infrastructure (Adapters) must be explicitly separated by technology type inside the `internal/infrastructure/` folder:
- `persistence/`: Adapters for databases (e.g., `postgres.go`).
- `storage/`: Adapters for object storage (e.g., `minio.go`, `s3.go`).
- `grpc/`: Adapters for gRPC handlers.
- `events/`: Adapters for message brokers (Kafka/RabbitMQ).

## Domain Modeling
- Always use strongly typed enums instead of raw strings for status fields (e.g., `type ResourceStatus string`).
- Models should be storage-agnostic but track where they live (e.g., `StorageProvider` and `Bucket` fields) to support multi-cloud deployments.

---

# Folder Organization Philosophy

Whenever creating packages ask:

Platform?

or

Video-specific?

Platform examples:

```
scheduler/
worker/
dispatcher/
retry/
heartbeat/
progress/
queue/
events/
```

Video-specific:

```
ffmpeg/
thumbnail/
preview/
metadata/
transcoder/
```

Never mix workload logic into the execution platform.

---

# Immediate Next Step

Next implementation task:

Design and implement:

```
POST /jobs
```

Responsibilities:

* Verify MinIO object exists
* Update video status
* Create Job
* Insert Outbox event
* Commit transaction

After this:

* Integrate Debezium
* Kafka
* Worker Consumer Group
* Generic Worker Dispatcher

No new technologies should be introduced before completing this pipeline.

---

# Overall Project Goal

By completion, Nimbus should resemble a production backend infrastructure product rather than a student project.

Desired interview description:

> "Nimbus is a cloud-native distributed job execution platform that reliably executes long-running asynchronous workloads. The platform provides scheduling, retries, worker orchestration, progress tracking, observability, and fault recovery. Video transcoding is implemented as the first workload to demonstrate the platform's capabilities."

This is the guiding architectural vision going forward.
