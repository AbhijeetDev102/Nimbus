# Nimbus Sequence Diagrams

These sequence diagrams represent the "source of truth" for the Nimbus Job Execution Platform architecture.

## Sequence 1 — Upload & Job Creation

```mermaid
sequenceDiagram
    autonumber

    actor User
    participant Gateway as API Gateway
    participant Resource as Resource Service
    participant MinIO
    participant Job as Job Service
    participant Postgres
    participant Outbox
    participant Debezium
    participant Kafka
    participant Worker

    %% Upload URL
    User->>Gateway: POST /resources/upload-url
    Gateway->>Resource: gRPC GenerateUploadURL()

    Resource->>Postgres: INSERT Resource(status=UPLOAD_REQUESTED)

    Resource->>MinIO: Generate Presigned URL

    MinIO-->>Resource: Upload URL

    Resource-->>Gateway: Upload URL
    Gateway-->>User: Upload URL

    %% Direct Upload
    User->>MinIO: PUT Video
    MinIO-->>User: HTTP 200 OK

    %% Create Job
    User->>Gateway: POST /jobs(resourceId, jobType, pipeline)

    Gateway->>Job: gRPC CreateJob()

    Job->>Resource: Verify Resource Exists()

    Resource->>MinIO: HEAD Object

    MinIO-->>Resource: Exists

    Resource-->>Job: Resource Verified

    %% Transaction
    Job->>Postgres: BEGIN

    Job->>Postgres: Update Resource → UPLOADED

    Job->>Postgres: INSERT Job(status=QUEUED)

    Job->>Outbox: INSERT JobCreated Event

    Job->>Postgres: COMMIT

    Job-->>Gateway: Job Created

    Gateway-->>User: 201 Created

    %% CDC
    Outbox->>Debezium: WAL Change

    Debezium->>Kafka: Publish JobCreated

    Kafka->>Worker: Consume JobCreated
```

## Sequence 2 — Worker Execution

```mermaid
sequenceDiagram
    autonumber

    participant Kafka
    participant Worker
    participant Redis
    participant MinIO
    participant Postgres
    participant Outbox
    participant Debezium
    participant Gateway
    participant User

    Kafka->>Worker: JobCreated

    Worker->>Postgres: BEGIN

    Worker->>Postgres: Job → RUNNING

    Worker->>Postgres: Resource → PROCESSING

    Worker->>Postgres: COMMIT

    %% Download
    Worker->>MinIO: Download Resource

    %% Processing
    Note over Worker: FFprobe<br/>FFmpeg<br/>Thumbnail

    loop Progress Updates
        Worker->>Redis: progress=10%
        Worker->>Redis: progress=35%
        Worker->>Redis: progress=70%
        Worker->>Redis: progress=100%
    end

    %% Upload Outputs
    Worker->>MinIO: Upload Processed Files

    %% Completion
    Worker->>Postgres: BEGIN

    Worker->>Postgres: Job → COMPLETED

    Worker->>Postgres: Resource → COMPLETED

    Worker->>Postgres: Save Output Metadata

    Worker->>Outbox: INSERT JobCompleted Event

    Worker->>Postgres: COMMIT

    Outbox->>Debezium: WAL Change

    Debezium->>Kafka: JobCompleted

    Kafka->>Gateway: Completion Event

    Gateway-->>User: WebSocket Notification
```

## Sequence 3 — Internal Worker Flow

```mermaid
sequenceDiagram
    autonumber

    participant Consumer
    participant Dispatcher
    participant VideoHandler
    participant MinIO
    participant FFmpeg
    participant Repository

    Consumer->>Dispatcher: Execute(Job)

    Dispatcher->>VideoHandler: Execute()

    VideoHandler->>Repository: Update RUNNING

    VideoHandler->>MinIO: Download

    MinIO-->>VideoHandler: Video

    VideoHandler->>FFmpeg: Process

    FFmpeg-->>VideoHandler: Outputs

    VideoHandler->>MinIO: Upload Outputs

    VideoHandler->>Repository: Update COMPLETED

    VideoHandler-->>Dispatcher: Success

    Dispatcher-->>Consumer: ACK
```
